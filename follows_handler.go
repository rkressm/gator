package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rkressm/gator/internal/database"
)

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error getting user in handlerFollowing: %w", err)
	}
	followed, err := s.db.GetFeedFollowsForUser(ctx, currentUser.ID)
	if err != nil {
		return fmt.Errorf("error getting followed in handlerFollowing: %w", err)
	}
	for _, elem := range followed {
		fmt.Printf("%s\n", elem.FeedName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("not enough arguments")
	}
	url := cmd.arguments[0]
	ctx := context.Background()
	currentUser, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error getting user in handlerFollow: %w", err)
	}
	currentFeed, err := s.db.GetFeedByUrl(ctx, url)
	if err != nil {
		return fmt.Errorf("error getting feed in handlerFollow: %w", err)
	}
	record, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    currentFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating record: %w", err)
	}
	fmt.Printf("%s\n", record.FeedName)
	fmt.Printf("%s\n", record.UserName)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("need url argument")
	}
	ctx := context.Background()
	feed, err := s.db.GetFeedByUrl(ctx, cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("error in handlerUnfollow: %w", err)
	}
	err = s.db.DeleteRecord(ctx, database.DeleteRecordParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error in handlerUnfollow during deleterecord: %w", err)
	}
	return nil
}
