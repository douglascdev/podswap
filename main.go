package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	podswap "podswap/src"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	flagset := flag.NewFlagSet("main", flag.ExitOnError)
	flagset.SetOutput(os.Stdout)
	arguments, err := podswap.ParseArguments(flagset, os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse arguments: %s", err)
	}

	var (
		port    = *arguments.Port
		host    = *arguments.Host
		workdir = arguments.WorkDir
	)

	log.Printf("using port \"%d\"", port)
	log.Printf("using host %q", host)
	log.Printf("using workdir %q", workdir)

	fmt.Println("Press Ctrl+C to trigger a graceful shutdown.")

	err = podswap.Start(ctx, arguments)
	if err != nil {
		log.Printf("Server err: %v", err)
	}
	<-ctx.Done()
	fmt.Println("Main routine exiting. All workers have been notified.")
}
