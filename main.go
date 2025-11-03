package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	podswap "podswap/src"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	fmt.Println("Press Ctrl+C to trigger a graceful shutdown.")
	err := podswap.Start(ctx)
	log.Printf("server err: %v", err)
	<-ctx.Done()
	fmt.Println("Main routine exiting. All workers have been notified.")
}
