package main

import "github.com/ca0s/despiste/tracker"

type StaticUpstreamProvider struct {
	addr     string
	upstream *tracker.Upstream
}

func NewStaticUpstreamProvider(addr string, name string) *StaticUpstreamProvider {
	return &StaticUpstreamProvider{
		addr: addr,
		upstream: &tracker.Upstream{
			Address:   addr,
			Key:       name,
			Enabled:   true,
			Available: true,
		},
	}
}

func (p *StaticUpstreamProvider) GetUpstream() (*tracker.Upstream, error) {
	return p.upstream, nil
}
