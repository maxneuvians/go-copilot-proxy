package pkg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogin(t *testing.T) {
	// Mock server for GitHub login endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		expectedHeaders := map[string]string{
			"accept":                "application/json",
			"content-type":          "application/json",
			"editor-version":        "vscode/1.83.1",
			"editor-plugin-version": "copilot-chat/0.8.0",
			"user-agent":           "githubCopilot/1.155.0",
		}
		
		for header, expectedValue := range expectedHeaders {
			if r.Header.Get(header) != expectedValue {
				t.Errorf("Expected header %s: %s, got %s", header, expectedValue, r.Header.Get(header))
			}
		}
		
		// Parse request body
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		
		// Verify request body
		if req.ClientID != "Iv1.b507a08c87ecfe98" {
			t.Errorf("Expected client_id: Iv1.b507a08c87ecfe98, got %s", req.ClientID)
		}
		if req.Scopes != "read:user" {
			t.Errorf("Expected scopes: read:user, got %s", req.Scopes)
		}
		
		// Return mock response
		response := LoginResponse{
			DeviceCode:      "test-device-code",
			Interval:        5,
			UserCode:        "TEST-CODE",
			VerificationURI: "https://github.com/login/device",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override the login endpoint for testing
	originalEndpoint := github_login_endpoint
	github_login_endpoint = server.URL
	defer func() { github_login_endpoint = originalEndpoint }()
	
	// Test the Login function
	loginResp, err := Login()
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	
	// Verify response
	if loginResp.DeviceCode != "test-device-code" {
		t.Errorf("Expected device_code: test-device-code, got %s", loginResp.DeviceCode)
	}
	if loginResp.UserCode != "TEST-CODE" {
		t.Errorf("Expected user_code: TEST-CODE, got %s", loginResp.UserCode)
	}
	if loginResp.Interval != 5 {
		t.Errorf("Expected interval: 5, got %d", loginResp.Interval)
	}
}

func TestLoginError(t *testing.T) {
	// Mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid client ID",
				"type":    "invalid_request",
				"code":    "bad_client_id",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override the login endpoint for testing
	originalEndpoint := github_login_endpoint
	github_login_endpoint = server.URL
	defer func() { github_login_endpoint = originalEndpoint }()
	
	// Test the Login function with error
	_, err := Login()
	if err == nil {
		t.Fatal("Expected Login to return an error, but it succeeded")
	}
	
	expectedError := "API error: Invalid client ID (code: bad_client_id, type: invalid_request)"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %s", expectedError, err.Error())
	}
}

