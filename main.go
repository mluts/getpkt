package main

import (
	"fmt"
	"log"
	"os"
	"path"
)

const (
	configFilePath   = "$HOME/.config/getpkt/config.json"
	articlesFilePath = "$HOME/.config/getpkt/articles.json"
	authURLTemplate  = "https://getpocket.com/auth/authorize?request_token={{.Code}}&redirect_uri={{.RedirectURL}}"
)

var defaultRedirectURL = "http://localhost:9998"

func showUsage() {
	fmt.Printf(`
Usage:
	%s command

Commands:
	auth
	sync
	list
`, path.Base(os.Args[0]))
}

func collectArticles(config *appConfig) (result []*Article, err error) {
	step := 5000
	offset := 0

	request := RetrieveRequest{
		ConsumerKey: config.ConsumerKey,
		AccessToken: config.AccessToken,
		Sort:        SortNewest,
	}

	result = make([]*Article, 0)

	fmt.Print("Syncing")
	for {
		fmt.Print(".")
		request.Count = step
		request.Offset = offset
		response := RetrieveResponse{}

		err := retrieve(&request, &response)
		if err != nil {
			return nil, err
		}

		for _, article := range response.List {
			result = append(result, article)
		}

		offset += step

		if len(response.List) != step {
			break
		}
	}
	fmt.Print("\n")

	return result, nil
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
	case os.Args[1] == "sync":
		config := mustInitConfig()

		articles, err := collectArticles(config)
		if err != nil {
			log.Fatalf("Failed to download articles: %v", err)
		}

		err = writeJSON(articlesFilePath, articles)
		if err != nil {
			log.Fatalf("Failed to save articles: %v", err)
		}
	case os.Args[1] == "list":
	default:
		showUsage()
		os.Exit(1)
	}
}
