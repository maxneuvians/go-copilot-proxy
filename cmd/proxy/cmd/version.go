package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Go Copilot Proxy",
	Long:  `All software has versions. This is Go Copilot Proxy's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Go Copilot Proxy v0.1 -- HEAD")
	},
}
