package scraper

import (
	"context"
	"database/sql"
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
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type FeedItem struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description *string  `xml:"description"`
	PubDate     string   `xml:"pubDate"`
}

type FeedData struct {
	XMLName xml.Name   `xml:"rss"`
	Items   []FeedItem `xml:"channel>item"`
}

type CachedFeed struct {
	Data         FeedData
	LastCachedAt time.Time
}

type CacheHitError struct{}

func (e CacheHitError) Error() string {
	return "Cache hit"
}

type Scraper struct {
	Config        api.ApiConfig
	Cache         map[string]CachedFeed
	CacheInterval time.Duration
	mux           sync.Mutex
}

func (s *Scraper) shouldFetch(feed database.Feed) bool {
	if !feed.LastFetchedAt.Valid {
		return true
	}

	cachedFeed, ok := s.Cache[feed.Url]
	if !ok {
		return true
	}

	return time.Since(cachedFeed.LastCachedAt) >= s.CacheInterval
}

func (s *Scraper) checkCache(feed database.Feed) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.shouldFetch(feed) {
		return CacheHitError{}
	}

	cachedFeed := s.Cache[feed.Url]
	cachedFeed.LastCachedAt = time.Now().UTC()
	s.Cache[feed.Url] = cachedFeed // Update the cache time first so others know it's already being cached

	return nil
}

func (s *Scraper) updateCacheData(feed database.Feed, feedData FeedData) {
	s.mux.Lock()
	defer s.mux.Unlock()

	cachedFeed := s.Cache[feed.Url]
	cachedFeed.Data = feedData
	s.Cache[feed.Url] = cachedFeed // Update the cache data when we finally get it
}

func (s *Scraper) fetchDataFromFeed(feed database.Feed) (FeedData, error) {
	err := s.checkCache(feed)
	if err != nil {
		// Cache hit
		_, markErr := s.Config.DB.MarkFeedFetched(context.Background(), feed.ID)
		if markErr != nil {
			log.Printf("Error marking feed as fetched: %v\n", markErr)
		}
		return FeedData{}, err
	}

	req, err := http.NewRequest("GET", feed.Url, nil)
	if err != nil {
		return FeedData{}, errors.New("failed to create GET request")
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

	s.updateCacheData(feed, feedData)

	_, err = s.Config.DB.MarkFeedFetched(context.Background(), feed.ID)
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

func (s *Scraper) Start(interval time.Duration, numFeeds int) chan bool {
	ticker := time.NewTicker(interval)
	done := make(chan bool)

	log.Printf("Starting Scraper with interval %v and limit %d", interval, numFeeds)
	go func() {
		for {
			select {
			case <-done:
				return

			case <-ticker.C:
				s.scrape(numFeeds)
			}
		}
	}()

	return done
}

func (s *Scraper) scrape(numFeeds int) {
	log.Println("Finding feeds in need of fetching...")
	feedsToFetch, err := s.Config.DB.GetNextFeedsToFetch(context.Background(), int32(numFeeds))
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
			feedData, err := s.fetchDataFromFeed(feed)
			if err != nil {
				if errors.Is(err, CacheHitError{}) {
					log.Println(err)
					return
				}

				log.Printf("Error fetching %s: %v\n", feed.Url, err)
				return
			}

			s.processFeed(feedData, feed.ID)
		}(feed)
	}

	wg.Wait()
	log.Println("Finished processing feeds. Waiting for next cycle...")
}

func parsePubDate(pubDate string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
	}

	var converted *time.Time
	for _, format := range formats {
		result, err := time.Parse(format, pubDate)
		if err == nil {
			converted = new(time.Time)
			*converted = result
			break
		}
	}

	if converted == nil {
		return time.Time{}, fmt.Errorf("Failed to convert date from string:, %s", pubDate)
	}

	return converted.UTC(), nil
}

func (s *Scraper) processFeed(data FeedData, feedID uuid.UUID) {
	fmt.Printf("Found %d entries\n", len(data.Items))

	for _, item := range data.Items {
		currentTime := time.Now().UTC()

		pubDate, err := parsePubDate(item.PubDate)
		if err != nil {
			log.Println(err)
			log.Println("Skipping Post:", item.Title)
			continue
		}

		description := sql.NullString{
			String: "",
			Valid:  false,
		}

		if item.Description != nil {
			description.String = *item.Description
			description.Valid = true
		}

		_, err = s.Config.DB.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   currentTime,
			UpdatedAt:   currentTime,
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: pubDate,
			FeedID:      feedID,
		})
		if err != nil {
			if err, ok := err.(*pq.Error); ok {
				if err.Code.Name() == "unique_violation" {
					log.Println("Post with URL already in DB, skipping.")
					continue
				}
			}

			log.Println("Failed to save Post to DB:", item.Title)
			continue
		}

		log.Println("Saved Post to DB:", item.Title)
	}
}
