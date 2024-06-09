package scrapper

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/PFrek/gorss/internal/api"
)

type FeedItem struct {
	XMLName xml.Name `xml:"item"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link"`
	PubDate string   `xml:"pubDate"`
}

type FeedData struct {
	XMLName xml.Name   `xml:"rss"`
	Items   []FeedItem `xml:"channel>item"`
}

func fetchDataFromFeed(feedUrl string) (FeedData, error) {
	resp, err := http.Get(feedUrl)
	if err != nil {
		return FeedData{}, fmt.Errorf("Failed to fetch RSS feeed: %v", err)
	}
	defer resp.Body.Close()

	feedData, err := parseXML(resp.Body)

	return feedData, nil
}

func parseXML(xmlData io.Reader) (FeedData, error) {
	var feedData FeedData

	decoder := xml.NewDecoder(xmlData)
	err := decoder.Decode(&feedData)
	if err != nil {
		return FeedData{}, err
	}

	return feedData, nil
}

func StartScraper(config api.ApiConfig) {

}
