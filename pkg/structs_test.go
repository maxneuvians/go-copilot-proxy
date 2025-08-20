package pkg

import (
	"encoding/json"
	"testing"
)

func TestAuthenticationRequestSerialization(t *testing.T) {
	req := AuthenticationRequest{
		ClientID:   "test-client-id",
		DeviceCode: "test-device-code",
		GrantType:  "test-grant-type",
	}

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal AuthenticationRequest: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled AuthenticationRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal AuthenticationRequest: %v", err)
	}

	// Verify fields
	if unmarshaled.ClientID != req.ClientID {
		t.Errorf("ClientID mismatch: expected %s, got %s", req.ClientID, unmarshaled.ClientID)
	}
	if unmarshaled.DeviceCode != req.DeviceCode {
		t.Errorf("DeviceCode mismatch: expected %s, got %s", req.DeviceCode, unmarshaled.DeviceCode)
	}
	if unmarshaled.GrantType != req.GrantType {
		t.Errorf("GrantType mismatch: expected %s, got %s", req.GrantType, unmarshaled.GrantType)
	}
}

func TestAuthenticationResponseSerialization(t *testing.T) {
	resp := AuthenticationResponse{
		AccessToken: "test-access-token",
		Interval:    30,
		TokenType:   "Bearer",
		Scope:       "read:user",
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal AuthenticationResponse: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled AuthenticationResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal AuthenticationResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.AccessToken != resp.AccessToken {
		t.Errorf("AccessToken mismatch: expected %s, got %s", resp.AccessToken, unmarshaled.AccessToken)
	}
	if unmarshaled.Interval != resp.Interval {
		t.Errorf("Interval mismatch: expected %d, got %d", resp.Interval, unmarshaled.Interval)
	}
	if unmarshaled.TokenType != resp.TokenType {
		t.Errorf("TokenType mismatch: expected %s, got %s", resp.TokenType, unmarshaled.TokenType)
	}
	if unmarshaled.Scope != resp.Scope {
		t.Errorf("Scope mismatch: expected %s, got %s", resp.Scope, unmarshaled.Scope)
	}
}

func TestCompletionResponseOpenAICompatibility(t *testing.T) {
	// Test OpenAI-compatible response structure
	resp := CompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1234567890,
		Model:   "claude-3.7-sonnet",
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Hello! How can I help you?",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     15,
			CompletionTokens: 8,
			TotalTokens:      23,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionResponse: %v", err)
	}

	// Verify the JSON structure matches OpenAI format
	var jsonObj map[string]interface{}
	err = json.Unmarshal(data, &jsonObj)
	if err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Check required OpenAI fields
	requiredFields := []string{"id", "object", "created", "model", "choices", "usage"}
	for _, field := range requiredFields {
		if _, exists := jsonObj[field]; !exists {
			t.Errorf("Missing required OpenAI field: %s", field)
		}
	}

	// Verify specific values
	if jsonObj["id"] != "chatcmpl-123" {
		t.Errorf("ID mismatch: expected chatcmpl-123, got %v", jsonObj["id"])
	}
	if jsonObj["object"] != "chat.completion" {
		t.Errorf("Object mismatch: expected chat.completion, got %v", jsonObj["object"])
	}
	if jsonObj["model"] != "claude-3.7-sonnet" {
		t.Errorf("Model mismatch: expected claude-3.7-sonnet, got %v", jsonObj["model"])
	}

	// Verify choices structure
	choices, ok := jsonObj["choices"].([]interface{})
	if !ok {
		t.Fatalf("Choices is not an array")
	}
	if len(choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(choices))
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Choice is not an object")
	}

	// Check choice fields - index might be omitted if 0 due to omitempty tag
	if index, exists := choice["index"]; exists && index != float64(0) {
		t.Errorf("Choice index mismatch: expected 0, got %v", choice["index"])
	}
	if choice["finish_reason"] != "stop" {
		t.Errorf("Finish reason mismatch: expected stop, got %v", choice["finish_reason"])
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		t.Fatalf("Message is not an object")
	}
	if message["role"] != "assistant" {
		t.Errorf("Message role mismatch: expected assistant, got %v", message["role"])
	}
	if message["content"] != "Hello! How can I help you?" {
		t.Errorf("Message content mismatch: expected 'Hello! How can I help you?', got %v", message["content"])
	}

	// Verify usage structure
	usage, ok := jsonObj["usage"].(map[string]interface{})
	if !ok {
		t.Fatalf("Usage is not an object")
	}
	if usage["prompt_tokens"] != float64(15) {
		t.Errorf("Prompt tokens mismatch: expected 15, got %v", usage["prompt_tokens"])
	}
	if usage["completion_tokens"] != float64(8) {
		t.Errorf("Completion tokens mismatch: expected 8, got %v", usage["completion_tokens"])
	}
	if usage["total_tokens"] != float64(23) {
		t.Errorf("Total tokens mismatch: expected 23, got %v", usage["total_tokens"])
	}
}

func TestMessageSerialization(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello, how are you?",
	}

	// Test JSON marshaling
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal Message: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Message: %v", err)
	}

	// Verify fields
	if unmarshaled.Role != msg.Role {
		t.Errorf("Role mismatch: expected %s, got %s", msg.Role, unmarshaled.Role)
	}
	if unmarshaled.Content != msg.Content {
		t.Errorf("Content mismatch: expected %s, got %s", msg.Content, unmarshaled.Content)
	}
}

