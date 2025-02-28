package cmd

import (
	"bufio"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maxneuvians/go-copilot-proxy/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var session_token string

var Model = "claude-3.7-sonnet"
var Completion_temperature = 0.3
var Completion_top_p = 0.9
var Completion_n = int64(1)
var Completion_stream = true

type Payload struct {
	Completion_N *int64        `json:"n,omitempty"`
	Messages     []pkg.Message `json:"messages"`
	Model        *string       `json:"model,omitempty"`
	Temperature  *float64      `json:"temperature,omitempty"`
	TopP         *float64      `json:"top_p,omitempty"`
}

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the proxy server",
	Long:  `Start the proxy server to enable GitHub Copilot proxy.`,
	Run: func(cmd *cobra.Command, args []string) {

		app := fiber.New()

		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

		// Validate that the TOKEN_FILE exists
		if _, err := os.Stat(TOKEN_FILE); os.IsNotExist(err) {
			log.Error().Msgf("The file %s does not exist, please run login first", TOKEN_FILE)
		}

		// Get the authentication token
		file, err := os.Open(TOKEN_FILE)

		if err != nil {
			log.Error().Msgf("Error opening token file: %s", err)
			return
		}

		// If the file exists, read the first line
		r := bufio.NewReader(file)
		buffer, _, err := r.ReadLine()

		if err != nil {
			log.Error().Msgf("Error reading token from file: %s", err)
			return
		}

		token := string(buffer)

		// Get a session token from the token
		sessionResponse, err := pkg.GetSessionToken(token)

		if err != nil {
			log.Error().Msgf("Error getting session token: %s", err)
			return
		}

		session_token = sessionResponse.Token

		// Start a ticker to refresh the session token every 25 minutes
		go func() {
			for {
				log.Info().Msg("Refreshing session token")
				// Sleep for 25 minutes
				time.Sleep(25 * time.Minute)

				// Get a new session token
				sessionResponse, err := pkg.GetSessionToken(token)

				if err != nil {
					log.Error().Msgf("Error getting session token: %s", err)
					return
				}
				session_token = sessionResponse.Token
			}
		}() // Start the ticker

		app.Post("/chat", func(c *fiber.Ctx) error {
			var payload Payload

			if err := c.BodyParser(&payload); err != nil {
				return err
			}

			n := Completion_n
			if payload.Completion_N != nil {
				n = *payload.Completion_N
			}

			model := Model
			if payload.Model != nil {
				model = *payload.Model
			}

			temperature := Completion_temperature
			if payload.Temperature != nil {
				temperature = *payload.Temperature
			}

			topP := Completion_top_p
			if payload.TopP != nil {
				topP = *payload.TopP
			}

			resp := ""

			err := pkg.Chat(session_token, payload.Messages, model, temperature, topP, n, false, func(completionResponse pkg.CompletionResponse) error {
				resp = completionResponse.Choices[0].Message.Content
				return nil
			})

			if err != nil {
				log.Error().Msgf("Error sending message: %s", err)
			}

			c.Set("Content-Type", "application/json")
			return c.JSON(resp)
		})

		app.Listen(":3000")

	},
}
