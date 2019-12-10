package page

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/clamber/pkg/logging"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type (

	// Page holds page data
	Page struct {
		Uid        string  `json:"-"`
		Url        string  `json:"url,omitempty"`
		Links      []*Page `json:"links,omitempty"`
		Parent     *Page   `json:"-"`
		Depth      int     `json:"-"`
		Timestamp  int64   `json:"timestamp,omitempty"`
		StartUrl   string  `json:"-"`
		StatusCode int     `json:"status_code,omitempty"`
	}

	// JsonPage is used to turn Page into a dgraph compatible struct
	JsonPage struct {
		Uid        string      `json:"uid,omitempty"`
		Url        string      `json:"url,omitempty"`
		Depth      int         `json:"depth,omitempty"`
		Timestamp  int64       `json:"timestamp,omitempty"`
		Children   []*JsonPage `json:"links,omitempty"`
		StatusCode int         `json:"status_code,omitempty"`
	}

	JsonResult struct {
		Result []*JsonPage `json:"result,omitempty"`
	}
	// Page holds page data
	SQSPage struct {
		Url       string   `json:"url,omitempty"`
		Parent    *SQSPage `json:"parent,omitempty"`
		Depth     int      `json:"depth,omitempty"`
		Timestamp int64    `json:"timestamp,omitempty"`
		StartUrl  string   `json:"start_url,omitempty"`
	}

	// JsonPredicate is used to hold the Predicate result from dgraph
	JsonPredicate struct {
		Matching int `json:"matching"`
	}
)

// FetchChildPages function converts http response into child page objects
func (page *Page) FetchChildPages(resp *http.Response) (childPages []*Page, err error) {
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		_ = level.Error(logging.Logger).Log("context", "failed to parse HTML", "url", page.Url, "msg", err.Error())
		return
	}
	defer resp.Body.Close()
	localProcessed := make(map[string]struct{})
	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && page.IsRelativeUrl(href) && page.IsRelativeHtml(href) && href != "" {
			absoluteUrl, _ := page.ParseRelativeUrl(href)
			_, isPresent := localProcessed[absoluteUrl.Path]
			if !isPresent {
				localProcessed[absoluteUrl.Path] = struct{}{}
				childPage := Page{
					Url:       strings.TrimRight(absoluteUrl.String(), "/"),
					Parent:    page,
					StartUrl:  page.StartUrl,
					Timestamp: time.Now().Unix(),
				}
				childPages = append(childPages, &childPage)
			}
		}
	})
	return
}

// MaxDepth function gets the max depth of the recursive page structure
func (page *Page) MaxDepth() (countDepth int) {
	if page.Links != nil {
		var childDepths []int
		for _, childPage := range page.Links {
			childDepths = append(childDepths, childPage.MaxDepth())
		}
		countDepth = maxIntSlice(childDepths) + 1
	}
	return
}

// ParseRelativeUrl function parses a relative URL string into a URL object
func (page *Page) ParseRelativeUrl(relativeUrl string) (absoluteUrl *url.URL, err error) {
	parsedRootUrl, err := url.Parse(page.Url)
	if err != nil {
		return nil, err
	}
	absoluteUrl, err = url.Parse(parsedRootUrl.Scheme + "://" + parsedRootUrl.Host + path.Clean("/"+relativeUrl))
	if err != nil {
		return nil, err
	}
	absoluteUrl.Fragment = "" // Removes '#' identifiers from Url
	return
}

// IsRelativeUrl function checks for relative URL path
func (page *Page) IsRelativeUrl(href string) bool {
	match, _ := regexp.MatchString("^(?:[a-zA-Z]+:)?//", href)
	return !match
}

// IsRelativeHtml function checks to see if relative URL points to a HTML file
func (page *Page) IsRelativeHtml(href string) bool {
	htmlMatch, _ := regexp.MatchString(`(\.html$)`, href) // Doesn't cover all allowed file extensions
	if htmlMatch {
		return htmlMatch
	}
	match, _ := regexp.MatchString(`(\.[a-zA-Z0-9]+$)`, href)
	return !match

}

// Converts JSONPage into a Page
func convertJsonPageToPage(parentPage *Page, jsonPage *JsonPage) (currentPage *Page) {
	currentPage = &Page{
		Uid:       jsonPage.Uid,
		Url:       jsonPage.Url,
		Timestamp: jsonPage.Timestamp,
	}
	if parentPage != nil {
		currentPage.Parent = parentPage
	}
	wg := sync.WaitGroup{}
	convertPagesChan := make(chan *Page)
	for _, childJsonPage := range jsonPage.Children {
		wg.Add(1)
		go func(childJsonPage *JsonPage) {
			defer wg.Done()
			childPage := convertJsonPageToPage(currentPage, childJsonPage)
			convertPagesChan <- childPage
		}(childJsonPage)
	}
	go func() {
		wg.Wait()
		close(convertPagesChan)

	}()
	for childPages := range convertPagesChan {
		currentPage.Links = append(currentPage.Links, childPages)
	}
	return
}

// Converts a Page to a JSONPage
func convertPageToJsonPage(currentPage *Page) (jsonPage JsonPage) {
	return JsonPage{
		Uid:        currentPage.Uid,
		Url:        currentPage.Url,
		Timestamp:  currentPage.Timestamp,
		StatusCode: currentPage.StatusCode,
	}
}

// Converts SQSPage into a Page
func convertSOSPageToPage(sqsPage *SQSPage) *Page {
	return &Page{
		Url:      sqsPage.Url,
		Depth:    sqsPage.Depth,
		StartUrl: sqsPage.StartUrl,
	}
}

// Converts a Page to a SQSPage
func ConvertPageToSQSPage(currentPage *Page) *SQSPage {
	return &SQSPage{
		Url:      currentPage.Url,
		Depth:    currentPage.Depth,
		StartUrl: currentPage.StartUrl,
	}
}

// Turns a Page into a JSON string
func SerializeJsonPage(currentPage *Page) (pb []byte, err error) {
	p := convertPageToJsonPage(currentPage)
	pb, err = json.Marshal(p)
	return
}

// Turns JSON dgraph result into a Page
func DeserializeJsonPage(pb []byte) (currentPage *Page, err error) {
	var jsonPages JsonResult
	err = json.Unmarshal(pb, &jsonPages)
	if len(jsonPages.Result) > 0 {
		currentPage = convertJsonPageToPage(nil, jsonPages.Result[0])
	}
	return
}

func DeserializeSQSPage(msg *sqs.Message) (currentPage *Page, err error) {
	var sqsPage SQSPage
	err = json.Unmarshal([]byte(*msg.Body), &sqsPage)
	if err != nil {
		_ = level.Info(logging.Logger).Log("msg", "Error converting payload to page", "error", err.Error())
	}
	if &sqsPage != nil {
		currentPage = convertSOSPageToPage(&sqsPage)
		if sqsPage.Parent != nil {
			currentPage.Parent = convertSOSPageToPage(sqsPage.Parent)
		}
	} else {
		_ = level.Info(logging.Logger).Log("msg", "Error converting payload to page", "message", msg)
	}
	return
}

// Checks JSON dgraph edge result to see if edge exists
func DeserializePredicate(pb []byte) (exists bool, err error) {
	var jsonPredicate JsonPredicate
	err = json.Unmarshal(pb, &jsonPredicate)
	exists = jsonPredicate.Matching > 0
	return
}

// Return max int in slice
func maxIntSlice(v []int) int {
	sort.Ints(v)
	return v[len(v)-1]
}
