package pkg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/rs/zerolog/log"
)

var Completion_temperature = 0.3
var Completion_top_p = 0.9
var Completion_n = int64(1)
var Completion_stream = true

var editor_client_id = "Iv1.b507a08c87ecfe98"
var editor_version = "vscode/1.83.1"
var editor_plugin_version = "copilot-chat/0.8.0"

var github_authentication_endpoint = "https://github.com/login/oauth/access_token"
var github_completion_endpoint = "https://api.githubcopilot.com/chat/completions"
var github_login_endpoint = "https://github.com/login/device/code"
var github_session_endpoint = "https://api.github.com/copilot_internal/v2/token"

var user_agent = "githubCopilot/1.155.0"

func Authenticate(login LoginResponse) (AuthenticationResponse, error) {
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
	return authResponse, nil
}

func Chat(token string, messages []Message, stream bool, callback CompletionResponseHandler) error {
	model := os.Getenv("MODEL")
	if model == "" {
		model = "claude-3.7-sonnet"
	}

	body := CompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: float64(Completion_temperature),
		TopP:        float64(Completion_top_p),
		N:           Completion_n,
		Stream:      stream,
	}

	jsonBody, err := json.Marshal(body)

	if err != nil {
		log.Error().Msgf("Error marshalling json: %s", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, github_completion_endpoint, bytes.NewBuffer(jsonBody))

	req.Header.Set("editor-version", editor_version)
	req.Header.Set("editor-plugin-version", editor_plugin_version)
	req.Header.Set("user-agent", user_agent)

	if err != nil {
		log.Error().Msgf("Error creating request: %s", err)
		return err
	}

	req.Header.Set("authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Error().Msgf("Error sending request: %s", err)
		return err
	}

	defer resp.Body.Close()

	var completionResponse CompletionResponse

	if stream {

		scn := bufio.NewScanner(resp.Body)

		scnBuf := make([]byte, 0, 4096)
		scn.Buffer(scnBuf, cap(scnBuf))

		for scn.Scan() {
			b := scn.Bytes()

			if !bytes.HasPrefix(b, []byte("data:")) {
				continue
			}

			b = bytes.TrimSpace(b[5:])

			if bytes.Equal(b, []byte("[DONE]")) {
				return nil
			}

			err = json.Unmarshal(b, &completionResponse)

			if err != nil {
				log.Error().Msgf("Error decoding response: %s", err)
				return err
			}

			if len(completionResponse.Choices) == 0 {
				continue
			}

			callback(completionResponse)
		}
	}

	err = json.NewDecoder(resp.Body).Decode(&completionResponse)

	if err != nil {
		log.Error().Msgf("Error decoding response: %s", err)
		return err
	}

	callback(completionResponse)
	return nil
}

func GetSessionToken(accessToken string) (SessionResponse, error) {
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

	re := regexp.MustCompile(`exp=(\d+)`)
	matches := re.FindStringSubmatch(sessionResponse.Token)

	if len(matches) < 2 {
		log.Error().Msgf("Error parsing token: %s", err)
		return sessionResponse, err
	}

	exp, err := strconv.ParseInt(matches[1], 10, 64)

	if err != nil {
		log.Error().Msgf("Error parsing token: %s", err)
		return sessionResponse, err
	}

	sessionResponse.ExpiresAt = exp

	return sessionResponse, nil
}

func Login() (LoginResponse, error) {
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

	return loginResponse, nil
}
