package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// Article represents a single article in /v3/get list
type Article struct {
	ItemID        string `json:"item_id"`
	ResolvedID    string `json:"resolved_id"`
	GivenURL      string `json:"given_url"`
	GivenTitle    string `json:"given_title"`
	Favorite      string `json:"favorite"`
	Status        string `json:"status"`
	ResolvedURL   string `json:"resolved_url"`
	ResolvedTitle string `json:"resolved_title"`
	Excerpt       string `json:"excerpt"`
	IsArticle     string `json:"is_article"`
	HasVideo      string `json:"has_video"`
	HasImage      string `json:"has_image"`
	WordsCount    string `json:"words_count"`
	TimeAdded     string `json:"time_added"`
}

type ApiCredentials struct {
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`
}

// TimeAddedTime returns TimeAdded as time.Time
func (a *Article) TimeAddedTime() time.Time {
	t, err := strconv.Atoi(a.TimeAdded)
	if err != nil {
		panic(err)
	}
	return time.Unix(int64(t), 0)
}

// Articles is a collection of Article
type Articles []*Article

func (a Articles) Len() int {
	return len(a)
}

func (a Articles) Less(i, j int) bool {
	return strings.Compare(a[i].TimeAdded, a[j].TimeAdded) == 1
}

func (a Articles) Swap(i, j int) {
	buf := a[i]
	a[i] = a[j]
	a[j] = buf
}

const (
	// StateUnread requests only unread articles
	StateUnread = "unread"
	// StateArchived requests only archived articles
	StateArchived = "archive"
	// StateAll requests unread and archived articles
	StateAll = "all"
	// TagUntagged requests only articles without tags
	TagUntagged = "_untagged_"
	// DetailTypeSimple requests only basic details
	DetailTypeSimple = "simple"
	// DetailTypeComplete requests all details
	DetailTypeComplete = "complete"
	// SortNewest sorts from new to old
	SortNewest = "newest"
	// SortOldest sorts from old to new
	SortOldest = "oldest"
	// ActionArchive is a part of ModifyAction
	ActionArchive = "archive"
)

// RetrieveRequest holds the json scheme for /v3/get request
type RetrieveRequest struct {
	State      string `json:"state"`
	Favorite   int    `json:"favorite"`
	Tag        string `json:"tag"`
	Count      int    `json:"count"`
	Offset     int    `json:"offset"`
	Sort       string `json:"sort"`
	DetailType string `json:"detailType"`
	ApiCredentials
}

// RetrieveResponse holds the json scheme for /v3/get response
type RetrieveResponse struct {
	Status int `json:"status"`
	List   map[string]*Article
}

// ModifyAction is a part of ModifyRequest
type ModifyAction struct {
	Action string `json:"action"`
	ItemID string `json:"item_id"`
	Time   int64  `json:"time"`
}

// ModifyRequest holds the json scheme for /v3/send request
type ModifyRequest struct {
	Actions []*ModifyAction `json:"actions"`
	ApiCredentials
}

// ModifyResponse holds the json scheme for /v3/send response
type ModifyResponse struct {
	ActionResults []bool `json:"action_results"`
	Status        int    `json:"status"`
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

func doJSONRequest(url string, reqJSON interface{}, respJSON interface{}) (err error) {
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

func retrieve(request *RetrieveRequest, response *RetrieveResponse) error {
	return doJSONRequest(
		"https://getpocket.com/v3/get",
		request,
		response,
	)
}

func modify(request *ModifyRequest, response *ModifyResponse) error {
	return doJSONRequest(
		"https://getpocket.com/v3/send",
		request,
		response,
	)
}

func archive(credentials ApiCredentials, itemID string) (err error) {
	actions := []*ModifyAction{
		&ModifyAction{
			ActionArchive,
			itemID,
			time.Now().Unix(),
		},
	}
	request := &ModifyRequest{
		actions, credentials,
	}
	response := &ModifyResponse{}

	err = modify(request, response)
	if err != nil {
		return err
	}

	if response.Status == 0 {
		return fmt.Errorf("Failed to do an update on item %s", itemID)
	}

	return nil
}
