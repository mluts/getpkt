package main

import (
	"bufio"
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

func readConfig(path string) (config *appConfig, err error) {
	config = &appConfig{}
	err = readJSON(path, config)
	return config, err
}

func writeConfig(path string, config *appConfig) (err error) {
	err = writeJSON(path, config)
	log.Println("Config file written")
	return err
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
