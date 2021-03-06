package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
)

const (
	configFilePath       = "$HOME/.config/getpkt/config.json"
	articlesFilePath     = "$HOME/.config/getpkt/articles.json"
	authURLTemplate      = "https://getpocket.com/auth/authorize?request_token={{.Code}}&redirect_uri={{.RedirectURL}}"
	articlesDownloadStep = 4000
)

var (
	defaultRedirectURL = "http://localhost:9998"
	useCache           = false
	articlesLimit      int
)

func showUsage() {
	log.Printf(`
Usage:
	%s command

Commands:
	auth
	sync
	list
	archive ITEM_ID

Flags:
`, path.Base(os.Args[0]))

	flag.PrintDefaults()
}

func printArticle(article *Article) {
	fmt.Printf(
		"URL: %s\nTitle: %s\nID: %s\nTime Added: %v\n\n",
		article.ResolvedURL,
		article.ResolvedTitle,
		article.ItemID,
		article.TimeAddedTime(),
	)
}

func cmdAuth() {
	config := initConfig()
	accessToken, err := authenticate(config.ConsumerKey)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	config.AccessToken = accessToken
	writeConfig(configFilePath, config)
}

func cmdSync() {
	config := mustInitConfig()

	articles, err := collectArticles(config, 0, articlesDownloadStep)
	if err != nil {
		log.Fatalf("Failed to download articles: %v", err)
	}

	err = writeJSON(articlesFilePath, articles)
	if err != nil {
		log.Fatalf("Failed to save articles: %v", err)
	}
}

func cmdList() {
	if useCache {
		articles, err := loadArticles()
		if err != nil {
			log.Fatalf("Failed to load articles: %v", err)
		}
		for i, article := range articles {
			if articlesLimit > 0 && i >= articlesLimit {
				break
			}
			printArticle(article)
		}
	} else {
		config := initConfig()
		articles, err := collectArticles(config, articlesLimit, articlesLimit)
		if err != nil {
			log.Fatalf("Failed to get articles: %v", err)

		}
		for _, article := range articles {
			printArticle(article)
		}
	}
}

func cmdRand() {
	var (
		i, max *big.Int
		err    error
	)
	articles, err := loadArticles()
	if err != nil {
		log.Fatalf("Failed to load articles: %v", err)
	}

	max = big.NewInt(int64(len(articles)))

	i, err = rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}
	article := articles[i.Int64()]
	printArticle(article)
}

func cmdArchive(item string) {
	config := initConfig()
	err := archive(ApiCredentials(*config), item)
	if err != nil {
		log.Fatalf("Failed to archive item %s: %v", item, err)
	} else {
		log.Print("Archived!")
	}
}

func init() {
	flag.Usage = showUsage
	flag.IntVar(&articlesLimit, "limit", 10, "Articles limit")
	flag.BoolVar(&useCache, "cache", false, "Use articles cache")
	flag.Parse()

	log.SetOutput(os.Stderr)
}

func main() {
	cmd := flag.Arg(0)

	if len(flag.Args()) == 0 {
		showUsage()
		os.Exit(1)
	}

	switch cmd {
	case "auth":
		cmdAuth()
	case "sync":
		cmdSync()
	case "list":
		cmdList()
	case "rand":
		cmdRand()
	case "archive":
		item := flag.Arg(1)
		if len(item) == 0 {
			showUsage()
			os.Exit(1)
		}
		cmdArchive(flag.Arg(1))
	default:
		showUsage()
		os.Exit(1)
	}
}
