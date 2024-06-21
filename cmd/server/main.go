package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"powtcp"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	s, err := powtcp.Init(ctx)
	if err != nil {
		err = fmt.Errorf("main: %w", err)
		log.Fatal(err)
	}

	s.Run()
}
