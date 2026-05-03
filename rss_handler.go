package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
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

func handlerfetchFeed(s *state, cmd command) error {
	ctx := context.Background()
	feed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error in handlerfetchfee %w", err)
	}
	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("not enough args for addfeed command")
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
	fmt.Printf("%+v\n", newfeed)
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
