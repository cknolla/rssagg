package main

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"log"
	"rssagg/internal/database"
	"strings"
	"sync"
	"time"
)

func weBeScrapin(db *database.Queries, concurrency int32, timeBetweenRequest time.Duration) {
	log.Printf("we be scrapin with %d goroutines every %v duration\n", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), concurrency)
		if err != nil {
			log.Println("error while fetching feeds", err)
			continue
		}
		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeed(db, feed, wg)
		}
		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, feed database.Feed, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("error while marking feed", err)
		return
	}
	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Println("error while fetching feed url", err)
		return
	}
	for _, item := range rssFeed.Channel.Item {
		description := sql.NullString{
			String: item.Description,
		}
		if item.Description != "" {
			description.Valid = true
		}
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Println("error while parsing pub date", item.PubDate, err)
			continue
		}
		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: description,
			PublishedAt: pubDate,
			Url:         item.Link,
			FeedID:      feed.ID,
		})
		if err != nil {
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				log.Println("error while creating post", err)
			}
			continue
		}
	}
	log.Printf("Feed %s collected, %d posts found", feed.Name, len(rssFeed.Channel.Item))
}
