package podswap

import (
	"context"
	"log"
	"net/http"
	"time"
)

func WebhookHandler(response http.ResponseWriter, request *http.Request) {
	log.Println("received webhook payload")

	event := request.Header.Get("x-github-event")
	switch event {
	case "push":
		log.Println("push webhook payload received, continuing")
	case "":
		log.Println("x-github-event header not found, bad request")
		response.WriteHeader(400)
		return
	default:
		log.Printf("returning due to unhandled webhook event: %v", event)
		response.WriteHeader(200)
		return
	}

	// Tell github the payload was delivered successfully
	response.WriteHeader(200)
}

func Start(ctx context.Context, addr string) error {
	server := &http.Server{Addr: addr}
	http.HandleFunc("/webhook", WebhookHandler)

	var serveErrCh chan error
	go func() {
		err := server.ListenAndServe()
		serveErrCh <- err
	}()

	for {
		var err error
		select {
		case err = <-serveErrCh:
			return err
		case <-ctx.Done():
			log.Print("stopping server")
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				log.Printf("error trying to shut down server: %v", err)
				return err
			}

			return nil
		}
	}
}
