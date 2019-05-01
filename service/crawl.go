package service

import (
	"context"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Crawler holds objects related to the crawler
type Crawler struct {
	AlreadyCrawled map[string]struct{}
	sync.Mutex
	DbWaitGroup          sync.WaitGroup
	BgWaitGroup          sync.WaitGroup
	BgNotified           bool
	BgWaitNotified       bool
	Config               Config
	Db                   DbStore
	Logger               log.Logger
	BackgroundCrawlDepth int
	CrawlUid             uuid.UUID
	StartUrl             string
}

// Get function manages HTTP request for page
func (crawler *Crawler) Get(currentPage *Page) (resp *http.Response, err error) {
	var req *http.Request
	maxAttempts := crawler.Config.General.HttpRetryAttempts + 1
	backOffDuration := time.Duration(crawler.Config.General.HttpBackOffDuration) * time.Second
	client := http.Client{}
	req, err = http.NewRequest("GET", currentPage.Url, nil)
	if err != nil {
		_ = level.Error(crawler.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "msg", err.Error())
		return
	}
	req.Header.Set("User-Agent", "stevenayers/clamber")
	count := 0
	for maxAttempts > count {
		count++
		resp, err = client.Do(req)
		if err != nil {
			_ = level.Error(crawler.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "msg", err.Error())
			return
		}
		switch {
		case resp.StatusCode == http.StatusOK:
			break
		case resp.StatusCode < 500:
			err = errors.New("received bad HTTP status code")
			_ = level.Debug(crawler.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "statusCode", resp.StatusCode, "msg", err.Error())
			break
		default:
			if maxAttempts == count {
				err = errors.New("received bad HTTP status code")
				_ = level.Debug(crawler.Logger).Log("context", "HTTP failure", "url", currentPage.Url, "statusCode", resp.StatusCode, "msg", err.Error())
			}
			time.Sleep(backOffDuration)
		}
	}
	return
}

// Crawl function adds page to db (in a goroutine so it doesn't stop initiating other crawls), gets the child pages then
// initiates crawls for each one.
func (crawler *Crawler) Crawl(currentPage *Page) {
	resp, err := crawler.Get(currentPage)
	if err != nil {
		return
	}
	crawler.DbWaitGroup.Add(1)
	go func(currentPage *Page) {
		defer crawler.DbWaitGroup.Done()
		err := crawler.Create(currentPage)
		if err != nil {
			return
		}
	}(currentPage)
	if crawler.hasAlreadyCrawled(currentPage.Url) ||
		(currentPage.Depth == 0 && crawler.BackgroundCrawlDepth == 0) ||
		!strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return
	}
	childPages, _ := currentPage.FetchChildPages(resp, crawler.Logger)
	if currentPage.Depth > 0 {
		var childDepth int
		if currentPage.Depth == -1 {
			childDepth = -1
		} else {
			childDepth = currentPage.Depth - 1
		}
		pageWaitGroup := sync.WaitGroup{}
		childPagesChan := make(chan *Page)
		for _, childPage := range childPages {
			pageWaitGroup.Add(1)
			go func(childPage *Page) {
				defer pageWaitGroup.Done()
				childPage.Depth = childDepth
				crawler.Crawl(childPage)
				childPagesChan <- childPage
			}(childPage)
		}
		go func() {
			pageWaitGroup.Wait()
			close(childPagesChan)

		}()
		for childPages := range childPagesChan {
			currentPage.Links = append(currentPage.Links, childPages)
		}
	} else if currentPage.Depth <= 0 && crawler.BackgroundCrawlDepth != 0 {
		if !crawler.BgNotified {
			crawler.BgNotified = true
			_ = level.Info(crawler.Logger).Log(
				"url", crawler.StartUrl,
				"crawlUid", crawler.CrawlUid,
				"context", "BackgroundCrawl",
				"msg", "Initiating background crawl",
				"backgroundDepth", crawler.BackgroundCrawlDepth,
			)
		}
		detachedParent := *currentPage
		detachedParent.Parent = nil
		for _, childPage := range childPages {
			crawler.BgWaitGroup.Add(1)
			go func(childPage *Page) {
				defer crawler.BgWaitGroup.Done()
				childPage.Parent = &detachedParent
				childPage.Depth = crawler.BackgroundCrawlDepth
				crawler.BackgroundCrawl(childPage)
			}(childPage)
		}

	}
}

