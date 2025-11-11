package podswap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func WebhookHandler(response http.ResponseWriter, request *http.Request) {
	slog.Info("received webhook payload")

	event := request.Header.Get("x-github-event")
	switch event {
	case "push":
		slog.Info("push webhook payload received, continuing")
	case "":
		slog.Info("x-github-event header not found, bad request")
		response.WriteHeader(400)
		return
	default:
		slog.Info(fmt.Sprintf("returning due to unhandled webhook event: %v\n", event))
		response.WriteHeader(200)
		return
	}

	// Tell github the payload was delivered successfully
	response.WriteHeader(200)
}

func Start(ctx context.Context, arguments *Arguments) error {
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", *arguments.Host, *arguments.Port)}
	http.HandleFunc("/webhook", WebhookHandler)

	var serveErrCh chan error
	go func() {
		slog.Info("started server")
		err := server.ListenAndServe()
		slog.Info("stopped the server")
		serveErrCh <- err
	}()

	for {
		var err error
		select {
		case err = <-serveErrCh:
			return err
		case <-ctx.Done():
			slog.Info("context is done, stopping the server")
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				slog.Error(fmt.Sprintf("error trying to shut down server: %v\n", err))
				return err
			}

			return nil
		}
	}
}
