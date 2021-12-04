package cmd

import (
	"github.com/herlon214/sonarqube-pr-issues/cmd/cli"
	"github.com/herlon214/sonarqube-pr-issues/cmd/server"
	"github.com/spf13/cobra"
	"os"

	"github.com/sirupsen/logrus"
)

var rootCmd = &cobra.Command{
	Use:   "sqpr",
	Short: "SQPR publishes Sonarqube the issues into your PRs",
}

// Flags
var mode string

func init() {
	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(cli.CliCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err)
		os.Exit(1)
	}
}
