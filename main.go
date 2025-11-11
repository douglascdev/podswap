package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
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
		err := fmt.Errorf("failed to parse arguments: %v", err)
		slog.Error(err.Error())
		return
	}

	if _, isSet := os.LookupEnv("NGROK_AUTHTOKEN"); !isSet {
		slog.Warn("environment variable NGROK_AUTHTOKEN not set, sign up at https://dashboard.ngrok.com/signup")
	}

	if _, isSet := os.LookupEnv("WEBHOOK_SECRET"); !isSet {
		slog.Warn("environment variable WEBHOOK_SECRET not set")
	}

	var (
		buildCmd  = *arguments.BuildCommand
		deployCmd = *arguments.DeployCommand
		workdir   = arguments.WorkDir
	)

	slog.Info(fmt.Sprintf("using buildCmd %q", buildCmd))
	slog.Info(fmt.Sprintf("using deployCmd %q", deployCmd))
	slog.Info(fmt.Sprintf("using workdir %q", workdir))

	slog.Info("Press Ctrl+C to trigger a graceful shutdown.")

	err = podswap.Start(ctx, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("server err: %v", err))
	}
	<-ctx.Done()
	slog.Info("main routine exiting.")
}
