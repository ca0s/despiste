package main

import (
	"encoding/json"
	"os"

	"github.com/armon/go-socks5"
)

func ReadCredentialsFile(path string) (socks5.StaticCredentials, error) {
	credentials := make(map[string]string)

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(fd).Decode(&credentials)
	if err != nil {
		return nil, err
	}

	return credentials, nil
}