func (crawler *Crawler) BackgroundCrawl(currentPage *Page) {
	if !crawler.BgWaitNotified {
		crawler.BgWaitNotified = true
		go func() {
			crawler.BgWaitGroup.Wait()
			ctx := context.Background()
			txn := crawler.Db.NewTxn()
			resultDepth, err := crawler.Db.FindNodeDepth(&ctx, txn, crawler.StartUrl)
			if err != nil {
				_ = level.Error(crawler.Logger).Log(
					"url", crawler.StartUrl,
					"crawlUid", crawler.CrawlUid,
					"context", "BackgroundCrawl",
					"msg", "Could not find total depth crawled.",
				)
			}
			_ = level.Info(crawler.Logger).Log(
				"url", crawler.StartUrl,
				"crawlUid", crawler.CrawlUid,
				"context", "BackgroundCrawl",
				"crawledDepth", resultDepth,
				"msg", "Finished background Crawl",
			)
		}()
	}
	resp, err := crawler.Get(currentPage)
	if err != nil {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(currentPage *Page) {
		defer wg.Done()
		err := crawler.Create(currentPage)
		if err != nil {
			return
		}
	}(currentPage)
	if crawler.hasAlreadyCrawled(currentPage.Url) || currentPage.Depth == 0 ||
		!strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return
	}
	if currentPage.Depth != -1 {
		currentPage.Depth--
	}
	childPages, _ := currentPage.FetchChildPages(resp, crawler.Logger)
	for _, childPage := range childPages {
		crawler.BgWaitGroup.Add(1)
		go func(childPage *Page) {
			defer crawler.BgWaitGroup.Done()
			crawler.BackgroundCrawl(childPage)
		}(childPage)
	}
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

// Create function checks for current page, creates if doesn't exist. Checks for parent page, creates if doesn't exist. Checks for edge
// between them, creates if doesn't exist.
func (crawler *Crawler) Create(currentPage *Page) (err error) {
	txnUid := uuid.New().String()
	ctx := context.Background()
	var currentUid string
	if currentPage.Url != "" {
		for currentUid == "" {
			txn := crawler.Db.NewTxn()
			currentUid, err = crawler.Db.FindOrCreateNode(&ctx, txn, currentPage)
			if err != nil {
				if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
					!strings.Contains(err.Error(), "Transaction is too old") {
					_ = level.Error(crawler.Logger).Log(
						"msg", err.Error(),
						"context", "create current page",
						"url", currentPage.Url,
						"txnUid", txnUid,
					)
					return
				}
			}
		}
	}
	if currentPage.Parent != nil {
		var parentUid string
		for parentUid == "" {
			txn := crawler.Db.NewTxn()
			parentUid, err = crawler.Db.FindOrCreateNode(&ctx, txn, currentPage.Parent)
			if err != nil {
				if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
					!strings.Contains(err.Error(), "Transaction is too old") {
					_ = level.Error(crawler.Logger).Log(
						"msg", err.Error(),
						"context", "create parent page",
						"url", currentPage.Parent.Url,
						"txnUid", txnUid,
					)
					return
				}
			}
		}
		attempts := 10
		for attempts > 0 {
			attempts--
			txn := crawler.Db.NewTxn()
			success, err := crawler.Db.CheckOrCreatePredicate(&ctx, txn, parentUid, currentUid)
			if err != nil {
				if !strings.Contains(err.Error(), "Transaction has been aborted. Please retry.") &&
					!strings.Contains(err.Error(), "Transaction is too old") {
					_ = level.Error(crawler.Logger).Log(
						"context", "create predicate",
						"msg", err.Error(),
						"parentUid", parentUid,
						"childUid", currentUid,
						"txnUid", txnUid,
					)
					break
				}
			}
			if success {
				break
			}
		}
	}
	return
}
