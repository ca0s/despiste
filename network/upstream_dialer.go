package network

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"net"

	"github.com/ca0s/despiste/tracker"
	"golang.org/x/net/proxy"
)

type UpstreamDialer struct {
	upstreamProvider UpstreamProvider
	tlsDialer        *TLSDialer
}

type UpstreamProvider interface {
	GetUpstream() (*tracker.Upstream, error)
}

func NewUpstreamDialer(provider UpstreamProvider, caCert *x509.Certificate, clientCert *x509.Certificate, clientKey *ecdsa.PrivateKey) (*UpstreamDialer, error) {
	tlsDialer, err := NewTLSDialer(caCert, clientCert, clientKey)
	if err != nil {
		return nil, err
	}

	return &UpstreamDialer{
		upstreamProvider: provider,
		tlsDialer:        tlsDialer,
	}, nil
}

func (us *UpstreamDialer) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	upstream, err := us.upstreamProvider.GetUpstream()
	if err != nil {
		return nil, err
	}

	dialSocksProxy, err := proxy.SOCKS5(network, upstream.Address, nil, us.tlsDialer.ForServerName(upstream.Key))
	if err != nil {
		return nil, err
	}

	ctxDialer := (dialSocksProxy).(interface {
		DialContext(ctx context.Context, network, addr string) (net.Conn, error)
	})

	return ctxDialer.DialContext(ctx, network, addr)
}
