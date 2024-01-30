package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"powtcp/client"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	c, err := client.Init(ctx)
	if err != nil {
		log.Fatal(err)
	}

	c.Run()
}
