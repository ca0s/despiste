#!/bin/bash

rm data/certs/*.pem
./authority -action init-ca -subject despiste -ca data/certs/cafull.pem -capub data/certs/ca.pem
./authority -action cert -subject server -ca data/certs/cafull.pem -cert data/certs/server.pem
./authority -action cert -subject proxy1 -ca data/certs/cafull.pem -cert data/certs/proxy1.pem
./authority -action cert -subject client -ca data/certs/cafull.pem -cert data/certs/client.pem