package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/maxneuvians/go-copilot-proxy/pkg"
)

// Helper function to create a test Fiber app with the chat endpoint
func createTestApp() *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Accept,Authorization,Content-Type,Content-Length,Accept-Encoding",
		AllowCredentials: true,
	}))

	// Mock the chat endpoint with the same logic as in start.go
	app.Post("/chat", func(c *fiber.Ctx) error {
		var payload Payload

		if err := c.BodyParser(&payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request payload",
			})
		}

		// Determine streaming mode
		stream := false
		if payload.Stream != nil {
			stream = *payload.Stream
		}

		// Use defaults if not provided
		model := Model
		if payload.Model != nil {
			model = *payload.Model
		}

		if stream {
			// Set SSE headers for streaming
			c.Set("Content-Type", "text/event-stream")
			c.Set("Cache-Control", "no-cache")
			c.Set("Connection", "keep-alive")
			c.Set("Access-Control-Allow-Origin", "*")

			// Mock streaming response
			completionID := "chatcmpl-test123"
			created := time.Now().Unix()

			// Simulate streaming chunks
			chunks := []string{"Hello", "! How", " can I", " help you", " today", "?"}

			for i, chunk := range chunks {
				streamChunk := pkg.CompletionResponse{
					ID:      completionID,
					Object:  "chat.completion.chunk",
					Created: created,
					Model:   model,
					Choices: []pkg.Choice{
						{
							Index: 0,
							Delta: &pkg.Message{
								Role:    "assistant",
								Content: chunk,
							},
							FinishReason: "",
						},
					},
				}

				// For the last chunk, set finish_reason
				if i == len(chunks)-1 {
					streamChunk.Choices[0].FinishReason = "stop"
					streamChunk.Choices[0].Delta.Content = ""
				}

				chunkBytes, _ := json.Marshal(streamChunk)
				fmt.Fprintf(c.Response().BodyWriter(), "data: %s\n\n", string(chunkBytes))
			}

			// Send final [DONE] message
			fmt.Fprintf(c.Response().BodyWriter(), "data: [DONE]\n\n")
			return nil
		} else {
			// Non-streaming response (existing logic)
			resp := "Hello! How can I help you today?"

			// Create usage estimation
			promptTokens := int64(len(fmt.Sprintf("%v", payload.Messages)) / 4)
			completionTokens := int64(len(resp) / 4)
			usage := pkg.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      promptTokens + completionTokens,
			}

			// Create OpenAI-compatible response
			openAIResponse := pkg.CompletionResponse{
				ID:      "chatcmpl-test123",
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
						FinishReason: "stop",
					},
				},
				Usage: usage,
			}

			c.Set("Content-Type", "application/json")
			return c.JSON(openAIResponse)
		}
	})

	return app
}

func TestChatEndpointValidRequest(t *testing.T) {
	app := createTestApp()

	// Create test request
	payload := Payload{
		Messages: []pkg.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain application/json, got %s", contentType)
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var completionResp pkg.CompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate OpenAI-compatible response structure
	if completionResp.ID == "" {
		t.Error("Response ID should not be empty")
	}
	if !strings.HasPrefix(completionResp.ID, "chatcmpl-") {
		t.Errorf("Expected ID to start with 'chatcmpl-', got %s", completionResp.ID)
	}
	if completionResp.Object != "chat.completion" {
		t.Errorf("Expected object: chat.completion, got %s", completionResp.Object)
	}
	if completionResp.Created == 0 {
		t.Error("Created timestamp should not be zero")
	}
	if completionResp.Model != Model {
		t.Errorf("Expected model: %s, got %s", Model, completionResp.Model)
	}
	if len(completionResp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(completionResp.Choices))
	}
	if completionResp.Choices[0].Index != 0 {
		t.Errorf("Expected choice index: 0, got %d", completionResp.Choices[0].Index)
	}
	if completionResp.Choices[0].Message != nil && completionResp.Choices[0].Message.Role != "assistant" {
		t.Errorf("Expected message role: assistant, got %s", completionResp.Choices[0].Message.Role)
	}
	if completionResp.Choices[0].Message != nil && completionResp.Choices[0].Message.Content == "" {
		t.Error("Message content should not be empty")
	}
	if completionResp.Choices[0].FinishReason != "stop" {
		t.Errorf("Expected finish reason: stop, got %s", completionResp.Choices[0].FinishReason)
	}
	if completionResp.Usage.TotalTokens == 0 {
		t.Error("Total tokens should not be zero")
	}
}

