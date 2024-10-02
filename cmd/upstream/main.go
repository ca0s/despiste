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
	flag.StringVar(&configPath, "config", "/etc/despiste/upstream.json", "despiste config path")
	flag.Parse()

	cfg, err := config.ReadConfig(configPath, false)
	if err != nil {
		log.Printf("error reading config: %s\n", err.Error())
		return
	}

	trackerClient := tracker.NewTrackerClient(
		cfg.NodeID,
		cfg.TrackerURL, cfg.NodeAddress, cfg.KeepAlive,
		cfg.TrackerID,
		cfg.CACert,
	)
	go trackerClient.Run()

	conf := socks5.Config{
		AuthMethods: []socks5.Authenticator{},
		Rules: &socks5.PermitCommand{
			EnableConnect:   true,
			EnableBind:      false,
			EnableAssociate: false,
		},
	}

	tlsListener, err := network.NewTLSListener(cfg.NodeAddress, cfg.CACert, cfg.Cert, cfg.Key)
	if err != nil {
		log.Printf("could not create tls listener: %s\n", err)
		return
	}

	server, err := socks5.New(&conf)
	if err != nil {
		log.Printf("could not create socks server: %s\n", err.Error())
		return
	}

	err = server.Serve(tlsListener)
	log.Printf("server finished: %s\n", err.Error())
}
