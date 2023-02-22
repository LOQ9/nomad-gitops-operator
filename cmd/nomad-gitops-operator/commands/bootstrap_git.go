package commands

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"nomad-gitops-operator/pkg/nomad"
	"nomad-gitops-operator/pkg/repository"
)

type gitFlags struct {
	url      string
	branch   string
	path     string
	username string
	password string
}

var gitArgs gitFlags

var bootstrapGitCmd = &cobra.Command{
	Use:   "git [git repo]",
	Short: "Bootstrap Nomad using a git repository",
	Long:  ``,
	// Args:  cobra.MinimumNArgs(1),
	RunE: ReconcileLoop,
}

func init() {
	bootstrapCmd.AddCommand(bootstrapGitCmd)
	bootstrapGitCmd.Flags().StringVar(&gitArgs.url, "url", "", "git repository URL")
	bootstrapGitCmd.Flags().StringVar(&gitArgs.branch, "branch", "main", "git branch")
	bootstrapGitCmd.Flags().StringVar(&gitArgs.path, "path", "/", "path relative to the repository root")
	bootstrapGitCmd.Flags().StringVar(&gitArgs.username, "username", "", "username from the git account")
	bootstrapGitCmd.Flags().StringVar(&gitArgs.password, "password", "", "password from the git account")
}

func ReconcileLoop(cmd *cobra.Command, args []string) error {
	// Create Nomad client
	client, err := nomad.NewClient()
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Could not create the nomad client")
	}

	// Reconcile
	for true {
		repositoryURL, err := url.Parse(gitArgs.url)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Minute*5))
		defer cancel()

		worktree, err := repository.CLone(ctx, repositoryURL, gitArgs.branch, gitArgs.username, gitArgs.password)

		if err != nil {
			return fmt.Errorf("unable to clone the repository, err=%s", err)
		}
		worktree.Pull(&git.PullOptions{RemoteName: "origin"})

		fs := worktree.Filesystem
		path := gitArgs.path
		files, err := fs.ReadDir(path)
		if err != nil {
			return fmt.Errorf("unable to read the folder, err=%s", err)
		}

		desiredStateJobs := make(map[string]interface{})

		// Parse and apply all jobs from within the git repo
		for _, file := range files {
			filePath := fs.Join(path, file.Name())
			f, err := fs.Open(filePath)
			if err != nil {
				return err
			}

			b, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			// Parse job
			job, err := client.ParseJob(string(b))
			if err != nil {
				// If a parse error occurs we skip the job an continue with the next job
				logrus.WithFields(logrus.Fields{"job": job.GetName(), "error": err}).Error("Unable to parse the job")
				continue
			}
			desiredStateJobs[job.GetName()] = job

			// Apply job
			logrus.WithFields(logrus.Fields{"job": job.GetName()}).Info("Applying job")
			_, err = client.ApplyJob(job)
			if err != nil {
				logrus.WithFields(logrus.Fields{"job": job.GetName(), "error": err}).Error("Unable to apply job")
				continue
			}
		}

		// List all jobs managed by Monoporator
		currentStateJobs, err := client.ListJobs()
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err}).Error("Unable to list jobs")
		}

		// Check if job has the required metadata
		// Check if job is one of the parsed jobs
		for _, job := range currentStateJobs {
			meta := job.GetMeta()

			if _, isManaged := meta["nomoporater"]; isManaged {
				// If the job is managed by Nomoporator and is part of the desired state
				if _, inDesiredState := desiredStateJobs[job.GetName()]; inDesiredState {

				} else {
					logrus.WithFields(logrus.Fields{"job": job.GetName()}).Info("Deleting job")
					err = client.DeleteJob(job)
					if err != nil {
						logrus.WithFields(logrus.Fields{"job": job.GetName(), "error": err}).Error("Unable to delete job")
					}
				}
			}
		}

		time.Sleep(30 * time.Second)
	}

	return nil
}
