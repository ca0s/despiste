package network

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"net"
)

func NewTLSListener(addr string, caCert *x509.Certificate, clientCert *x509.Certificate, clientKey *ecdsa.PrivateKey) (net.Listener, error) {
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	var cert tls.Certificate
	cert.Certificate = append(cert.Certificate, clientCert.Raw)
	cert.PrivateKey = clientKey

	return tls.Listen("tcp", addr, &tls.Config{
		MinVersion:   tls.VersionTLS13,
		RootCAs:      roots,
		ClientCAs:    roots,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
	})
}
