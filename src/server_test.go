package podswap_test

import (
	"context"
	"flag"
	"net/http"
	"net/http/httptest"
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

			args, err := podswap.ParseArguments(flag.NewFlagSet("", flag.PanicOnError), []string{})
			if err != nil {
				t.Fatal(err)
			}
			gotErr := podswap.Start(ctx, args)
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
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		response     http.ResponseWriter
		request      *http.Request
		expectedCode int
		githubEvent  string
	}{
		{
			"Push",
			httptest.NewRecorder(),
			httptest.NewRequest("POST", "/webhook", strings.NewReader("")),
			200,
			"push",
		},
		{
			"Unhandled",
			httptest.NewRecorder(),
			httptest.NewRequest("POST", "/webhook", strings.NewReader("")),
			200,
			"asd",
		},
		{
			"Empty",
			httptest.NewRecorder(),
			httptest.NewRequest("POST", "/webhook", strings.NewReader("")),
			400,
			"",
		},
	}

	for _, tt := range tests {
		if tt.githubEvent != "" {
			tt.request.Header.Set("x-github-event", tt.githubEvent)
		}
		podswap.WebhookHandler(tt.response, tt.request)
		response := tt.response.(*httptest.ResponseRecorder)
		if response.Code != tt.expectedCode {
			t.Errorf("request with %s event should return %d, got %d", tt.name, tt.expectedCode, response.Code)
		}
	}
}
