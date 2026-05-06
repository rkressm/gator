package main

import (
	"context"
	"fmt"

	"github.com/rkressm/gator/internal/database"
)

type commands struct {
	cmds map[string]func(*state, command) error
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		ctx := context.Background()
		user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("error getting user in middleware: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func commandsInitializer() (commands, error) {
	commandsList := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	commandsList.register("login", handlerLogin)
	commandsList.register("register", handlerRegister)
	commandsList.register("reset", handlerReset)
	commandsList.register("users", handlerUsers)
	commandsList.register("agg", handlerfetchFeed)
	commandsList.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commandsList.register("feeds", handlerFeeds)
	commandsList.register("follow", middlewareLoggedIn(handlerFollow))
	commandsList.register("following", middlewareLoggedIn(handlerFollowing))
	commandsList.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	return commandsList, nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	value, ok := c.cmds[cmd.name]
	if !ok {
		return fmt.Errorf("command not found by run")
	}
	return value(s, cmd)
}
