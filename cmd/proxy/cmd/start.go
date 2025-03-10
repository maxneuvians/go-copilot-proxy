package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
		 // Add CORS middleware
		 app.Use(cors.New(cors.Config{
            AllowOrigins:     "http://localhost:5173",
            AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
            AllowHeaders:     "Accept,Authorization,Content-Type,Content-Length,Accept-Encoding",
            AllowCredentials: true,
        }))

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

			// Log incoming request
			log.Debug().
				Str("path", "/chat").
				Str("method", "POST").
				Str("remote_ip", c.IP()).
				Msg("Incoming chat request")

			if err := c.BodyParser(&payload); err != nil {
				log.Error().
					Err(err).
					Str("path", "/chat").
					Interface("payload", payload).
					Msg("Failed to parse request body")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid request payload",
				})
			}

			// Log parsed payload details
			log.Debug().
				Int("message_count", len(payload.Messages)).
				Str("model", *payload.Model).
				Interface("messages", payload.Messages).
				Msg("Processing chat request")

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
			startTime := time.Now()

			err := pkg.Chat(session_token, payload.Messages, model, temperature, topP, n, false, func(completionResponse pkg.CompletionResponse) error {
				// Add validation and logging
				if len(completionResponse.Choices) == 0 {
					log.Error().
						Interface("response", completionResponse).
						Msg("Empty choices array in completion response")
					return fmt.Errorf("no choices in completion response")
				}
				
				resp = completionResponse.Choices[0].Message.Content
				return nil
			})

			if err != nil {
				log.Error().
					Err(err).
					Str("model", model).
					Float64("temperature", temperature).
					Float64("top_p", topP).
					Int64("n", n).
					Interface("messages", payload.Messages).
					Msg("Failed to get chat completion")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to process chat request: %v", err),
				})
			}

			// Log successful response
			responseJSON := fiber.Map{"content": resp}
			log.Debug().
				Str("model", model).
				Int("response_length", len(resp)).
				Float64("duration_ms", float64(time.Since(startTime).Milliseconds())).
				Interface("response", responseJSON).  // Changed from RawJSON to Interface
				Msg("Chat request completed successfully")

			c.Set("Content-Type", "application/json")
			return c.JSON(responseJSON)
		})

		app.Listen(":3000")

	},
}
