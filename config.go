package main

import (
	"bufio"
	"encoding/json"
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
	file, err := os.OpenFile(
		os.ExpandEnv(path),
		os.O_WRONLY|os.O_CREATE,
		0600,
	)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(config)
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
