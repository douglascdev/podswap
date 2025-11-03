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
		log.Printf("returning due to unhandled webhook event: %v\n", event)
		response.WriteHeader(200)
		return
	}

	// Tell github the payload was delivered successfully
	response.WriteHeader(200)
}

func Start(ctx context.Context) error {
	server := &http.Server{Addr: ":8888"}
	http.HandleFunc("/webhook", WebhookHandler)

	var serveErrCh chan error
	go func() {
		log.Println("started server")
		err := server.ListenAndServe()
		log.Println("stopped the server")
		serveErrCh <- err
	}()

	for {
		var err error
		select {
		case err = <-serveErrCh:
			return err
		case <-ctx.Done():
			log.Println("context is done, stopping the server")
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				log.Printf("error trying to shut down server: %v\n", err)
				return err
			}

			return nil
		}
	}
}
