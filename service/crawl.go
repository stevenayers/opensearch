package service

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"sync"
)

// Crawler holds objects related to the crawler
type Crawler struct {
	AlreadyCrawled map[string]struct{}
	sync.Mutex
	DbWaitGroup sync.WaitGroup
	Config      Config
	Db          DbStore
	Logger      log.Logger
}

// Crawl function adds page to db (in a goroutine so it doesn't stop initiating other crawls), gets the child pages then
// initiates crawls for each one.
func (crawler *Crawler) Crawl(currentPage *Page, depth int) {
	//client := http.Client{
	//	Timeout: service.AppConfig.General.HttpTimeout,
	//}
	resp, err := http.Get(currentPage.Url)
	if err != nil {
		_ = level.Error(crawler.Logger).Log("context", "failed to get URL", "url", currentPage.Url, "msg", err.Error())
		return
	}
	if resp.StatusCode != 200 {
		_ = level.Debug(crawler.Logger).Log("context", "HTTP Failure", "url", currentPage.Url, "statusCode", resp.StatusCode)
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
	if crawler.hasAlreadyCrawled(currentPage.Url) || depth <= 0 ||
		!strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		return
	}
	pageWaitGroup := sync.WaitGroup{}
	childPagesChan := make(chan *Page)
	childPages, _ := currentPage.FetchChildPages(resp, crawler.Logger)
	for _, childPage := range childPages {
		pageWaitGroup.Add(1)
		go func(childPage *Page) {
			defer pageWaitGroup.Done()
			crawler.Crawl(childPage, depth-1)
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
				continue
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
				continue
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
				continue
			}
			if success {
				break
			}
		}
	}
	return
}
