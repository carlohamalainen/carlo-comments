package main

import (
	"fmt"

	"github.com/carlohamalainen/carlo-comments/config"
	"github.com/carlohamalainen/carlo-comments/simple"

	"golang.org/x/term"
	"syscall"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	us := simple.NewUserService(cfg.HmacSecret)

	fmt.Printf("Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(us.HashPassword(bytePassword)))
}
