package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type codeRequest struct {
	ConsumerKey string `json:"consumer_key"`
	RedirectURL string `json:"redirect_uri"`
}

type codeResponse struct {
	Code string `json:"code"`
}

type authorizeRequest struct {
	ConsumerKey string `json:"consumer_key"`
	Code        string `json:"code"`
}

type authorizeResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

func obtainCode(consumerKey string) (code string, err error) {
	var (
		request = codeRequest{
			consumerKey,
			defaultRedirectURL,
		}
		response = codeResponse{}
	)

	err = doJSONRequest(
		"https://getpocket.com/v3/oauth/request",
		&request,
		&response,
	)
	if err != nil {
		return "", fmt.Errorf("Failed to obtain code: %v", err)
	}

	return response.Code, nil
}

func authorize(consumerKey, code string) (*authorizeResponse, error) {
	var (
		request = authorizeRequest{
			consumerKey,
			code,
		}
		response = authorizeResponse{}
	)

	err := doJSONRequest(
		"https://getpocket.com/v3/oauth/authorize",
		&request,
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to authorize: %v", err)
	}

	return &response, nil
}

func authenticate(consumerKey string) (accessToken string, err error) {
	code, err := obtainCode(consumerKey)
	if len(code) == 0 {
		return "", fmt.Errorf("Received no code")
	}

	log.Println("Please proceed to:", makeAuthURL(code, defaultRedirectURL))
	log.Print("Press enter when ready:")
	bufio.NewReader(os.Stdin).ReadLine()

	response, err := authorize(consumerKey, code)
	if err != nil {
		return "", fmt.Errorf("Failed to authorize: %v", err)
	}

	return response.AccessToken, nil
}
