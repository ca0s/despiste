package config

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/ca0s/despiste/certificates"
)

type Config struct {
	// common fields
	CAFile   string `json:"ca"`
	CertFile string `json:"cert"`

	CACert *x509.Certificate `json:"-"`
	Cert   *x509.Certificate `json:"-"`
	Key    *ecdsa.PrivateKey `json:"-"`

	NodeID      string `json:"-"`
	NodeAddress string `json:"node_address"`

	// server fields
	TrackerAddress   string        `json:"tracker_address"`
	UpstreamKeys     []string      `json:"upstreams"`
	UpstreamDeadline time.Duration `json:"upstream_deadline"`

	// upstream fields
	KeepAlive  time.Duration `json:"keepalive"`
	TrackerID  string        `json:"tracker_id"`
	TrackerURL string        `json:"tracker_url"`
}

func ReadConfig(path string, isServer bool) (*Config, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	cfg := Config{
		CAFile:           "/etc/despiste/ca.pem",
		CertFile:         "/etc/despiste/cert.pem",
		UpstreamDeadline: time.Minute,
		KeepAlive:        30 * time.Second,
	}

	err = json.NewDecoder(fd).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.NodeAddress == "" {
		return nil, errors.New("node_address cannot be empty")
	}

	if isServer {
		if cfg.TrackerAddress == "" {
			return nil, errors.New("tracker_address cannot be empty")
		}
	}

	if !isServer {
		if cfg.TrackerURL == "" {
			return nil, errors.New("tracker_url cannot be empty")
		}

		if cfg.TrackerID == "" {
			return nil, errors.New("tracker_id cannot be empty")
		}
	}

	caCert, _, err := certificates.ReadCert(cfg.CAFile, false)
	if err != nil {
		return nil, err
	}

	cfg.CACert = caCert

	cert, key, err := certificates.ReadCert(cfg.CertFile, true)
	if err != nil {
		return nil, err
	}
	cfg.Cert = cert
	cfg.Key = key

	cfg.NodeID = cert.Subject.CommonName

	return &cfg, nil
}
