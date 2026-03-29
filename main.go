package main

import (
	"github.com/spf13/cobra"
	"github.com/yuanjiecloud/fire/log"
)

var rootCommand = &cobra.Command{
	Use:   "fire",
	Short: "fire — a YAML-driven task pipeline runner",
	Long: `fire executes task pipelines defined in fire.yaml.

It supports local (bash/sh), remote (ssh), containerized (docker),
and batch execution modes, with dependency management backed by git.

Common usage:
  fire run              run all tasks in fire.yaml
  fire run <task>       run a specific task
  fire install          clone/checkout declared dependencies
  fire update           git-pull all cached dependencies
  fire list             list available tasks
  fire clean            remove cached dependency directories
  fire version          print the build version`,
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}
