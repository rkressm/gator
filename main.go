package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/rkressm/gator/internal/config"
	"github.com/rkressm/gator/internal/database"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Not enough args")
		os.Exit(1)
	}
	commandsList := commands{
		cmds: make(map[string]func(*state, command) error),
	}
	commandsList.register("login", handlerLogin)
	commandsList.register("register", handlerRegister)
	commandsList.register("reset", handlerReset)
	commandsList.register("users", handlerUsers)
	actualCommand := command{
		name:      args[1],
		arguments: args[2:],
	}
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	actualState := state{}
	actualState.cfg = &cfg
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Printf("error connecting to db: %v\n", err)
		os.Exit(1)
	}
	actualState.db = database.New(db)
	err = commandsList.run(&actualState, actualCommand)
	if err != nil {
		fmt.Println("Error when running the command:", err)
		os.Exit(1)
	}
}
