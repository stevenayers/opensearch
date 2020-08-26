/*
Package app provides the opensearch crawling package.

To initiate a crawl, create a Crawler with an empty sync.WaitGroup and struct map. DbWaitGroup is needed to ensure the
opensearch process does not exit before the crawler is done writing to the database. AlreadyCrawled keeps track of the
URLs which have been crawled already in that crawl process. The rest are self explanatory.

		crawler := app.Crawler{
			DbWaitGroup: sync.WaitGroup{},
			AlreadyCrawled: make(map[string]struct{}),
			Logger: log.Logger,
			Store: app.DbStore,
		}

Create a page object with the starting URL of your crawl.

	page := &app.Page{Url: "https://golang.org"}

Call Crawl on the Crawler object, passing in your page, and the depth of the crawl you want.

	crawler.Crawl(result, 5)

Ensure your go process does not end before the crawled data has been saved to dgraph. If you need more logic to execute
first, put the line below after this, as your application will hang on Wait() until we're done writing.

	crawler.DbWaitGroup.Wait()

*/
package crawl

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"github.com/stevenayers/opensearch/pkg/config"
	"github.com/stevenayers/opensearch/pkg/database/relationship"
	"github.com/stevenayers/opensearch/pkg/logging"
	"github.com/stevenayers/opensearch/pkg/page"
	"github.com/stevenayers/opensearch/pkg/queue"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Crawler holds objects related to the crawler
type (
	Crawler struct {
		AlreadyCrawled map[string]struct{}
		sync.Mutex
		DbWaitGroup          sync.WaitGroup
		BgWaitGroup          sync.WaitGroup
		BgNotified           bool
		BgWaitNotified       bool
		Store                *relationship.Store
		BackgroundCrawlDepth int
		CrawlUid             uuid.UUID
		Queue                *queue.Queue
	}
)

func New() (c Crawler) {
	c = Crawler{
		DbWaitGroup:    sync.WaitGroup{},
		Store:          &relationship.Store{},
		CrawlUid:       uuid.New(),
		AlreadyCrawled: make(map[string]struct{}),
	}
	c.Queue = queue.NewQueue()
	c.Store.Connect()
	return
}

func (crawler *Crawler) Start() (err error) {
	for i := 1; i <= config.AppConfig.Service.NumConsumers; i++ {
		go crawler.Queue.Poll()
	}
	for msg := range crawler.Queue.ReceiveChan {
		go func(msg *sqs.Message) {
			var currentPage *page.Page
			currentPage, err = page.DeserializeSQSPage(msg)
			if err != nil {
				return
			}
			crawler.Crawl(currentPage)
		}(msg)
	}
	return
}

// Get function manages HTTP request for page
func (crawler *Crawler) Get(currentPage *page.Page) (resp *http.Response, err error) {
	var req *http.Request
	maxAttempts := config.AppConfig.Service.HttpRetryAttempts + 1
	backOffDuration := time.Duration(config.AppConfig.Service.HttpBackOffDuration) * time.Second
	client := http.Client{}
	req, err = http.NewRequest("GET", currentPage.Url, nil)
	if err != nil {
		_ = level.Error(logging.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "msg", err.Error())
		return
	}
	req.Header.Set("User-Agent", "stevenayers/opensearch")
	count := 0
	for maxAttempts > count {
		count++
		resp, err = client.Do(req)
		if err != nil {
			_ = level.Error(logging.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "msg", err.Error())
			return
		}
		switch {
		case resp.StatusCode == http.StatusOK:
			break
		case resp.StatusCode < 500:
			err = errors.New("received bad HTTP status code")
			_ = level.Debug(logging.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "statusCode", resp.StatusCode, "msg", err.Error())
			break
		default:
			if maxAttempts == count {
				err = errors.New("received bad HTTP status code")
				_ = level.Debug(logging.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "statusCode", resp.StatusCode, "msg", err.Error())
			}
			time.Sleep(backOffDuration)
		}
	}
	return
}

// Crawl function adds page to db (in a goroutine so it doesn't stop initiating other crawls), gets the child pages then
// initiates crawls for each one.
func (crawler *Crawler) Crawl(currentPage *page.Page) {
	resp, err := crawler.Get(currentPage)
	currentPage.StatusCode = http.StatusOK
	currentPage.Timestamp = time.Now().Unix()
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		currentPage.StatusCode = http.StatusNotFound
		go func(currentPage *page.Page) {
			err = crawler.Create(currentPage)
			if err != nil {
				return
			}
		}(currentPage)
		return
	}
	if err != nil {
		return
	}

	if !crawler.hasAlreadyCrawled(currentPage.Url) {
		go func(currentPage *page.Page) {
			err = crawler.Create(currentPage)
			if err != nil {
				return
			}
		}(currentPage)
	}

	if currentPage.Depth <= 0 ||
		!strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return
	}

	childPages, _ := currentPage.FetchChildPages(resp)
	for _, childPage := range childPages {
		go func(childPage *page.Page) {
			childPage.Depth = currentPage.Depth - 1
			crawler.Queue.Publish(childPage)
		}(childPage)
	}
}

// Create function checks for current page, creates if doesn't exist. Checks for parent page, creates if doesn't exist. Checks for edge
// between them, creates if doesn't exist.
func (crawler *Crawler) Create(currentPage *page.Page) (err error) {
	ctx := context.Background()
	currentUid, err := crawler.FindOrCreatePage(&ctx, currentPage)
	if err != nil {
		return
	}
	if currentPage.Parent != nil {
		var parentUid string
		parentUid, err = crawler.FindOrCreatePage(&ctx, currentPage.Parent)
		if err != nil {
			return
		}
		err = crawler.FindOrCreateLink(&ctx, parentUid, currentUid)
		if err != nil {
			return
		}
	}
	return
}

func (crawler *Crawler) FindOrCreateLink(ctx *context.Context, parentUid string, currentUid string) (err error) {
	attempts := 10
	for attempts > 0 {
		attempts--
		success, err := crawler.Store.CheckOrCreatePredicate(ctx, parentUid, currentUid)
		if err != nil {
			if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry") &&
				!strings.Contains(err.Error(), "Transaction is too old") {
				_ = level.Error(logging.Logger).Log(
					"context", "create predicate",
					"msg", err.Error(),
					"parentUid", parentUid,
					"childUid", currentUid,
				)
				break
			}
		}
		if success {
			break
		}
	}
	return
}

func (crawler *Crawler) FindOrCreatePage(ctx *context.Context, p *page.Page) (uid string, err error) {
	for uid == "" {
		uid, err = crawler.Store.FindOrCreateNode(ctx, p)
		if err != nil {
			if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry") &&
				!strings.Contains(err.Error(), "Transaction is too old") {
				_ = level.Error(logging.Logger).Log(
					"msg", err.Error(),
					"context", "create page",
					"url", p.Url,
				)
				return
			}
		}
	}
	return
}

// Locks crawl, then returns true/false dependent on Url being in map. If false, we store the Url.
func (crawler *Crawler) hasAlreadyCrawled(Url string) (isPresent bool) {
	cleanUrl := strings.TrimRight(Url, "/")
	defer crawler.Unlock()
	crawler.Lock()
	_, isPresent = crawler.AlreadyCrawled[cleanUrl]
	if !isPresent {
		crawler.AlreadyCrawled[cleanUrl] = struct{}{}
	}
	return
}
