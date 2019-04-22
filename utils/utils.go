package utils

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"sort"
)

type (
	JsonPage struct {
		Uid       string      `json:"uid,omitempty"`
		Url       string      `json:"url,omitempty"`
		Timestamp int64       `json:"timestamp,omitempty"`
		Children  []*JsonPage `json:"links,omitempty"`
	}

	JsonPredicate struct {
		Matching int `json:"matching"`
	}

	Pool struct {
		buffer chan bool
	}
)

var POOL chan bool

func InitBuffer(bufferSize int) {
	POOL = make(chan bool, bufferSize)
}

func ParseHtml(resp *http.Response) (doc *goquery.Document, err error) {

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	return
}

func ParseRelativeUrl(rootUrl string, relativeUrl string) (absoluteUrl *url.URL) {
	parsedRootUrl, err := url.Parse(rootUrl)
	if err != nil {
		return nil
	}
	absoluteUrl, err = url.Parse(parsedRootUrl.Scheme + "://" + parsedRootUrl.Host + path.Clean("/"+relativeUrl))
	if err != nil {
		return nil
	}
	absoluteUrl.Fragment = "" // Removes '#' identifiers from Url
	return
}

func IsRelativeUrl(href string) bool {
	match, _ := regexp.MatchString("^(?:[a-zA-Z]+:)?//", href)
	return !match
}

func IsRelativeHtml(href string) bool {
	htmlMatch, _ := regexp.MatchString(`(\.html$)`, href) // Doesn't cover all allowed file extensions
	if htmlMatch {
		return htmlMatch
	} else {
		match, _ := regexp.MatchString(`(\.[a-zA-Z0-9]+$)`, href)
		return !match
	}

}

func DeserializePredicate(pb []byte) (exists bool, err error) {
	jsonMap := make(map[string][]JsonPredicate)
	err = json.Unmarshal(pb, &jsonMap)
	if err != nil {
		return
	}
	edges := jsonMap["edges"]
	if len(edges) > 0 {
		exists = edges[0].Matching > 0
	} else {
		exists = false
	}
	return
}

func MaxIntSlice(v []int) int {
	sort.Ints(v)
	return v[len(v)-1]
}
