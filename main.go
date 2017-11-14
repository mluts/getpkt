package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

const (
	configFilePath  = "$HOME/.config/getpkt/config.json"
	authURLTemplate = "https://getpocket.com/auth/authorize?request_token={{.Code}}&redirect_uri={{.RedirectURL}}"
)

var defaultRedirectURL = "http://localhost:9998"

func showUsage() {
	fmt.Printf(`
Usage:
	%s command

Commands:
	auth
	list
`, path.Base(os.Args[0]))
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
		log.Println("List")
	default:
		showUsage()
		os.Exit(1)
	}
}
