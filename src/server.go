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
	"strings"
	"time"

	"golang.ngrok.com/ngrok/v2"
)

type WebhookServer struct {
	buildCmd    string
	deployCmd   string
	workdir     string
	preBuildCmd string
	podswapReq  chan bool
}

func NewServer(preBuildCmd string, buildCmd string, deployCmd string, workdir string) *WebhookServer {
	return &WebhookServer{
		preBuildCmd: preBuildCmd,
		buildCmd:    buildCmd,
		deployCmd:   deployCmd,
		workdir:     workdir,
		podswapReq:  make(chan bool, 50),
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
	slog.Info("command runner is active")

	runPodswapReq := func(timeout time.Duration, cmdStr string) (cmdOutput string, err error) {
		cmdTimeoutCtx, buildCancel := context.WithTimeout(ctx, timeout)
		defer buildCancel()
		var cmd *exec.Cmd
		args := strings.Split(cmdStr, " ")
		switch len(args) {
		case 0:
			return "", fmt.Errorf("failed to run podswap request, command %s is empty", cmd)
		case 1:
			cmd = exec.CommandContext(cmdTimeoutCtx, args[0])
		default:
			cmd = exec.CommandContext(cmdTimeoutCtx, args[0], args[1:]...)
		}
		cmd.Dir = s.workdir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), fmt.Errorf("failed to run buildCmd: %w", err)
		}

		return string(out), nil
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("command runner stopped")
			return
		case <-s.podswapReq:
			slog.Info("command runner received a podswap request")
			// TODO: command argument for timeout duration
			timeout := time.Second * 500

			out, err := runPodswapReq(timeout, s.preBuildCmd)
			if err != nil {
				slog.Error("pre-build command failed", slog.Any("err", err), slog.Any("output", out))
				continue
			}
			slog.Info("pre-build command succeeded", slog.Any("output", out))

			out, err = runPodswapReq(timeout, s.buildCmd)
			if err != nil {
				slog.Error("build command failed", slog.Any("err", err), slog.Any("output", out))
				continue
			}
			slog.Info("build command succeeded", slog.Any("output", out))

			out, err = runPodswapReq(timeout, s.deployCmd)
			if err != nil {
				slog.Error("deploy command failed", slog.Any("err", err), slog.Any("output", out))
				continue
			}
			slog.Info("deploy command succeeded", slog.Any("output", out))
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
