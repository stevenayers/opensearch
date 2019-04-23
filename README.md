# clamber

[![Go Report Card](https://goreportcard.com/badge/github.com/stevenayers/clamber)](https://goreportcard.com/report/github.com/stevenayers/clamber)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/stevenayers/clamber)

A distributed system designed to crawl the internet.

Proposed tech stack:
- Design as a monolith first, move to distributed where/when necessary.
- Golang
- HTTP/REST
- JSON
- [Dgraph](https://dgraph.io)


## Getting Started
Warning: Expect performance issues when running clamber and dgraph locally, avoid running a depth higher than 3.


1. Clone project
    ```bash
    git clone git@github.com:stevenayers/clamber.git
    ```
1. Install dependencies & build binary (`cd` to project directory)
    ```bash
    dep ensure
    go build clamber.go
    ```
1. Run dgraph (if you don't have an existing instance already).
    ```bash
    mkdir -p ~/dgraph
    
    # Run dgraphzero
    docker run -i -p 5080:5080 -p 6080:6080 -p 8080:8080 -p 9080:9080 -p 8008:8008 -v ~/dgraph:/dgraph --name dgraph dgraph/dgraph dgraph zero
    
    # In another terminal, now run dgraph
    docker exec -i dgraph dgraph alpha --lru_mb 4096 --zero localhost:5080
    
    # And in another, run ratel (Dgraph UI)
    docker exec -i dgraph dgraph-ratel -port 8008
    ```
1. Run clamber
    ```bash
    ./clamber -config ./Config.toml
    ```
1. You're good to go.
    Example query url: [http://localhost:8000/search?url=https://golang.org&depth=3](http://localhost:8000/search?url=https://golang.org&depth=3)

## Design

### Goals
- Must be able to crawl internet infinitely, just domain based or on a fixed length
- TDD will be used for development
- Must be able to cater to pages changing and updating the sitemap accordingly.

## Endpoints

### Search
Takes a URL, depth, allow_external_links, checks Page Database to see if we already have the info. If we do, query and return it. If not, initiate recursive crawl.

| Parameter            | Type   | Description |
|----------------------|--------|-------------|
| url                  | string | starting url for sitemap |
| depth                | int    | 0 is infinite. If you specified 10, that would be your max depth to crawl. |
| display_depth        | int    | how deep a depth to return in JSON (Not yet implemented) |
| allow_external_links | bool   | whether to crawl external links or not (Not yet implemented) |


`/search` will take the following query parameters:
```json
{
    "url": "https://example.com",
    "depth": 0,
    "display_depth": 10,
    "allow_external_links": false
}
```
Sample result:
```json
{
    "query": {
      "url": "https://example.com",
      "depth": 1, 
      "display_depth": 10,
      "allow_external_links": false
    },
    "status": {
      "message": "5 pages found at a depth of 1.",
      "code": "200"
    },
    "results": [
        {
            "URL": "https://example.com",
            "timestamp": "<time>",
            "links": [
                {
                    "URL": "https://example.com/about",
                    "timestamp": "<time>",
                    "links": []
                },
                {
                    "URL": "https://example.com/contact",
                    "timestamp": "<time>",
                    "links": []
                },
                {
                    "URL": "https://example.com/faq",
                    "timestamp": "<time>",
                    "links": []
                },
                {
                    "URL": "https://example.com/offices",
                    "timestamp": "<time>",
                    "links": []
                }
            ]
        }
    ]
}
```





