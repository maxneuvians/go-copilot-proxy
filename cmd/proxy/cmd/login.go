package cmd

import (
	"os"
	"time"

	"github.com/maxneuvians/go-copilot-proxy/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to GitHub Copilot",
	Long:  `Login to GitHub Copilot using your GitHub account.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Info().Msg("Authorizing user with Copilot")

		if _, err := os.Stat(".github_copilot_token"); err == nil {
			log.Error().Msg("You are already logged in.")
			return
		}

		loginResponse, err := pkg.Login()
		if err != nil {
			log.Error().Msgf("Error logging in: %s", err)
			return
		}

		log.Info().Msgf("Please visit %s to authenticate and enter the code: %s", loginResponse.VerificationURI, loginResponse.UserCode)

		// Sleep for the interval time
		time.Sleep(time.Duration(loginResponse.Interval+1) * time.Second)

		var authResponse pkg.AuthenticationResponse

		for {
			authResponse, err = pkg.Authenticate(loginResponse)
			if err != nil {
				log.Error().Msgf("Error authenticating: %s", err)
				return
			}

			if authResponse.AccessToken != "" {
				log.Info().Msg("Authenticated successfully!")
				break
			}

			// If the interval is 0, set it to 5
			if authResponse.Interval == 0 {
				authResponse.Interval = 5
			}

			// Sleep for the interval time
			time.Sleep(time.Duration(authResponse.Interval+1) * time.Second)
		}

		// Write the token to a file
		file, err := os.Create(TOKEN_FILE)
		if err != nil {
			log.Error().Msgf("Error creating token file: %s", err)
			return
		}

		_, err = file.WriteString(authResponse.AccessToken)
		if err != nil {
			log.Error().Msgf("Error writing token to file: %s", err)
			return
		}
	},
}
