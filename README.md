# clamber
A distributed system designed to crawl the internet.

Proposed tech stack:
- Design as a monolith first, move to distributed where/when necessary.
- Golang
- HTTP/REST
- JSON
- ~RDBMS~
- [Dgraph](https://dgraph.io)

This is an extension of: https://github.com/stevenayers/golang-webcrawler


## Getting Started
Warning: Expect performance issues when running clamber and dgraph locally, avoid running a depth higher than 3.


1. Clone project
    ```bash
    git clone git@github.com:stevenayers/clamber.git
    ```
1. Install dependencies & build binary (`cd` to project directory)
    ```bash
    dep ensure && go build clamber.go
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
1. You're good to go. Example query url: [http://localhost:8000/search?url=https://golang.org&depth=3](http://localhost:8000/search?url=https://golang.org&depth=3)

## Design

### Roadmap (order of priority)
- Fix and close all issues [1/3]
- Update design documentation
- ~add getting started section.~
- Improve testing of handlers
- Check safety of recursive functions
- Separate out http client into a different package

### Goals
- Must be able to crawl internet infinitely, just domain based or on a fixed length
- TDD will be used for development
- Must be able to cater to pages changing and updating the sitemap accordingly.

### Workflow (out of date)
![app-workflow](docs/imgs/go-clamber-simple.png)


## Components

### Map API
1. Takes a URL, depth, allow_external_links, checks Page Database to see if we already have the info. If we do, query and return it.
2. If not, initiate recursive crawl.

`/search` will take the following query parameters:
```json
{
    "url": "https://example.com",
    "depth": 0,
    "display_depth": 10,
    "allow_external_links": false
}
```
API will return
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
This will specify the starting URL. Maybe it would be a good idea to put a depth limit on this and option to search just the domain or all links. The sitemap service will use the map_id to identify which map to render. This would allow us to generate maps for the entire web, or just for singular sites.

depth (int) - 0 is infinite. If you specified 10, that would be your max depth to crawl. This would decrement on every crawl, and the child links of a page would be reposted back to the queue with the new decremented crawl depth.
allow_external_links (boolean) - Enables clamber to crawl outside of the website's domain.


#### Workflow
1. Check if URL exists
3. If url exists and data is not empty, stop process, if doesn't exist or timestamp is more than n seconds ago, continue.
4. Query page
5. Parse page
6. Update URL record with page field to database
7. Spawn new goroutine for child links



## Data Model

### Mutations
* Create single node with no predicates to existing nodes.
* Create singular node with a predicate to the parent node in that crawl sequence.
* Create predicate between two existing nodes
* Mutations will always be made with predicates facing upstream (eg. childPage -> parentPage)

### Queries
* Retrieve node by URL
* Retrieve node by Uid
* Recursive retrieval (reverse predicate calls)

### Questions
* What happens if the creation of a node fails?
* How do you manage predicates to a failed parent node?
