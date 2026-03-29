package main

import (
	"github.com/spf13/cobra"
	"github.com/yuanjiecloud/fire/log"
)

type runCommand struct {
	*contextCommand
	taskName string
}

func (t *runCommand) Run(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		t.taskName = args[0]
	}
	var err error
	if t.taskName != "" {
		err = t.pipeline.RunTask(t.taskName, nil)
	} else {
		err = t.pipeline.RunAll(nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (t *runCommand) Prepare(cmd *cobra.Command) {
}

func (t *runCommand) InitFlag(cmd *cobra.Command) {
	t.contextCommand.InitFlag(cmd)
	cmd.PersistentFlags().StringVarP(&t.taskName, "task", "t", "", "task name")
}

func (t *runCommand) BeforeRun(cmd *cobra.Command) {
	t.contextCommand.BeforeRun(cmd)
	err := t.pipeline.Preload()
	log.CheckAndFatal(err)
}

func init() {
	cmd := NewCommandBuilder("run", &runCommand{contextCommand: &contextCommand{}}).Build()
	rootCommand.AddCommand(cmd)
}
