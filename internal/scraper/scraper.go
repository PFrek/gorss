package scraper

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/PFrek/gorss/internal/api"
	"github.com/PFrek/gorss/internal/database"
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

type Scraper struct {
	Config        api.ApiConfig
	Cache         map[string]FeedData
	CacheInterval time.Duration
}

func (scraper Scraper) shouldFetch(feed database.Feed) bool {
	if !feed.LastFetchedAt.Valid {
		return true
	}

	return time.Since(feed.LastFetchedAt.Time) >= scraper.CacheInterval
}

func (scraper *Scraper) fetchDataFromFeed(feed database.Feed) (FeedData, error) {
	req, err := http.NewRequest("GET", feed.Url, nil)
	if err != nil {
		return FeedData{}, errors.New("failed to create GET request")
	}

	if match, ok := scraper.Cache[feed.Url]; ok {
		if !scraper.shouldFetch(feed) {
			// Feed is cached and shouldn't fetch yet
			log.Println("Cache hit")
			return match, nil

		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return FeedData{}, fmt.Errorf("failed HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return FeedData{}, fmt.Errorf("failed HTTP request: status %v", resp.Status)
	} else {
		log.Println("Response:", resp.Status)
	}

	feedData, err := parseXML(resp.Body)

	scraper.Cache[feed.Url] = feedData

	_, err = scraper.Config.DB.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Error marking feed as fetched: %v\n", err)
	}
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

func (scraper *Scraper) Start(interval time.Duration, numFeeds int) chan bool {
	ticker := time.NewTicker(interval)
	done := make(chan bool)

	log.Printf("Starting Scraper with interval %v and limit %d", interval, numFeeds)
	go func() {
		for {
			select {
			case <-done:
				return

			case <-ticker.C:
				scraper.scrape(numFeeds)
			}
		}
	}()

	return done
}

func (scraper *Scraper) scrape(numFeeds int) {
	log.Println("Finding feeds in need of fetching...")
	feedsToFetch, err := scraper.Config.DB.GetNextFeedsToFetch(context.Background(), int32(numFeeds))
	if err != nil {
		log.Printf("Error getting feeds to fetch: %v", err)
		return
	}

	if len(feedsToFetch) == 0 {
		log.Println("No feeds in need of fetching. Waiting for next cycle...")
		return
	}

	var wg sync.WaitGroup

	for _, feed := range feedsToFetch {
		wg.Add(1)

		go func(feed database.Feed) {
			defer wg.Done()

			log.Println("Fetching feed from ", feed.Url)
			feedData, err := scraper.fetchDataFromFeed(feed)
			if err != nil {
				log.Printf("Error fetching %s: %v\n", feed.Url, err)
				return
			}

			processFeed(feedData)
		}(feed)
	}

	wg.Wait()
	log.Println("Finished processing feeds. Waiting for next cycle...")
}

func processFeed(data FeedData) {
	// for i, entry := range data.Items {
	// 	fmt.Printf("%d > %s\n", i, entry.Title)
	// }
	fmt.Printf("Found %d entries\n", len(data.Items))
}
