package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-copilot-proxy",
	Short: "Go Copilot Proxy is a proxy server for GitHub Copilot",
	Long:  `Go Copilot Proxy is a proxy server for GitHub Copilot that allows you to use the GitHub Copilot API without needing to use Visual Studio Code.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Info().Msg("Welcome to Go Copilot Proxy! Use the --help flag to see available commands.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
