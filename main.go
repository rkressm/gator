package main

import (
	"fmt"
	"os"

	"github.com/rkressm/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = cfg.SetUser("Remi")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfg2, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", cfg2)
}
