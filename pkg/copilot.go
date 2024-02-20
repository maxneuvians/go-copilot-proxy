package pkg

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

var Completion_max_tokens = 1000
var Completion_temperature = 0.3
var Completion_top_p = 0.9
var Completion_n = int64(1)
var Completion_stop = []string{"\n"}
var Completion_nwo = "github/copilot.vim"
var Completion_stream = false
var Completion_language = "python"

var editor_client_id = "Iv1.b507a08c87ecfe98"
var editor_version = "vscode/1.83.1"
var editor_plugin_version = "copilot-chat/0.8.0"

var github_authentication_endpoint = "https://github.com/login/oauth/access_token"
var github_completion_endpoint = "https://api.githubcopilot.com/chat/completions"
var github_login_endpoint = "https://github.com/login/device/code"
var github_session_endpoint = "https://api.github.com/copilot_internal/v2/token"

var user_agent = "githubCopilot/1.155.0"

func Authenticate(login LoginResponse) (AuthenticationResponse, error) {
	log.Debug().Msg("Authenticating...")

	var authResponse AuthenticationResponse

	body := AuthenticationRequest{
		ClientID:   editor_client_id,
		DeviceCode: login.DeviceCode,
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		log.Error().Msgf("Error marshalling json: %s", err)
		return authResponse, err
	}

	req, err := http.NewRequest(http.MethodPost, github_authentication_endpoint, bytes.NewBuffer(jsonBody))

	if err != nil {
		log.Error().Msgf("Error creating request: %s", err)
		return authResponse, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("editor-version", editor_version)
	req.Header.Set("editor-plugin-version", editor_plugin_version)
	req.Header.Set("user-agent", user_agent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Error().Msgf("Error sending request: %s", err)
		return authResponse, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&authResponse)

	if err != nil {
		log.Error().Msgf("Error decoding response: %s", err)
		return authResponse, err
	}

	log.Debug().Msgf("Access token: %s", authResponse.AccessToken)

	return authResponse, nil
}

func DoCompletion(token string, prompt string) (string, error) {
	log.Debug().Msg("Doing completion...")

	var completionResponse CompletionResponse

	body := CompletionRequest{
		Model: "gpt-4",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: float64(Completion_temperature),
		TopP:        float64(Completion_top_p),
		N:           Completion_n,
		Stream:      Completion_stream,
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		log.Error().Msgf("Error marshalling json: %s", err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, github_completion_endpoint, bytes.NewBuffer(jsonBody))

	req.Header.Set("editor-version", editor_version)
	req.Header.Set("editor-plugin-version", editor_plugin_version)
	req.Header.Set("user-agent", user_agent)

	if err != nil {
		log.Error().Msgf("Error creating request: %s", err)
		return "", err
	}

	req.Header.Set("authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Error().Msgf("Error sending request: %s", err)
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Error().Msgf("Error reading response body: %s", err)
		return "", err
	}

	err = json.Unmarshal(bodyBytes, &completionResponse)

	if err != nil {
		log.Error().Msgf("Error decoding response: %s", err)
		return "", err
	}

	log.Debug().Msgf("Completion: %s", completionResponse.Choices[0].Message.Content)

	return "", nil
}

func GetSessionToken(accessToken string) (SessionResponse, error) {
	log.Debug().Msg("Getting session token...")

	var sessionResponse SessionResponse

	req, err := http.NewRequest(http.MethodGet, github_session_endpoint, nil)

	if err != nil {
		log.Error().Msgf("Error creating request: %s", err)
		return sessionResponse, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("authorization", "token "+accessToken)
	req.Header.Set("editor-version", editor_version)
	req.Header.Set("editor-plugin-version", editor_plugin_version)
	req.Header.Set("user-agent", user_agent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Error().Msgf("Error sending request: %s", err)
		return sessionResponse, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&sessionResponse)

	if err != nil {
		log.Error().Msgf("Error decoding response: %s", err)
		return sessionResponse, err
	}

	log.Debug().Msgf("Session token: %s", sessionResponse.Token)

	return sessionResponse, nil
}

func Login() (LoginResponse, error) {
	log.Debug().Msg("Logging in...")

	var loginResponse LoginResponse

	body := LoginRequest{
		ClientID: editor_client_id,
		Scopes:   "read:user",
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		return loginResponse, err
	}

	req, err := http.NewRequest(http.MethodPost, github_login_endpoint, bytes.NewBuffer(jsonBody))

	if err != nil {
		return loginResponse, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("editor-version", editor_version)
	req.Header.Set("editor-plugin-version", editor_plugin_version)
	req.Header.Set("user-agent", user_agent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return loginResponse, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&loginResponse)

	if err != nil {
		return loginResponse, err
	}

	log.Debug().Msgf("Device code: %s", loginResponse.DeviceCode)
	log.Debug().Msgf("Interval: %d", loginResponse.Interval)
	log.Debug().Msgf("User code: %s", loginResponse.UserCode)
	log.Debug().Msgf("Verification URI: %s", loginResponse.VerificationURI)

	return loginResponse, nil
}
