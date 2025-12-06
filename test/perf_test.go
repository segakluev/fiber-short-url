package test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
}

// ===========================
// 1) Sequential 10000 requests
// ===========================

func TestSequential10000(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	// Увеличили timeout
	client := &http.Client{Timeout: 500 * time.Millisecond}

	start := time.Now()

	var okCount int
	for i := 0; i < 10000; i++ {
		resp, err := client.Get(server.URL + "/hit")
		if err == nil {
			okCount++
			resp.Body.Close()
		}
	}

	elapsed := time.Since(start)
	rps := float64(okCount) / elapsed.Seconds()

	t.Logf("Sequential 10000: %v | %.2f req/sec | OK=%d", elapsed, rps, okCount)
}
