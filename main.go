package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()
	fmt.Println("Press Ctrl+C to trigger a graceful shutdown.")
	<-ctx.Done()
	fmt.Println("Main routine exiting. All workers have been notified.")
}
