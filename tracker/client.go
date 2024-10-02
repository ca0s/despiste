package tracker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
)

type TrackerClient struct {
	serverURL     string
	keepAlive     time.Duration
	clientKey     string
	clientAddress string

	httpClient *http.Client

	keepAliveURL string
}

func NewTrackerClient(clientKey string, serverURL string, clientAddress string, keepAlive time.Duration, serverName string, serverCA *x509.Certificate) *TrackerClient {
	certPool := x509.NewCertPool()
	certPool.AddCert(serverCA)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    certPool,
				ServerName: serverName,
			},
		},
	}
	return &TrackerClient{
		serverURL:     serverURL,
		keepAlive:     keepAlive,
		clientKey:     clientKey,
		clientAddress: clientAddress,

		httpClient: httpClient,

		keepAliveURL: fmt.Sprintf("%s/api/keepalive", serverURL),
	}
}

func (tc *TrackerClient) Run() {
	for {
		err := tc.SendKeepAlive()
		if err != nil {
			log.Printf("error sending keepalive: %s\n", err)
		}

		time.Sleep(tc.keepAlive)
	}
}

func (tc *TrackerClient) SendKeepAlive() error {
	request := KeepAliveRequest{
		ClientKey: tc.clientKey,
		Address:   tc.clientAddress,
	}

	encodedRequest, err := json.Marshal(&request)
	if err != nil {
		return errors.Wrap(err, "could not encode request")
	}

	response, err := tc.httpClient.Post(
		tc.keepAliveURL,
		"application/json",
		bytes.NewBuffer(encodedRequest),
	)

	if err != nil {
		return errors.Wrap(err, "could not send keepalive request")
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("keepalive response: %d\n", response.StatusCode)
		io.Copy(os.Stdout, response.Body)

	} else {
		io.Copy(io.Discard, response.Body)
	}

	defer response.Body.Close()

	return nil
}
