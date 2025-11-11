package podswap

import (
	"context"
	"crypto"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"golang.ngrok.com/ngrok/v2"
)

type WebhookServer struct {
	buildCmd   string
	deployCmd  string
	workdir    string
	podswapReq chan bool
}

func NewServer(buildCmd string, deployCmd string, workdir string) *WebhookServer {
	return &WebhookServer{
		buildCmd:   buildCmd,
		deployCmd:  deployCmd,
		workdir:    workdir,
		podswapReq: make(chan bool, 50),
	}
}

func (s *WebhookServer) WebhookHandler(response http.ResponseWriter, request *http.Request) {
	slog.Info("received webhook payload")

	event := request.Header.Get("x-github-event")
	switch event {
	case "push":
		slog.Info("push webhook payload received, continuing")
	case "":
		slog.Info("x-github-event header not found, bad request")
		response.WriteHeader(http.StatusBadRequest)
		return
	default:
		slog.Info(fmt.Sprintf("returning due to unhandled webhook event: %v\n", event))
		response.WriteHeader(http.StatusOK)
		return
	}

	env, _ := os.LookupEnv("WEBHOOK_SECRET")
	envHash := hmac.New(crypto.SHA256.New, []byte(env))
	body, err := io.ReadAll(request.Body)
	if err != nil {
		slog.Error("failed to read body", slog.Any("error", err))
		response.WriteHeader(http.StatusBadRequest)
	}
	envHash.Write(body)
	expectedSignature := fmt.Sprintf("sha256=%s", hex.EncodeToString(envHash.Sum(nil)))
	headerSignature := request.Header.Get("X-Hub-Signature-256")
	if !hmac.Equal([]byte(expectedSignature), []byte(headerSignature)) {
		slog.Error("webhook secret does not match", slog.String("headerSecret", headerSignature), slog.String("expectedSignature", expectedSignature))
		response.WriteHeader(http.StatusForbidden)
		return
	}
	slog.Info("webhook triggered successfully, forwarding podswap request")
	s.podswapReq <- true

	// Tell github the payload was delivered successfully
	response.WriteHeader(http.StatusOK)
}

func (s *WebhookServer) commandRunner(ctx context.Context) {
	slog.Info("command runner is waiting")

	for {
		select {
		case <-ctx.Done():
			slog.Info("command runner stopped")
			return
		case <-s.podswapReq:
			slog.Info("command runner received a podswap request")
			// TODO: command argument for timeout duration
			buildTimeoutCtx, buildCancel := context.WithTimeout(ctx, time.Second*500)
			defer buildCancel()
			buildCmd := exec.CommandContext(buildTimeoutCtx, s.buildCmd)
			buildCmd.Dir = s.workdir
			err := buildCmd.Run()
			if err != nil {
				slog.Error("failed to run buildCmd", slog.Any("err", err))
				continue
			}

			deployTimeoutCtx, deployCancel := context.WithTimeout(ctx, time.Second*500)
			defer deployCancel()
			deployCmd := exec.CommandContext(deployTimeoutCtx, s.deployCmd)
			deployCmd.Dir = s.workdir
			err = deployCmd.Run()
			if err != nil {
				slog.Error("failed to run deployCmd", slog.Any("err", err))
				continue
			}
		}
	}
}

func (s *WebhookServer) Start(ctx context.Context, ln net.Listener) error {
	var (
		err error
		url string
	)
	if ln == nil {
		ln, err = ngrok.Listen(ctx)
		if err != nil {
			return fmt.Errorf("failed to start ngrok listener: %v", err)
		}
		url = fmt.Sprintf("https://%s", ln.Addr().String())
	} else {
		url = ln.Addr().String()
	}

	slog.Info("ngrok listener started", slog.String("url", url))

	var serveErrCh chan error
	go func() {
		slog.Info("started server")
		err := http.Serve(ln, http.HandlerFunc(s.WebhookHandler))
		slog.Info("stopped server")
		serveErrCh <- err
	}()

	go s.commandRunner(ctx)

	for {
		var err error
		select {
		case err = <-serveErrCh:
			return err
		case <-ctx.Done():
			slog.Info("context is done, stopping the server")
			ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			closedWithErr := make(chan error)
			go func() {
				closedWithErr <- ln.Close()
			}()

			select {
			case err = <-closedWithErr:
				if err != nil {
					slog.Error(fmt.Sprintf("error trying to shut down server: %v\n", err))
				}
			case <-ctxTimeout.Done():
				slog.Error("server connection refused to close gracefully, ending anyway")
			}

			return nil
		}
	}
}
