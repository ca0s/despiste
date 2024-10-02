package main

import (
	"flag"
	"log"

	"github.com/armon/go-socks5"
	"github.com/ca0s/despiste/config"
	"github.com/ca0s/despiste/network"
	"github.com/ca0s/despiste/tracker"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "/etc/despiste/server.json", "despiste config path")
	flag.Parse()

	cfg, err := config.ReadConfig(configPath, true)
	if err != nil {
		log.Printf("error reading config: %s\n", err.Error())
		return
	}

	log.Printf("starting tracker API server at %s\n", cfg.TrackerAddress)
	trackerServer := tracker.NewTrackerServer(cfg.TrackerAddress, cfg.UpstreamKeys, cfg.UpstreamDeadline, cfg.CertFile)
	go trackerServer.Run()

	upstreamSelector, err := network.NewUpstreamDialer(trackerServer, cfg.CACert, cfg.Cert, cfg.Key)
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
		Dial: upstreamSelector.Dial,
	}

	server, err := socks5.New(&conf)
	if err != nil {
		panic(err)
	}

	tlsListener, err := network.NewTLSListener(cfg.NodeAddress, cfg.CACert, cfg.Cert, cfg.Key)
	if err != nil {
		log.Printf("could not start tls listener at %s: %s\n", cfg.NodeAddress, err.Error())
		return
	}

	log.Printf("starting socks5 server at %s\n", cfg.NodeAddress)

	if err := server.Serve(tlsListener); err != nil {
		panic(err)
	}
}
