# despiste

despiste is a single inbound / multiple outbound socks 5 proxy server. It has the following components:

- server: which serves as the primary socks5 proxy for users and tracks all upstreams
- upstream: an outbound node which is used by the server
- despiste: a local socks5 proxy which encapsulates user's traffic in TLS1.3 and authenticates against the server
- authority: a simple utility to manage the PKI used by all components

## Build

```
$ make
```

## Deploy

Create a CA

```
$ ./authority -action init-ca
```

Create the server certificate
```
$ ./authority -action cert -subject server
```

Create X outbound nodes certificates
```
$ ./authority -action cert -subject upstream-X
$ ./authority -action cert -subject upstream-Y
```

Create a certificate for the client
```
$ ./authority -action cert -subject client-x
```

Upload the ```server``` binary, ```data/server.json```, and the ```data/certs/server.pem``` certificate to your inbound node.

Upload the ```upstream``` binary, ```data/upstream.json``` and the appropriate ```data/certs/upstream-X.pem```certificate to your outbound nodes.

Upload the ```despiste``` binary and its ```data/certs/client-x.pem``` certificate to wherever you want to access the proxy network from.

Upload ```data/certs/ca.pem``` to all nodes. __do not upload cafull.pem anywhere__, it contains your CA's private key!

Lets say that you have the following nodes:

- A server located at 1.1.1.1, called ```server```
- An upstream located at 2.2.2.2, called ```upstream-X```
- Another upstream located at 3.3.3.3, called ```upstream-Y```

```server.json``` would be:

```json
{
	"ca": "ca.pem",
	"cert": "server.pem",

	"node_address": "1.1.1.1:51080",
	"tracker_address": "1.1.1.1:8000",

	"upstreams": ["upstream-X", "upstream-Y"]
}
```

And ```upstream.json``` for ```upstream-X```

```json
{
	"ca": "ca.pem",
	"cert": "upstream-X.pem",

	"node_address": "2.2.2.2:41080",
	"tracker_id": "server",
	"tracker_url": "https://1.1.1.1:8000"
}

```

Execute the following commands:

server:
```
$ ./server -config server.json
2021/12/01 13:22:37 starting tracker API server at :8000
2021/12/01 13:22:37 starting socks5 server at :51080
```

upstream:
```
$ ./upstream -config upstream.json
```

client (your box, probably):
```
$ ./despiste -server-address 1.1.1.1:51080 -server-name server
```

And test it is working:
```
$ proxychains curl ip.ka0labs.net
```

You should see the IP of the upstream server.

# How it works

This is a proxy network which accepts client connections on one single inbound node and balances outbound connections via a set of upstreams. Connections between all nodes are secured by TLS1.3 and a locally generated CA.

The inbound server exposes the following services:
- tracker
- socks5

The tracker keeps record of the available upstream nodes. Upstreams periodically send a keep-alive HTTP message to this service, which tracks the last time it was seen and their last IP address.

The socks5 server accepts TLS connections from clients, demanding client cert authentication. Once a connection is accepted, the server chooses an upstream server from those available and forwards the socks connection.

All nodes have their own certificate, which you can generate with the ```authority``` binary. Each certificate must have a different subject, which needs to be added to the server's ```upstreams``` config key.



