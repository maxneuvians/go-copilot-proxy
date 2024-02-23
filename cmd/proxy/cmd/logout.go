package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(logoutCmd)
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of GitHub Copilot",
	Long:  `Logs you out of GitHub Copilot by deleting the token file.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Info().Msg("Logging out of GitHub Copilot")

		if _, err := os.Stat(TOKEN_FILE); err != nil {
			log.Error().Msg("You are not logged in.")
			return
		}

		err := os.Remove(TOKEN_FILE)
		if err != nil {
			log.Error().Msg("Failed to log out.")
			return
		}
	},
}
