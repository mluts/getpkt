package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
)

type appConfig struct {
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`
}

func isReadable(path string) bool {
	path = os.ExpandEnv(path)
	_, err := os.Open(path)
	return err == nil
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
	return writeJSON(path, config)
}

func writeJSON(path string, object interface{}) (err error) {
	file, err := os.OpenFile(
		os.ExpandEnv(path),
		os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		return err
	}

	unformatted := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})

	encoder := json.NewEncoder(unformatted)
	err = encoder.Encode(object)
	if err != nil {
		return err
	}

	err = json.Indent(formatted, unformatted.Bytes(), "", "  ")
	if err != nil {
		return err
	}

	_, err = io.Copy(file, formatted)

	return err
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

func validateConfig(config *appConfig) {
	if len(config.ConsumerKey) == 0 || len(config.AccessToken) == 0 {
		log.Fatal("Authenticate first")
	}
}

func mustInitConfig() (config *appConfig) {
	config = initConfig()
	validateConfig(config)
	return config
}
