package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Archiit19/School-management/pkg/correlation"
)

func TestDoContextPropagatesRequestID(t *testing.T) {
	const wantID = "trace-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get(correlation.RequestIDHeader); got != wantID {
			t.Fatalf("request id = %q, want %q", got, wantID)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(server.URL, "")
	ctx := correlation.ContextWithRequestID(context.Background(), wantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	resp, err := client.DoContext(ctx, req)
	if err != nil {
		t.Fatalf("DoContext: %v", err)
	}
	resp.Body.Close()
}
