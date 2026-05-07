package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rkressm/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func scrapeFeeds(s *state) {
	ctx := context.Background()
	feedToFetch, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		log.Println("something went wrong:", err)
		return
	}
	_, err = s.db.MarkFeedFetched(ctx, feedToFetch.ID)
	if err != nil {
		log.Println("something went wrong:", err)
		return
	}
	feeds, err := fetchFeed(ctx, feedToFetch.Url)
	if err != nil {
		log.Println("something went wrong:", err)
		return
	}
	for _, item := range feeds.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}
		_, err := s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			FeedID:      feedToFetch.ID,
			Title:       item.Title,
			Description: sql.NullString{String: item.Description, Valid: true},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Error while reading the body")
	}
	rssfeed := RSSFeed{}
	err = xml.Unmarshal(body, &rssfeed)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling feed: %w", err)
	}
	rssfeed.Channel.Description = html.UnescapeString(rssfeed.Channel.Description)
	rssfeed.Channel.Title = html.UnescapeString(rssfeed.Channel.Title)
	for i := range rssfeed.Channel.Item {
		rssfeed.Channel.Item[i].Title = html.UnescapeString(rssfeed.Channel.Item[i].Title)
		rssfeed.Channel.Item[i].Description = html.UnescapeString(rssfeed.Channel.Item[i].Description)
	}
	return &rssfeed, nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	var limit int32 = 2
	if len(cmd.arguments) == 1 {
		parsed, err := strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(parsed)
	}
	posts, err := s.db.GetPostsForUser(ctx, database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("something wrong happened: %w", err)
	}
	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s\n", post.PublishedAt.Time.Format("Mon Jan 2"))
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("incorrect args for agg command")
	}
	interval, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("error parsing duration: %w", err)
	}
	fmt.Printf("Collecting feeds every %v...\n", interval)
	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("incorrect args for addfeed command")
	}
	ctx := context.Background()
	name := cmd.arguments[0]
	feedUrl := cmd.arguments[1]
	currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error in handlerAddfeed: %w", err)
	}
	newfeed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating the feed: %w", err)
	}
	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    newfeed.ID,
	})
	if err != nil {
		return fmt.Errorf("error in handlerAddfeed: %w", err)
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching feeds: %w", err)
	}
	for _, feed := range feeds {
		fmt.Printf("%v\n", feed.Name)
		fmt.Printf("%v\n", feed.UserName)
		fmt.Printf("%v\n", feed.Url)
	}
	return nil
}
