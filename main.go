package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

const configFilePath = "$HOME/.config/getpkt/config.json"

type oauthRequest struct {
	ConsumerKey string `json:"consumer_key"`
	RedirectURI string `json:"redirect_uri"`
}

type codeResponse struct {
	Code string `json:"code"`
}

var config = struct {
	ConsumerKey string `json:"consumer_key"`
}{}

func readConfig() (err error) {
	var (
		file *os.File
		path string
	)
	path = os.ExpandEnv(configFilePath)

	file, err = os.Open(path)
	if err != nil {
		return err
	}

	io := bufio.NewReader(file)
	decoder := json.NewDecoder(io)
	for decoder.More() {
		err = decoder.Decode(&config)
		if err != nil {
			break
		}
	}

	return err
}

func jsonRequest(url string, reqJSON interface{}, respJSON interface{}) (err error) {
	var (
		buf     = bytes.NewBuffer([]byte{})
		client  = http.Client{}
		request *http.Request
	)

	encoder := json.NewEncoder(buf)
	err = encoder.Encode(reqJSON)
	if err != nil {
		return err
	}

	request, err = http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return err
	}
	request.Header.Set("X-Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"URL: %s\nHTTP Status: %d\nX-Error-Code: %s\nX-Error: %s",
			url,
			resp.StatusCode,
			resp.Header.Get("X-Error-Code"),
			resp.Header.Get("X-Error"),
		)
	}

	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		err = decoder.Decode(respJSON)
		if err != nil {
			err = fmt.Errorf("Failed to decode json response: %v", err)
			break
		}
	}

	return err
}

func authenticate(consumerKey string) (accessToken string, err error) {
	var (
		auth = oauthRequest{
			config.ConsumerKey,
			"http://localhost:9998",
		}
		code         string
		authResponse = codeResponse{}
	)

	err = jsonRequest(
		"https://getpocket.com/v3/oauth/request",
		&auth,
		&authResponse,
	)
	if err != nil {
		return "", fmt.Errorf("Failed to make a request for a code: %v", err)
	}

	code = authResponse.Code
	if len(code) == 0 {
		return "", fmt.Errorf("Received no code")
	}

	return accessToken, err
}

func main() {
	err := readConfig()
	if err != nil {
		log.Fatal(err)
	}
	code, err := authenticate(config.ConsumerKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Code is ", code)
}