func TestAuthenticate(t *testing.T) {
	// Mock server for GitHub authentication endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		var req AuthenticationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		
		// Verify request structure
		if req.ClientID != "Iv1.b507a08c87ecfe98" {
			t.Errorf("Expected client_id: Iv1.b507a08c87ecfe98, got %s", req.ClientID)
		}
		if req.DeviceCode != "test-device-code" {
			t.Errorf("Expected device_code: test-device-code, got %s", req.DeviceCode)
		}
		if req.GrantType != "urn:ietf:params:oauth:grant-type:device_code" {
			t.Errorf("Expected grant_type: urn:ietf:params:oauth:grant-type:device_code, got %s", req.GrantType)
		}
		
		// Return mock response
		response := AuthenticationResponse{
			AccessToken: "test-access-token",
			TokenType:   "bearer",
			Scope:       "read:user",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override the authentication endpoint for testing
	originalEndpoint := github_authentication_endpoint
	github_authentication_endpoint = server.URL
	defer func() { github_authentication_endpoint = originalEndpoint }()
	
	// Test the Authenticate function
	loginResp := LoginResponse{
		DeviceCode: "test-device-code",
	}
	
	authResp, err := Authenticate(loginResp)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	
	// Verify response
	if authResp.AccessToken != "test-access-token" {
		t.Errorf("Expected access_token: test-access-token, got %s", authResp.AccessToken)
	}
	if authResp.TokenType != "bearer" {
		t.Errorf("Expected token_type: bearer, got %s", authResp.TokenType)
	}
}

func TestGetSessionToken(t *testing.T) {
	// Mock server for GitHub session endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		
		// Verify authorization header
		authHeader := r.Header.Get("authorization")
		if authHeader != "token test-access-token" {
			t.Errorf("Expected authorization: token test-access-token, got %s", authHeader)
		}
		
		// Return mock response with a token that contains exp parameter
		response := SessionResponse{
			Token: "test-session-token;exp=1234567890;sig=abcdef",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override the session endpoint for testing
	originalEndpoint := github_session_endpoint
	github_session_endpoint = server.URL
	defer func() { github_session_endpoint = originalEndpoint }()
	
	// Test the GetSessionToken function
	sessionResp, err := GetSessionToken("test-access-token")
	if err != nil {
		t.Fatalf("GetSessionToken failed: %v", err)
	}
	
	// Verify response
	expectedToken := "test-session-token;exp=1234567890;sig=abcdef"
	if sessionResp.Token != expectedToken {
		t.Errorf("Expected token: %s, got %s", expectedToken, sessionResp.Token)
	}
	if sessionResp.ExpiresAt != 1234567890 {
		t.Errorf("Expected expires_at: 1234567890, got %d", sessionResp.ExpiresAt)
	}
}

func TestChat(t *testing.T) {
	// Mock server for GitHub chat completion endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		
		// Verify authorization header
		authHeader := r.Header.Get("authorization")
		if authHeader != "Bearer test-session-token" {
			t.Errorf("Expected authorization: Bearer test-session-token, got %s", authHeader)
		}
		
		var req CompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		
		// Verify request structure
		if req.Model != "test-model" {
			t.Errorf("Expected model: test-model, got %s", req.Model)
		}
		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "user" {
			t.Errorf("Expected role: user, got %s", req.Messages[0].Role)
		}
		if req.Messages[0].Content != "Hello" {
			t.Errorf("Expected content: Hello, got %s", req.Messages[0].Content)
		}
		
		// Return mock response
		response := CompletionResponse{
			ID:      "chatcmpl-test123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "test-model",
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
				PromptTokens:     5,
				CompletionTokens: 7,
				TotalTokens:      12,
			},
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override the completion endpoint for testing
	originalEndpoint := github_completion_endpoint
	github_completion_endpoint = server.URL
	defer func() { github_completion_endpoint = originalEndpoint }()
	
	// Test the Chat function
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}
	
	var receivedResponse CompletionResponse
	err := Chat("test-session-token", messages, "test-model", 0.7, 0.9, 1, false, func(response CompletionResponse) error {
		receivedResponse = response
		return nil
	})
	
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	
	// Verify response
	if receivedResponse.ID != "chatcmpl-test123" {
		t.Errorf("Expected ID: chatcmpl-test123, got %s", receivedResponse.ID)
	}
	if receivedResponse.Object != "chat.completion" {
		t.Errorf("Expected object: chat.completion, got %s", receivedResponse.Object)
	}
	if receivedResponse.Model != "test-model" {
		t.Errorf("Expected model: test-model, got %s", receivedResponse.Model)
	}
	if len(receivedResponse.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(receivedResponse.Choices))
	}
	if receivedResponse.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Expected content: Hello! How can I help you?, got %s", receivedResponse.Choices[0].Message.Content)
	}
	if receivedResponse.Usage.TotalTokens != 12 {
		t.Errorf("Expected total_tokens: 12, got %d", receivedResponse.Usage.TotalTokens)
	}
}

func TestChatError(t *testing.T) {
	// Mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		response := map[string]interface{}{
			"error": map[string]string{
				"message": "Invalid token",
				"type":    "unauthorized",
				"code":    "invalid_token",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Override the completion endpoint for testing
	originalEndpoint := github_completion_endpoint
	github_completion_endpoint = server.URL
	defer func() { github_completion_endpoint = originalEndpoint }()
	
	// Test the Chat function with error
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}
	
	err := Chat("invalid-token", messages, "test-model", 0.7, 0.9, 1, false, func(response CompletionResponse) error {
		return nil
	})
	
	if err == nil {
		t.Fatal("Expected Chat to return an error, but it succeeded")
	}
	
	expectedError := "API error: Invalid token (code: invalid_token, type: unauthorized)"
	if err.Error() != expectedError {
		t.Errorf("Expected error: %s, got %s", expectedError, err.Error())
	}
}

// Test data structure validation
func TestCompletionRequestValidation(t *testing.T) {
	tests := []struct {
		name     string
		request  CompletionRequest
		expected bool
	}{
		{
			name: "Valid request",
			request: CompletionRequest{
				Model:       "test-model",
				Messages:    []Message{{Role: "user", Content: "Hello"}},
				Temperature: 0.7,
				TopP:        0.9,
				N:           1,
			},
			expected: true,
		},
		{
			name: "Empty messages",
			request: CompletionRequest{
				Model:       "test-model",
				Messages:    []Message{},
				Temperature: 0.7,
				TopP:        0.9,
				N:           1,
			},
			expected: true, // Empty messages should be allowed for validation
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling/unmarshaling
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Errorf("Failed to marshal request: %v", err)
			}
			
			var unmarshaled CompletionRequest
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Errorf("Failed to unmarshal request: %v", err)
			}
			
			// Verify structure integrity
			if unmarshaled.Model != tt.request.Model {
				t.Errorf("Model mismatch after marshal/unmarshal: expected %s, got %s", tt.request.Model, unmarshaled.Model)
			}
		})
	}
}

// Benchmark tests
func BenchmarkLoginRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LoginResponse{
			DeviceCode:      "test-device-code",
			Interval:        5,
			UserCode:        "TEST-CODE",
			VerificationURI: "https://github.com/login/device",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	originalEndpoint := github_login_endpoint
	github_login_endpoint = server.URL
	defer func() { github_login_endpoint = originalEndpoint }()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Login()
		if err != nil {
			b.Errorf("Login failed: %v", err)
		}
	}
}