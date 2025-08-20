package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/maxneuvians/go-copilot-proxy/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var session_token string

var (
	Model                  = "claude-3.7-sonnet"
	Completion_temperature = 0.3
	Completion_top_p       = 0.9
	Completion_n           = int64(1)
	Completion_stream      = true
)

type Payload struct {
	Completion_N *int64        `json:"n,omitempty"`
	Messages     []pkg.Message `json:"messages"`
	Model        *string       `json:"model,omitempty"`
	Temperature  *float64      `json:"temperature,omitempty"`
	TopP         *float64      `json:"top_p,omitempty"`
	Stream       *bool         `json:"stream,omitempty"`
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

		// Define the chat handler function that can be reused
		chatHandler := func(c *fiber.Ctx) error {
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

			// Determine streaming mode
			stream := false
			if payload.Stream != nil {
				stream = *payload.Stream
			}

			// Log parsed payload details
			modelStr := Model
			if payload.Model != nil {
				modelStr = *payload.Model
			}
			log.Debug().
				Int("message_count", len(payload.Messages)).
				Str("model", modelStr).
				Bool("stream", stream).
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

			startTime := time.Now()

			if stream {
				// Set SSE headers for streaming
				c.Set("Content-Type", "text/event-stream")
				c.Set("Cache-Control", "no-cache")
				c.Set("Connection", "keep-alive")
				c.Set("Access-Control-Allow-Origin", "*")

				// Generate unique ID for this completion
				completionID := "chatcmpl-" + uuid.New().String()
				created := time.Now().Unix()

				// Handle streaming response
				err := pkg.Chat(session_token, payload.Messages, model, temperature, topP, n, true, func(completionResponse pkg.CompletionResponse) error {
					if len(completionResponse.Choices) == 0 {
						return nil
					}

					choice := completionResponse.Choices[0]

					// Handle the case where we get a chunk with both content and finish_reason
					// This ensures we follow OpenAI's specification correctly
					if choice.FinishReason != "" && choice.Delta != nil && choice.Delta.Content != "" {
						// Send the content chunk first (without finish_reason)
						contentChunk := pkg.CompletionResponse{
							ID:      completionID,
							Object:  "chat.completion.chunk",
							Created: created,
							Model:   model,
							Choices: []pkg.Choice{
								{
									Index: choice.Index,
									Delta: &pkg.Message{
										Role:    choice.Delta.Role,
										Content: choice.Delta.Content,
									},
									FinishReason: "", // No finish reason for content chunk
								},
							},
						}

						chunkBytes, err := json.Marshal(contentChunk)
						if err != nil {
							log.Error().Err(err).Msg("Failed to marshal content chunk")
							return err
						}

						_, writeErr := fmt.Fprintf(c.Response().BodyWriter(), "data: %s\n\n", string(chunkBytes))
						if writeErr != nil {
							log.Error().Err(writeErr).Msg("Failed to write content chunk")
							return writeErr
						}

						// Flush the response
						if f, ok := c.Response().BodyWriter().(interface{ Flush() }); ok {
							f.Flush()
						}

						// Send the finish reason chunk separately (with empty content)
						finishChunk := pkg.CompletionResponse{
							ID:      completionID,
							Object:  "chat.completion.chunk",
							Created: created,
							Model:   model,
							Choices: []pkg.Choice{
								{
									Index: choice.Index,
									Delta: &pkg.Message{
										Role:    "",
										Content: "",
									},
									FinishReason: choice.FinishReason,
								},
							},
						}

						finishBytes, err := json.Marshal(finishChunk)
						if err != nil {
							log.Error().Err(err).Msg("Failed to marshal finish chunk")
							return err
						}

						_, writeErr = fmt.Fprintf(c.Response().BodyWriter(), "data: %s\n\n", string(finishBytes))
						if writeErr != nil {
							log.Error().Err(writeErr).Msg("Failed to write finish chunk")
							return writeErr
						}

						// Flush the response
						if f, ok := c.Response().BodyWriter().(interface{ Flush() }); ok {
							f.Flush()
						}

						return nil
					}

					// Handle normal chunks (either content-only or finish-only)
					streamChunk := pkg.CompletionResponse{
						ID:      completionID,
						Object:  "chat.completion.chunk",
						Created: created,
						Model:   model,
						Choices: []pkg.Choice{
							{
								Index:        choice.Index,
								Delta:        choice.Delta,
								FinishReason: choice.FinishReason,
							},
						},
					}

					// Marshal and send chunk
					chunkBytes, err := json.Marshal(streamChunk)
					if err != nil {
						log.Error().Err(err).Msg("Failed to marshal stream chunk")
						return err
					}

					// Write SSE formatted data
					_, writeErr := fmt.Fprintf(c.Response().BodyWriter(), "data: %s\n\n", string(chunkBytes))
					if writeErr != nil {
						log.Error().Err(writeErr).Msg("Failed to write stream chunk")
						return writeErr
					}

					// Flush the response
					if f, ok := c.Response().BodyWriter().(interface{ Flush() }); ok {
						f.Flush()
					}

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
						Msg("Failed to get streaming chat completion")
					// Send error in SSE format
					errorData := map[string]interface{}{
						"error": map[string]interface{}{
							"message": fmt.Sprintf("Failed to process chat request: %v", err),
							"type":    "server_error",
						},
					}
					errorBytes, _ := json.Marshal(errorData)
					fmt.Fprintf(c.Response().BodyWriter(), "data: %s\n\n", string(errorBytes))
					return nil
				}

				// Send final [DONE] message
				fmt.Fprintf(c.Response().BodyWriter(), "data: [DONE]\n\n")

				// Log streaming completion
				log.Debug().
					Str("model", model).
					Float64("duration_ms", float64(time.Since(startTime).Milliseconds())).
					Str("completion_id", completionID).
					Msg("Streaming chat request completed successfully")

				return nil
			} else {
				// Non-streaming response (existing logic)
				resp := ""
				var completionResp pkg.CompletionResponse

				err := pkg.Chat(session_token, payload.Messages, model, temperature, topP, n, false, func(completionResponse pkg.CompletionResponse) error {
					// Add validation and logging
					if len(completionResponse.Choices) == 0 {
						log.Error().
							Interface("response", completionResponse).
							Msg("Empty choices array in completion response")
						return fmt.Errorf("no choices in completion response")
					}

					choice := completionResponse.Choices[0]
					if choice.Message != nil {
						resp = choice.Message.Content
					} else if choice.Delta != nil {
						resp = choice.Delta.Content
					}
					completionResp = completionResponse
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

				// Create OpenAI-compatible response
				usage := completionResp.Usage
				// If usage is not available from the original response, create default values
				if usage.TotalTokens == 0 && usage.PromptTokens == 0 && usage.CompletionTokens == 0 {
					// Estimate token counts (rough approximation)
					promptTokens := int64(len(fmt.Sprintf("%v", payload.Messages)) / 4)
					completionTokens := int64(len(resp) / 4)
					usage = pkg.Usage{
						PromptTokens:     promptTokens,
						CompletionTokens: completionTokens,
						TotalTokens:      promptTokens + completionTokens,
					}
				}

				openAIResponse := pkg.CompletionResponse{
					ID:      "chatcmpl-" + uuid.New().String(),
					Object:  "chat.completion",
					Created: time.Now().Unix(),
					Model:   model,
					Choices: []pkg.Choice{
						{
							Index: 0,
							Message: &pkg.Message{
								Role:    "assistant",
								Content: resp,
							},
							FinishReason: pkg.FinishReasonStop,
						},
					},
					Usage: usage,
				}

				// Log successful response
				log.Debug().
					Str("model", model).
					Int("response_length", len(resp)).
					Float64("duration_ms", float64(time.Since(startTime).Milliseconds())).
					Interface("response", openAIResponse).
					Msg("Chat request completed successfully")

				c.Set("Content-Type", "application/json")
				return c.JSON(openAIResponse)
			}
		}

		// Register the chat handler for both endpoints
		app.Post("/chat", chatHandler)
		app.Post("/v1/chat/completions", chatHandler)

		app.Listen(":3000")
	},
}
