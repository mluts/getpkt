package main

import (
	"fmt"
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

func collectArticles(config *appConfig) (result Articles, err error) {
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

	sort.Sort(result)

	return result, nil
}