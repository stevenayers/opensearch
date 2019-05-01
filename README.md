# clamber
[![Build Status](https://travis-ci.org/stevenayers/clamber.svg?branch=master)](https://travis-ci.org/stevenayers/clamber)
[![codecov.io Code Coverage](https://img.shields.io/codecov/c/github/stevenayers/clamber.svg)](https://codecov.io/github/stevenayers/clamber?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/stevenayers/clamber)](https://goreportcard.com/report/github.com/stevenayers/clamber)
[![Release](https://img.shields.io/badge/release-v0.1--alpha-5272B4.svg)](https://github.com/stevenayers/clamber/releases/tag/v0.1-alpha)
[![GoDoc](https://godoc.org/github.com/stevenayers/clamber?status.svg)](https://godoc.org/github.com/stevenayers/clamber)

Proposed tech stack:
- Design as a monolith first, move to distributed where/when necessary.
- Golang
- HTTP/REST
- JSON
- [Dgraph](https://dgraph.io)

## Release Notes
This release is the last monolith release of the API. Adding in display depth and infinite crawling while still returning
a response in a reasonable time frame is not possible when the node is also having to run background workloads.

Next steps will be to improve monitoring and figure out the best distributed design to move to.

## Getting Started
Warning: Expect performance issues when running clamber and dgraph locally, avoid running a depth higher than 3.


1. Clone project and build binary
    ```bash
    git clone git@github.com:stevenayers/clamber.git
    cd clamber
    dep ensure
    go build cmd/clamber.go
    ```
1. Run dgraph (if you don't have an existing instance already).
    ```bash
    mkdir -p ~/dgraph
    
    docker-compose -f docs/docker-compose.yaml up
    ```
1. Run clamber
    ```bash
    ./clamber -config cmd/Config.toml
    ```
1. You're good to go.
    ```bash
    curl -s 'http://localhost:8000/search?url=https://golang.org&depth=2'
    ```
 
## Endpoints

### Search
Takes a URL, depth, allow_external_links, checks Page Database to see if we already have the info. If we do, query and return it. If not, initiate recursive crawl.

`/search` will take the following query parameters:

| Parameter            | Type   | Description |
|----------------------|--------|-------------|
| url                  | string | starting url for sitemap |
| depth                | int    | If you specified 10, that would be your max depth to crawl. |
| display_depth        | int    | how deep a depth to return in JSON (Not yet implemented) |
| allow_external_links | bool   | whether to crawl external links or not (Not yet implemented) |


Sample response:
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





