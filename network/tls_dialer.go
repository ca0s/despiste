package network

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"net"
)

type TLSDialer struct {
	rootCAs       *x509.CertPool
	clientCert    *x509.Certificate
	clientKey     *ecdsa.PrivateKey
	tlsCertficate *tls.Certificate

	serverName string
	tlsDialer  *tls.Dialer
}

func NewTLSDialer(caCert *x509.Certificate, clientCert *x509.Certificate, clientKey *ecdsa.PrivateKey) (*TLSDialer, error) {
	var tlsCert tls.Certificate
	tlsCert.Certificate = append(tlsCert.Certificate, clientCert.Raw)
	tlsCert.PrivateKey = clientKey

	dialer := &TLSDialer{
		rootCAs:       x509.NewCertPool(),
		clientCert:    clientCert,
		clientKey:     clientKey,
		tlsCertficate: &tlsCert,
	}

	dialer.rootCAs.AddCert(caCert)

	dialer.tlsDialer = &tls.Dialer{
		Config: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			RootCAs:      dialer.rootCAs,
			Certificates: []tls.Certificate{*dialer.tlsCertficate},
		},
	}

	return dialer, nil
}

func (d *TLSDialer) Dial(network, addr string) (net.Conn, error) {
	return d.tlsDialer.Dial("tcp", addr)
}

func (d *TLSDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.tlsDialer.DialContext(ctx, "tcp", addr)
}

func (d *TLSDialer) ForServerName(name string) *TLSDialer {
	return &TLSDialer{
		rootCAs:       d.rootCAs,
		clientCert:    d.clientCert,
		clientKey:     d.clientKey,
		tlsCertficate: d.tlsCertficate,

		serverName: name,

		tlsDialer: &tls.Dialer{
			Config: &tls.Config{
				MinVersion:   tls.VersionTLS13,
				RootCAs:      d.rootCAs,
				Certificates: []tls.Certificate{*d.tlsCertficate},
				ServerName:   name,
			},
		},
	}
}
