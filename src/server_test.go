package podswap_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"podswap/src"
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

			testServer := httptest.NewUnstartedServer(nil)
			gotErr := podswap.Start(ctx, testServer.URL)
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
		response http.ResponseWriter
		request  *http.Request
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			podswap.WebhookHandler(tt.response, tt.request)
		})
	}
}
