package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/signal"
	podswap "podswap/src"
	"syscall"
)

type Config struct {
	Port uint   `json:"port"`
	Host string `json:"host"`
}

func NewConfig() *Config {
	return &Config{
		Port: 8888,
		Host: "localhost",
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	var (
		template = flag.Bool("template", false, "output generated config template to stdout")
	)
	flag.Parse()

	if *template {
		cfg := NewConfig()
		s, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			log.Fatalf("failed to generate config template")
		}
		fmt.Printf("%s", s)
		return
	}

	fmt.Println("Press Ctrl+C to trigger a graceful shutdown.")

	err := podswap.Start(ctx)
	if err != nil {
		log.Printf("Server err: %v", err)
	}
	<-ctx.Done()
	fmt.Println("Main routine exiting. All workers have been notified.")
}
