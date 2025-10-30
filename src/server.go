package podswap

import (
	"context"
	"log"
	"net/http"
	"time"
)

func webhookHandler(response http.ResponseWriter, request *http.Request) {

}

func Start(ctx context.Context) error {
	server := &http.Server{Addr: ":6666"}
	http.HandleFunc("/webhook", webhookHandler)

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
