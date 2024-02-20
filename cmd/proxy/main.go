package main

import (
	"bufio"
	"os"
	"time"

	"github.com/maxneuvians/go-copilot-proxy/pkg"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting proxy...")

	// Check if .github_token file exists
	file, err := os.Open(".github_token")

	var token string

	if err != nil {
		token, err = doLoginProcess()

		if err != nil {
			log.Error().Msgf("Error getting token in: %s", err)
			return
		}

		// Write the token to a file
		file, err = os.Create(".github_token")

		if err != nil {
			log.Error().Msgf("Error creating token file: %s", err)
			return
		}

		_, err = file.WriteString(token)

		if err != nil {
			log.Error().Msgf("Error writing token to file: %s", err)
			return
		}

	}

	// If the file exists, read the first line
	r := bufio.NewReader(file)
	buffer, _, err := r.ReadLine()

	if err != nil {
		log.Error().Msgf("Error reading token from file: %s", err)
		return
	}

	token = string(buffer)

	// Get a session token from the token
	sessionResponse, err := pkg.GetSessionToken(token)

	if err != nil {
		log.Error().Msgf("Error getting session token: %s", err)
		return
	}

	// Make a request to the copilot API
	completionResponse, err := pkg.DoCompletion(sessionResponse.Token, "Can your write a python function that takes a list of strings and returns a list of strings with the first letter of each word capitalized?")

	if err != nil {
		log.Error().Msgf("Error getting completion: %s", err)
		return
	}

	log.Info().Msgf("Completion: %s", completionResponse)
}

func doLoginProcess() (string, error) {
	loginResponse, err := pkg.Login()

	if err != nil {
		log.Error().Msgf("Error logging in: %s", err)
		return "", err
	}

	// Sleep for the interval time
	time.Sleep(time.Duration(loginResponse.Interval+1) * time.Second)

	var authResponse pkg.AuthenticationResponse

	for {
		authResponse, err = pkg.Authenticate(loginResponse)

		if err != nil {
			log.Error().Msgf("Error authenticating: %s", err)
			return "", err
		}

		if authResponse.AccessToken != "" {
			log.Info().Msg("Authenticated!")
			break
		}

		// If the interval is 0, set it to 5
		if authResponse.Interval == 0 {
			authResponse.Interval = 5
		}

		// Sleep for the interval time
		time.Sleep(time.Duration(authResponse.Interval+1) * time.Second)
	}

	return authResponse.AccessToken, nil
}