func TestUsageSerialization(t *testing.T) {
	usage := Usage{
		CompletionTokens: 50,
		PromptTokens:     100,
		TotalTokens:      150,
	}

	// Test JSON marshaling
	data, err := json.Marshal(usage)
	if err != nil {
		t.Fatalf("Failed to marshal Usage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Usage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Usage: %v", err)
	}

	// Verify fields
	if unmarshaled.CompletionTokens != usage.CompletionTokens {
		t.Errorf("CompletionTokens mismatch: expected %d, got %d", usage.CompletionTokens, unmarshaled.CompletionTokens)
	}
	if unmarshaled.PromptTokens != usage.PromptTokens {
		t.Errorf("PromptTokens mismatch: expected %d, got %d", usage.PromptTokens, unmarshaled.PromptTokens)
	}
	if unmarshaled.TotalTokens != usage.TotalTokens {
		t.Errorf("TotalTokens mismatch: expected %d, got %d", usage.TotalTokens, unmarshaled.TotalTokens)
	}
}

func TestChoiceSerialization(t *testing.T) {
	choice := Choice{
		FinishReason: "stop",
		Index:        0,
		Message: Message{
			Role:    "assistant",
			Content: "Test response",
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(choice)
	if err != nil {
		t.Fatalf("Failed to marshal Choice: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled Choice
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Choice: %v", err)
	}

	// Verify fields
	if unmarshaled.FinishReason != choice.FinishReason {
		t.Errorf("FinishReason mismatch: expected %s, got %s", choice.FinishReason, unmarshaled.FinishReason)
	}
	if unmarshaled.Index != choice.Index {
		t.Errorf("Index mismatch: expected %d, got %d", choice.Index, unmarshaled.Index)
	}
	if unmarshaled.Message.Role != choice.Message.Role {
		t.Errorf("Message.Role mismatch: expected %s, got %s", choice.Message.Role, unmarshaled.Message.Role)
	}
	if unmarshaled.Message.Content != choice.Message.Content {
		t.Errorf("Message.Content mismatch: expected %s, got %s", choice.Message.Content, unmarshaled.Message.Content)
	}
}

func TestCompletionRequestSerialization(t *testing.T) {
	req := CompletionRequest{
		Model: "claude-3.7-sonnet",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
			{Role: "assistant", Content: "Hi there!"},
		},
		Stream:      false,
		Temperature: 0.7,
		TopP:        0.9,
		N:           1,
	}

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal CompletionRequest: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled CompletionRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal CompletionRequest: %v", err)
	}

	// Verify fields
	if unmarshaled.Model != req.Model {
		t.Errorf("Model mismatch: expected %s, got %s", req.Model, unmarshaled.Model)
	}
	if len(unmarshaled.Messages) != len(req.Messages) {
		t.Errorf("Messages length mismatch: expected %d, got %d", len(req.Messages), len(unmarshaled.Messages))
	}
	if unmarshaled.Stream != req.Stream {
		t.Errorf("Stream mismatch: expected %t, got %t", req.Stream, unmarshaled.Stream)
	}
	if unmarshaled.Temperature != req.Temperature {
		t.Errorf("Temperature mismatch: expected %f, got %f", req.Temperature, unmarshaled.Temperature)
	}
	if unmarshaled.TopP != req.TopP {
		t.Errorf("TopP mismatch: expected %f, got %f", req.TopP, unmarshaled.TopP)
	}
	if unmarshaled.N != req.N {
		t.Errorf("N mismatch: expected %d, got %d", req.N, unmarshaled.N)
	}
}

func TestLoginRequestSerialization(t *testing.T) {
	req := LoginRequest{
		ClientID: "test-client-id",
		Scopes:   "read:user",
	}

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal LoginRequest: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled LoginRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal LoginRequest: %v", err)
	}

	// Verify fields
	if unmarshaled.ClientID != req.ClientID {
		t.Errorf("ClientID mismatch: expected %s, got %s", req.ClientID, unmarshaled.ClientID)
	}
	if unmarshaled.Scopes != req.Scopes {
		t.Errorf("Scopes mismatch: expected %s, got %s", req.Scopes, unmarshaled.Scopes)
	}
}

func TestLoginResponseSerialization(t *testing.T) {
	resp := LoginResponse{
		DeviceCode:      "test-device-code",
		Interval:        5,
		UserCode:        "TEST-123",
		VerificationURI: "https://github.com/login/device",
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal LoginResponse: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled LoginResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal LoginResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.DeviceCode != resp.DeviceCode {
		t.Errorf("DeviceCode mismatch: expected %s, got %s", resp.DeviceCode, unmarshaled.DeviceCode)
	}
	if unmarshaled.Interval != resp.Interval {
		t.Errorf("Interval mismatch: expected %d, got %d", resp.Interval, unmarshaled.Interval)
	}
	if unmarshaled.UserCode != resp.UserCode {
		t.Errorf("UserCode mismatch: expected %s, got %s", resp.UserCode, unmarshaled.UserCode)
	}
	if unmarshaled.VerificationURI != resp.VerificationURI {
		t.Errorf("VerificationURI mismatch: expected %s, got %s", resp.VerificationURI, unmarshaled.VerificationURI)
	}
}

func TestSessionResponseSerialization(t *testing.T) {
	resp := SessionResponse{
		Token:     "test-session-token",
		ExpiresAt: 1234567890,
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal SessionResponse: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SessionResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SessionResponse: %v", err)
	}

	// Verify fields
	if unmarshaled.Token != resp.Token {
		t.Errorf("Token mismatch: expected %s, got %s", resp.Token, unmarshaled.Token)
	}
	if unmarshaled.ExpiresAt != resp.ExpiresAt {
		t.Errorf("ExpiresAt mismatch: expected %d, got %d", resp.ExpiresAt, unmarshaled.ExpiresAt)
	}
}
