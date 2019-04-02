# go-clamber
A distributed system designed to crawl the internet, fronted by a d3js sitemap for visualisation.

Proposed tech stack:
- Golang
- gRPC
- Apache Kafka
- Apache Cassandra
- Google BigQuery
- D3js
- Kubernetes

This is an extension of: https://github.com/stevenayers/golang-webcrawler
## Design

### Goals
- Must be able to crawl internet infinitely, just domain based or on a fixed length
- Efficient Scaling
- Monitoring
- TDD will be used for development
- Will mitigate duplicate posts to the indexing service
- Must be able to cater to pages changing and updating the sitemap accordingly.
- All inter-service communication will be encrypted and will use gRPC

### Workflow
![app-workflow](docs/imgs/go-clamber.png)


## Components

### SiteMap UI
Sits in front of the Map API

D3JS frontend which will show your main page, and all child pages visually. Once you click on a child link, it will query the database and retrieve that map.

We should also be able to display the data as a big overview map.

This frontend is where the User will enter the URL, depth and whether or not to search external pages.

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
            "depth": 0,
            "data": {
                "links": [
                    "https://example.com/about",
                    "https://example.com/contact",
                    "https://example.com/faq",
                    "https://example.com/offices"
                ],
                "body": null
            }
        },
        {
            "URL": "https://example.com/about",
            "depth": 1,
            "timestamp": "<time>",
            "data": {
                "links": [],
                "body": null
            }
        },
        {
            "URL": "https://example.com/contact",
            "depth": 1,
            "timestamp": "<time>",
            "data": {
                "links": [],
                "body": null
            }
        },
        {
            "URL": "https://example.com/faq",
            "depth": 1,
            "timestamp": "<time>",
            "data": {
                "links": [],
                "body": null
            }
        },
        {
            "URL": "https://example.com/offices",
            "depth": 1,
            "timestamp": "<time>",
            "data": {
                "links": [],
                "body": null
            }
        }
        
  ]
  
}
```
This will specify the starting URL. Maybe it would be a good idea to put a depth limit on this and option to search just the domain or all links. The sitemap service will use the map_id to identify which map to render. This would allow us to generate maps for the entire web, or just for singular sites.

depth (int) - 0 is infinite. If you specified 10, that would be your max depth to crawl. This would decrement on every crawl, and the child links of a page would be reposted back to the queue with the new decremented crawl depth.
allow_external_links (boolean) - Enables clamber to crawl outside of the website's domain.

If data isn't found in database, API will convert this message into a protobuf and send it to kafka.

### Queue
We're going to use Kafka. This is a decision based on it's ability to scale well, and also just personal preference and a curiousity for the technology.
- Minimum 3 instance zookeeper quorom
- Parition Replication factor of 3
- Authentication
- TLS Encryption?

The Crawler service will both read and write URL messages to the queue.
### Crawler Service
Crawler service will ingest a URL protobuf message, fetch the page data, post page data and child links to the indexer, and push child links back into the queue to be crawled.

This will probably be an autoscaling worker pool on Kubernetes. It will also read and write to the queue.

#### Workflow
1. Consume URL
2. Lock shard, Check if URL exists
3. If url exists and data is not empty, stop process, if does exist, or doesn't contain page data and timestamp is more than n seconds ago, write record, then unlock and continue.
4. Query page
5. Parse page
6. Update URL record with page field to database
7. Send links back to queue

### Index Database
Cassandra seems to have good linear scaling. I also have used MongoDB and I'm not a fan. We have a strict schema here so we could go for RDBMS, but I think a distributed NoSQL database would respond better to how much this system will need to scale.

Database will store something like this (displaying it as JSON here just for my own readability):
```json
{
    "URL": "https://example.com",
    "timestamp": "<time>",
    "data": {
        "links": [],
        "body": null
    }
}
```

