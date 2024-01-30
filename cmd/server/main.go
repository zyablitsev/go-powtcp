package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"powtcp/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	s, err := server.Init(ctx)
	if err != nil {
		log.Fatal(err)
	}

	s.Run()
}