func TestChatEndpointWithCustomParameters(t *testing.T) {
	app := createTestApp()

	// Create test request with custom parameters
	customModel := "gpt-4"
	customTemp := 0.8
	customTopP := 0.95
	customN := int64(2)

	payload := Payload{
		Messages: []pkg.Message{
			{Role: "user", Content: "Tell me a joke"},
		},
		Model:        &customModel,
		Temperature:  &customTemp,
		TopP:         &customTopP,
		Completion_N: &customN,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var completionResp pkg.CompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate that custom model was used
	if completionResp.Model != customModel {
		t.Errorf("Expected model: %s, got %s", customModel, completionResp.Model)
	}
}

func TestChatEndpointInvalidJSON(t *testing.T) {
	app := createTestApp()

	// Create invalid JSON request
	invalidJSON := `{"messages": [{"role": "user", "content": "Hello"}` // Missing closing brackets

	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Parse error response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	// Check error message
	if errorResp["error"] == nil {
		t.Error("Expected error field in response")
	}
	errorMsg, ok := errorResp["error"].(string)
	if !ok {
		t.Error("Error field should be a string")
	}
	if errorMsg != "Invalid request payload" {
		t.Errorf("Expected error: 'Invalid request payload', got %s", errorMsg)
	}
}

func TestChatEndpointEmptyRequest(t *testing.T) {
	app := createTestApp()

	// Create empty request
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestChatEndpointWithEmptyMessages(t *testing.T) {
	app := createTestApp()

	// Create request with empty messages array
	payload := Payload{
		Messages: []pkg.Message{},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Should still return 200 since the mock handler doesn't validate message content
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

func TestChatEndpointMultipleMessages(t *testing.T) {
	app := createTestApp()

	// Create request with multiple messages (conversation history)
	payload := Payload{
		Messages: []pkg.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello!"},
			{Role: "assistant", Content: "Hi there! How can I help you?"},
			{Role: "user", Content: "What's the weather like?"},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var completionResp pkg.CompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate response
	if len(completionResp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(completionResp.Choices))
	}
	if completionResp.Choices[0].Message != nil && completionResp.Choices[0].Message.Role != "assistant" {
		t.Errorf("Expected assistant role, got %s", completionResp.Choices[0].Message.Role)
	}

	// Token usage should be estimated based on input length
	if completionResp.Usage.PromptTokens == 0 {
		t.Error("Prompt tokens should be estimated based on input length")
	}
	if completionResp.Usage.TotalTokens != completionResp.Usage.PromptTokens+completionResp.Usage.CompletionTokens {
		t.Error("Total tokens should equal prompt tokens + completion tokens")
	}
}

func TestChatEndpointCORSHeaders(t *testing.T) {
	app := createTestApp()

	// Create OPTIONS request to test CORS
	req := httptest.NewRequest(http.MethodOptions, "/chat", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS headers
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:5173" {
		t.Errorf("Expected Access-Control-Allow-Origin: http://localhost:5173, got %s", allowOrigin)
	}

	allowMethods := resp.Header.Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "POST") {
		t.Errorf("Expected Access-Control-Allow-Methods to contain POST, got %s", allowMethods)
	}
}

// Test payload struct validation
func TestPayloadSerialization(t *testing.T) {
	customModel := "test-model"
	customTemp := 0.9
	customTopP := 0.8
	customN := int64(3)

	payload := Payload{
		Completion_N: &customN,
		Messages: []pkg.Message{
			{Role: "user", Content: "Test message"},
		},
		Model:       &customModel,
		Temperature: &customTemp,
		TopP:        &customTopP,
	}

	// Test JSON marshaling
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal Payload: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Payload
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Payload: %v", err)
	}

	// Verify fields
	if *unmarshaled.Model != *payload.Model {
		t.Errorf("Model mismatch: expected %s, got %s", *payload.Model, *unmarshaled.Model)
	}
	if *unmarshaled.Temperature != *payload.Temperature {
		t.Errorf("Temperature mismatch: expected %f, got %f", *payload.Temperature, *unmarshaled.Temperature)
	}
	if *unmarshaled.TopP != *payload.TopP {
		t.Errorf("TopP mismatch: expected %f, got %f", *payload.TopP, *unmarshaled.TopP)
	}
	if *unmarshaled.Completion_N != *payload.Completion_N {
		t.Errorf("Completion_N mismatch: expected %d, got %d", *payload.Completion_N, *unmarshaled.Completion_N)
	}
	if len(unmarshaled.Messages) != len(payload.Messages) {
		t.Errorf("Messages length mismatch: expected %d, got %d", len(payload.Messages), len(unmarshaled.Messages))
	}
}

// Benchmark tests for the endpoint
func BenchmarkChatEndpoint(b *testing.B) {
	app := createTestApp()

	payload := Payload{
		Messages: []pkg.Message{
			{Role: "user", Content: "Hello, this is a benchmark test"},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		b.Fatalf("Failed to marshal payload: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			b.Fatalf("Failed to execute request: %v", err)
		}
		resp.Body.Close()
	}
}

// Test token estimation logic
func TestTokenEstimation(t *testing.T) {
	// Test cases for token estimation
	testCases := []struct {
		name          string
		messages      []pkg.Message
		expectedRatio float64 // Expected ratio of characters to tokens (roughly 4:1)
	}{
		{
			name: "Single short message",
			messages: []pkg.Message{
				{Role: "user", Content: "Hi"},
			},
			expectedRatio: 4.0,
		},
		{
			name: "Long message",
			messages: []pkg.Message{
				{Role: "user", Content: "This is a much longer message that should result in more tokens being estimated. The token estimation should be roughly one token per four characters."},
			},
			expectedRatio: 4.0,
		},
		{
			name: "Multiple messages",
			messages: []pkg.Message{
				{Role: "system", Content: "You are a helpful assistant."},
				{Role: "user", Content: "Hello there!"},
				{Role: "assistant", Content: "Hi! How can I help you today?"},
			},
			expectedRatio: 4.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate estimated tokens like in the actual implementation
			messagesStr := fmt.Sprintf("%v", tc.messages)
			estimatedTokens := int64(len(messagesStr) / 4)
			actualCharacters := len(messagesStr)

			if estimatedTokens == 0 && actualCharacters > 0 {
				t.Error("Token estimation should not be zero for non-empty messages")
			}

			if actualCharacters > 0 {
				ratio := float64(actualCharacters) / float64(estimatedTokens)
				// Be more lenient with ratio for short messages
				if ratio < 2.0 || ratio > 8.0 {
					t.Errorf("Token estimation ratio seems off. Expected roughly 2-8:1, got %f:1", ratio)
				}
			}
		})
	}
}

func TestChatEndpointStreamingRequest(t *testing.T) {
	app := createTestApp()

	// Create test request with streaming enabled
	streamValue := true
	payload := Payload{
		Messages: []pkg.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
		Stream: &streamValue,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	// Check SSE headers
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type: text/event-stream, got %s", contentType)
	}

	cacheControl := resp.Header.Get("Cache-Control")
	if cacheControl != "no-cache" {
		t.Errorf("Expected Cache-Control: no-cache, got %s", cacheControl)
	}

	connection := resp.Header.Get("Connection")
	if connection != "keep-alive" {
		t.Errorf("Expected Connection: keep-alive, got %s", connection)
	}

	// Read and validate SSE response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Check for SSE format
	if !strings.Contains(bodyStr, "data: ") {
		t.Error("Response should contain SSE formatted data")
	}

	// Check for [DONE] marker
	if !strings.Contains(bodyStr, "data: [DONE]") {
		t.Error("Response should end with [DONE] marker")
	}

	// Parse individual chunks
	lines := strings.Split(bodyStr, "\n")
	var chunks []pkg.CompletionResponse
	for _, line := range lines {
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk pkg.CompletionResponse
			if err := json.Unmarshal([]byte(data), &chunk); err == nil {
				chunks = append(chunks, chunk)
			}
		}
	}

	// Validate chunks
	if len(chunks) == 0 {
		t.Error("Should have received at least one chunk")
	}

	for i, chunk := range chunks {
		// Check basic structure
		if !strings.HasPrefix(chunk.ID, "chatcmpl-") {
			t.Errorf("Chunk %d: Expected ID to start with 'chatcmpl-', got %s", i, chunk.ID)
		}
		if chunk.Object != "chat.completion.chunk" {
			t.Errorf("Chunk %d: Expected object 'chat.completion.chunk', got %s", i, chunk.Object)
		}
		if len(chunk.Choices) != 1 {
			t.Errorf("Chunk %d: Expected 1 choice, got %d", i, len(chunk.Choices))
		}

		choice := chunk.Choices[0]
		if choice.Delta == nil {
			t.Errorf("Chunk %d: Choice should have delta field for streaming", i)
		}
		if choice.Message != nil {
			t.Errorf("Chunk %d: Choice should not have message field for streaming", i)
		}
	}

	// Check that last chunk has finish_reason
	if len(chunks) > 0 {
		lastChunk := chunks[len(chunks)-1]
		if lastChunk.Choices[0].FinishReason != "stop" {
			t.Errorf("Last chunk should have finish_reason 'stop', got '%s'", lastChunk.Choices[0].FinishReason)
		}
	}
}

func TestChatEndpointNonStreamingDefault(t *testing.T) {
	app := createTestApp()

	// Create test request without stream parameter (should default to non-streaming)
	payload := Payload{
		Messages: []pkg.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
		// Stream is nil, should default to false
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Should be non-streaming response
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain application/json, got %s", contentType)
	}

	// Should not have SSE headers
	if resp.Header.Get("Cache-Control") == "no-cache" {
		t.Error("Non-streaming response should not have SSE cache headers")
	}
}

func TestChatEndpointStreamingFalse(t *testing.T) {
	app := createTestApp()

	// Create test request with streaming explicitly disabled
	streamValue := false
	payload := Payload{
		Messages: []pkg.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
		Stream: &streamValue,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	// Should be non-streaming response
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type to contain application/json, got %s", contentType)
	}

	// Parse response as regular JSON (not SSE)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var completionResp pkg.CompletionResponse
	err = json.Unmarshal(body, &completionResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should have message field, not delta
	if len(completionResp.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(completionResp.Choices))
	}

	choice := completionResp.Choices[0]
	if choice.Message == nil {
		t.Error("Non-streaming response should have message field")
	}
	if choice.Delta != nil {
		t.Error("Non-streaming response should not have delta field")
	}
}
