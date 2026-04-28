package main

import (
	"fmt"
)

type commands struct {
	cmds map[string]func(*state, command) error
}

func commandsInitializer() (commands, error) {
	commandsList := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	commandsList.register("login", handlerLogin)
	commandsList.register("register", handlerRegister)
	commandsList.register("reset", handlerReset)
	commandsList.register("users", handlerUsers)
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
