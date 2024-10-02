package main

import (
	"encoding/json"
	"os"
)

func ReadUpstreamKeyFile(path string) ([]string, error) {
	var keys []string

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(fd).Decode(&keys)
	if err != nil {
		return nil, err
	}

	return keys, nil
}
