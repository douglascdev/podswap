package podswap_test

import (
	"context"
	"flag"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	podswap "podswap/src"
	"strings"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		wantErr bool
	}{
		{"Start runs and stops gracefully", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(time.Second)
				cancel()
			}()

			_, err := podswap.ParseArguments(flag.NewFlagSet("", flag.PanicOnError), []string{})
			if err != nil {
				t.Fatal(err)
			}
			listener, err := net.Listen("tcp", ":8888")
			if err != nil {
				t.Fatalf("failed to open listener: %v", err)
			}
			gotErr := podswap.Start(ctx, listener)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Start() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Start() succeeded unexpectedly")
			}
		})
	}
}

func TestWebhookHandler(t *testing.T) {
	validRequest := httptest.NewRequest("POST", "/", strings.NewReader("Hello, World!"))
	validRequest.Header.Set("X-Hub-Signature-256", "sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17")

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		response      http.ResponseWriter
		request       *http.Request
		expectedCode  int
		githubEvent   string
		webhookSecret string
	}{
		{
			"push",
			httptest.NewRecorder(),
			validRequest,
			http.StatusOK,
			"push",
			"It's a Secret to Everybody",
		},
		{
			"unhandled",
			httptest.NewRecorder(),
			validRequest,
			http.StatusOK,
			"asd",
			"It's a Secret to Everybody",
		},
		{
			"empty",
			httptest.NewRecorder(),
			validRequest,
			http.StatusBadRequest,
			"",
			"It's a Secret to Everybody",
		},
		{
			"invalid secret",
			httptest.NewRecorder(),
			validRequest,
			http.StatusForbidden,
			"push",
			"invalid secret",
		},
	}

	for _, test := range tests {
		test.request.Header.Set("x-github-event", test.githubEvent)
		os.Setenv("WEBHOOK_SECRET", test.webhookSecret)
		podswap.WebhookHandler(test.response, test.request)
		response := test.response.(*httptest.ResponseRecorder)
		if response.Code != test.expectedCode {
			t.Errorf("request with %q event should return %d, got %d", test.name, test.expectedCode, response.Code)
		}
	}
}
