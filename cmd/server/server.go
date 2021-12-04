package server

import (
	"github.com/spf13/cobra"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Executes the server commands",
}

func init() {
	ServerCmd.AddCommand(RunCmd)
}
