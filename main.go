package main

import (
	"flag"
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

var (
	defaultRedirectURL = "http://localhost:9998"
	articlesLimit      int
)

func showUsage() {
	fmt.Printf(`
Usage:
	%s command

Commands:
	auth
	sync
	list

Flags:
`, path.Base(os.Args[0]))

	flag.PrintDefaults()
}

func init() {
	flag.Usage = showUsage
	flag.IntVar(&articlesLimit, "limit", 10, "Articles limit")
	flag.Parse()
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
		articles, err := loadArticles()
		if err != nil {
			log.Fatalf("Failed to load articles: %v", err)
		}
		for i, article := range articles {
			if i >= articlesLimit {
				break
			}
			fmt.Printf("URL: %s\nTitle: %s\nID: %s\n\n", article.ResolvedURL, article.ResolvedTitle, article.ItemID)
		}
	default:
		showUsage()
		os.Exit(1)
	}
}
