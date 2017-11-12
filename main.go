package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"text/template"
)

const (
	configFilePath  = "$HOME/.config/getpkt/config.json"
	authURLTemplate = "https://getpocket.com/auth/authorize?request_token={{.Code}}&redirect_uri={{.RedirectURL}}"
)

var defaultRedirectURL = "http://localhost:9998"

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

type appConfig struct {
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`
}

func makeAuthURL(code, redirectURL string) string {
	var data = struct {
		Code        string
		RedirectURL string
	}{
		url.QueryEscape(code),
		url.QueryEscape(redirectURL),
	}

	buf := bytes.NewBuffer([]byte{})
	tpl := template.Must(template.New("authURL").Parse(authURLTemplate))
	err := tpl.Execute(buf, data)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func readConfig(path string) (config *appConfig, err error) {
	var (
		file *os.File
	)
	path = os.ExpandEnv(path)
	config = &appConfig{}

	file, err = os.Open(path)
	if err != nil {
		return nil, err
	}

	io := bufio.NewReader(file)
	decoder := json.NewDecoder(io)
	for decoder.More() {
		err = decoder.Decode(config)
		if err != nil {
			return nil, err
		}
	}

	return config, err
}

func isReadable(path string) bool {
	path = os.ExpandEnv(path)
	_, err := os.Open(path)
	return err == nil
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

func obtainCode(consumerKey string) (code string, err error) {
	var (
		request = codeRequest{
			consumerKey,
			defaultRedirectURL,
		}
		response = codeResponse{}
	)

	err = jsonRequest(
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

	err := jsonRequest(
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

func showUsage() {
	fmt.Printf(`
Usage:
	%s command

Commands:
	auth
	list
`, path.Base(os.Args[0]))
}

func readConsumerKey() string {
	log.Print("Please enter consumer key: ")
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		log.Fatal(err)
	}
	return string(line)
}

func writeConfig(path string, config *appConfig) (err error) {
	file, err := os.OpenFile(
		os.ExpandEnv(path),
		os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(config)
}

func initConfig() (config *appConfig) {
	var err error

	if isReadable(configFilePath) {
		config, err = readConfig(configFilePath)
		if err != nil {
			log.Fatalf("Failed to read a config: %v", err)
		}
	} else {
		config = &appConfig{
			ConsumerKey: readConsumerKey(),
		}
		err := writeConfig(configFilePath, config)
		if err != nil {
			log.Fatalf("Failed to write a config: %v", err)
		}
	}

	return config
}

func main() {
	switch {
	case len(os.Args) == 1:
		showUsage()
		os.Exit(1)
	case os.Args[1] == "auth":
		config := initConfig()
		accessToken, err := authenticate(config.ConsumerKey)
		if err != nil {
			log.Fatalf("Failed to authenticate: %v", err)
		}
		config.AccessToken = accessToken
		writeConfig(configFilePath, config)
		log.Println("Written to config")
	case os.Args[1] == "list":
	default:
		showUsage()
		os.Exit(1)
	}
}
