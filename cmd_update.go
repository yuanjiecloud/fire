package main

import (
	"github.com/spf13/cobra"
	"github.com/yuanjiecloud/fire/log"
)

type updateCommand struct {
	*contextCommand
}

func (t *updateCommand) Prepare(cmd *cobra.Command) {
}

func (t *updateCommand) InitFlag(cmd *cobra.Command) {
	t.contextCommand.InitFlag(cmd)
}

func (t *updateCommand) BeforeRun(cmd *cobra.Command) {
	t.contextCommand.BeforeRun(cmd)
	err := t.pipeline.Preload()
	log.CheckAndFatal(err)
}

func (t *updateCommand) Run(cmd *cobra.Command, args []string) {
	err := t.pipeline.UpdateDependencies()
	log.CheckAndFatal(err)
}

func init() {
	rootCommand.AddCommand(NewCommandBuilder("update", &updateCommand{&contextCommand{}}).Build())
}
