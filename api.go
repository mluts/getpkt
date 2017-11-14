package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"text/template"
)

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
