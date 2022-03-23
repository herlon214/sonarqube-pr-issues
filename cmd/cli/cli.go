package cli

import (
	"github.com/spf13/cobra"
)

var CliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Executes CLI commands",
}

var project string
var branch string
var publishReview bool
var markAsPublished bool
var requestChanges bool

func init() {
	CliCmd.PersistentFlags().StringVar(&project, "project", "my-project", "Sonarqube project name")
	CliCmd.PersistentFlags().StringVar(&branch, "branch", "my-branch", "SCM branch name")
	CliCmd.PersistentFlags().BoolVar(&publishReview, "publish", false, "Publish review in the SCM")
	CliCmd.PersistentFlags().BoolVar(&markAsPublished, "mark", false, "Mark the issue as published to avoid sending it again")
	CliCmd.PersistentFlags().BoolVar(&requestChanges, "request-changes", true, "When issue is found, mark PR as changes requested")

	CliCmd.AddCommand(RunCmd)
}
