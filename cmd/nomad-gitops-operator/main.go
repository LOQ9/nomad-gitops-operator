package main

import (
	"nomad-gitops-operator/cmd/nomad-gitops-operator/commands"
	"os"
)

func main() {
	err := commands.Execute()
	if err != nil {
		os.Exit(1)
	}
}
