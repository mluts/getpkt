package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
)

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

func readJSON(path string, result interface{}) (err error) {
	var (
		file *os.File
	)

	path = os.ExpandEnv(path)

	file, err = os.Open(path)
	if err != nil {
		return err
	}

	io := bufio.NewReader(file)
	decoder := json.NewDecoder(io)

	for decoder.More() {
		err = decoder.Decode(result)
		if err != nil {
			return err
		}
	}

	return err
}
