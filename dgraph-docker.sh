#!/usr/bin/env bash
mkdir -p ~/dgraph

# Run dgraphzero
docker run -i -p 5080:5080 -p 6080:6080 -p 8080:8080 -p 9080:9080 -p 8008:8008 -v ~/dgraph:/dgraph --name dgraph dgraph/dgraph dgraph zero

# In another terminal, now run dgraph
docker exec -i dgraph dgraph alpha --lru_mb 4096 --zero localhost:5080

# And in another, run ratel (Dgraph UI)
docker exec -i dgraph dgraph-ratel -port 8008