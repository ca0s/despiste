.PHONY: all server upstream authority despiste

all: server upstream authority despiste

server:
	CGO_ENABLED=0 go build ./cmd/server

upstream:
	CGO_ENABLED=0 go build ./cmd/upstream

authority:
	CGO_ENABLED=0 go build ./cmd/authority

despiste:
	CGO_ENABLED=0 go build ./cmd/despiste