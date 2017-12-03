package main

import (
	"fmt"
	"log"
	"sort"
)

func loadArticles() (articles Articles, err error) {
	articles = Articles{}
	if isReadable(articlesFilePath) {
		err = readJSON(articlesFilePath, &articles)
	} else {
		err = fmt.Errorf("%s is not readable", articlesFilePath)
	}
	return articles, err
}

func collectArticles(config *appConfig, limit int, step int) (result Articles, err error) {
	offset := 0

	request := RetrieveRequest{}
	request.ConsumerKey = config.ConsumerKey
	request.AccessToken = config.AccessToken
	request.Sort = SortNewest

	result = make([]*Article, 0)

	log.Print("Downloading")
	for {
		log.Print(".")
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

		if len(response.List) != step || (limit > 0 && len(response.List) >= limit) {
			break
		}
	}
	log.Print("\n")

	sort.Sort(result)

	return result, nil
}
