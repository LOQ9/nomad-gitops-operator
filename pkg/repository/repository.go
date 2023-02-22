package repository

import (
	"context"
	"fmt"
	"net/url"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// CLone will clone a git repo
func CLone(ctx context.Context, repositoryURL *url.URL, branch string, username string, password string) (*git.Worktree, error) {
	storer := memory.NewStorage()
	fs := memfs.New()
	ranchRef := plumbing.NewBranchReferenceName(branch)
	gitClone := &git.CloneOptions{
		URL:           repositoryURL.String(),
		ReferenceName: ranchRef,
	}

	if password != "" {
		gitClone.Auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	repo, err := git.CloneContext(ctx, storer, fs, gitClone)
	if err != nil {
		return nil, fmt.Errorf("unable to clone repository %s for templating: %w", repositoryURL.String(), err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("unable to get worktree: %w", err)
	}

	return w, nil
}
