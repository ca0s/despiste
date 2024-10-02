package main

import (
	"flag"
	"log"

	"github.com/armon/go-socks5"
	"github.com/ca0s/despiste/certificates"
	"github.com/ca0s/despiste/network"
)

func main() {
	var (
		serverAddress string
		proxyAddress  string

		serverID string

		certFile string
		caFile   string
	)

	flag.StringVar(&serverAddress, "server-address", "", "despiste server address:port")
	flag.StringVar(&proxyAddress, "listen-address", "127.0.0.1:1080", "Listen address for the local socks5 server")
	flag.StringVar(&serverID, "server-id", "server", "Server name, as defined by its TLS certificate")
	flag.StringVar(&certFile, "cert", "data/certs/client.pem", "Certificate crt+key PEM file location")
	flag.StringVar(&caFile, "ca", "data/certs/ca.pem", "CA crt PEM file location")

	flag.Parse()

	if serverAddress == "" {
		log.Printf("server-address is mandatory\n")
		return
	}

	caCert, _, err := certificates.ReadCert(caFile, false)
	if err != nil {
		log.Printf("could not read CA certificate: %s\n", err)
		return
	}

	clientCert, clientKey, err := certificates.ReadCert(certFile, true)
	if err != nil {
		log.Printf("could not read client certificate: %s\n", err)
		return
	}

	staticUpstreamProvider := NewStaticUpstreamProvider(serverAddress, serverID)

	upstreamDialer, err := network.NewUpstreamDialer(staticUpstreamProvider, caCert, clientCert, clientKey)
	if err != nil {
		log.Printf("could not create upstream dialer: %s\n", err.Error())
		return
	}

	conf := socks5.Config{
		Rules: &socks5.PermitCommand{
			EnableConnect:   true,
			EnableBind:      false,
			EnableAssociate: false,
		},
		Dial: upstreamDialer.Dial,
	}

	server, err := socks5.New(&conf)
	if err != nil {
		panic(err)
	}

	log.Printf("starting socks5 server at %s\n", proxyAddress)

	if err := server.ListenAndServe("tcp", proxyAddress); err != nil {
		panic(err)
	}
}
