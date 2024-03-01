package pkg

type AuthenticationRequest struct {
	ClientID   string `json:"client_id"`
	DeviceCode string `json:"device_code"`
	GrantType  string `json:"grant_type"`
}

type AuthenticationResponse struct {
	AccessToken string `json:"access_token"`
	Interval    int    `json:"interval"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type Choice struct {
	FinishReason string  `json:"finish_reason,omitempty"`
	Index        int64   `json:"index,omitempty"`
	Message      Message `json:"message,omitempty"`
	Delta        Message `json:"delta,omitempty"`
}

type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream,omitempty"`
	Temperature float64   `json:"temperature"`
	TopP        float64   `json:"top_p"`
	N           int64     `json:"n"`
}

type CompletionResponse struct {
	Choices []Choice `json:"choices"`
	Created int64    `json:"created,omitempty"`
	ID      string   `json:"id"`
	Usage   Usage    `json:"usage,omitempty"`
}

type CompletionResponseHandler func(CompletionResponse) error

type LoginRequest struct {
	ClientID string `json:"client_id"`
	Scopes   string `json:"scopes"`
}

type LoginResponse struct {
	DeviceCode      string `json:"device_code"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type SessionResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

type Usage struct {
	CompletionTokens int64 `json:"completion_tokens"`
	PromptTokens     int64 `json:"prompt_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
}
