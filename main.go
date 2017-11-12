package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

const configFilePath = "$HOME/.config/getpkt/config.json"

var config = struct {
	ConsumerKey string `json:consumer_key`
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

func main() {
	err := readConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(config.ConsumerKey)
}
