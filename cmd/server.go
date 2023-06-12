package cmd

import (
	"cs598fts/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server [addr]",
	Short: "Shared register server",
	Long:  `Shared register replica server.`,
	Run: func(cmd *cobra.Command, args []string) {
		s, err := server.NewServer(args[0])
		if err != nil {
			logrus.Fatal(err)
		}

		if err := s.Serve(); err != nil {
			logrus.Fatal(err)
		}
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(serverCmd)

	serverCmd.Args = cobra.ExactArgs(1)
}
