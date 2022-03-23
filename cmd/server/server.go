package server

import (
	"github.com/spf13/cobra"
)

var serverPort int
var workers int
var requestChanges bool

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Executes the server commands",
}

func init() {
	ServerCmd.PersistentFlags().IntVarP(&serverPort, "port", "p", 8080, "Server port")
	ServerCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 30, "Workers count")
	ServerCmd.PersistentFlags().BoolVar(&requestChanges, "request-changes", true, "When issue is found, mark PR as changes requested")
	ServerCmd.AddCommand(RunCmd)
}
