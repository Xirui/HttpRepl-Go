package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type recordedReq struct {
	Method string
	Path   string
	Header http.Header
	Body   string
}

func TestIntegration(t *testing.T) {
	var mu sync.Mutex
	var lastReq recordedReq

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		bodyBytes, _ := io.ReadAll(r.Body)
		lastReq = recordedReq{
			Method: r.Method,
			Path:   r.URL.Path,
			Header: r.Header,
			Body:   string(bodyBytes),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	baseAddr = server.URL
	gLabel = "/api/v1/users"

	// Clear headers
	gHeaders = make(map[string]string)

	// 1. GET requests
	t.Run("GET request with no subpath", func(t *testing.T) {
		makeRequest("GET", []string{"get"})
		
		mu.Lock()
		req := lastReq
		mu.Unlock()

		if req.Method != "GET" {
			t.Errorf("expected GET, got %s", req.Method)
		}
		if req.Path != "/api/v1/users" {
			t.Errorf("expected /api/v1/users, got %s", req.Path)
		}
	})

	t.Run("GET request with subpath", func(t *testing.T) {
		makeRequest("GET", []string{"get", "info"})

		mu.Lock()
		req := lastReq
		mu.Unlock()

		if req.Path != "/api/v1/users/info" {
			t.Errorf("expected /api/v1/users/info, got %s", req.Path)
		}
	})

	// 2. Custom headers
	t.Run("Header commands", func(t *testing.T) {
		// Set header
		handleHeaderCommand([]string{"set", "header", "X-Test-Header", "HelloHeader"})
		if gHeaders["X-Test-Header"] != "HelloHeader" {
			t.Errorf("expected header to be set")
		}

		// Make a GET request and verify header is sent
		makeRequest("GET", []string{"get"})
		
		mu.Lock()
		req := lastReq
		mu.Unlock()

		if req.Header.Get("X-Test-Header") != "HelloHeader" {
			t.Errorf("expected header X-Test-Header to be HelloHeader, got %s", req.Header.Get("X-Test-Header"))
		}

		// Clear header
		handleClearCommand([]string{"clear", "header", "X-Test-Header"})
		if _, ok := gHeaders["X-Test-Header"]; ok {
			t.Errorf("expected header to be cleared")
		}
	})

	// 3. POST request with content
	t.Run("POST request with inline content", func(t *testing.T) {
		makeRequest("POST", []string{"post", "-c", `{"name":"test"}`})

		mu.Lock()
		req := lastReq
		mu.Unlock()

		if req.Method != "POST" {
			t.Errorf("expected POST, got %s", req.Method)
		}
		if req.Body != `{"name":"test"}` {
			t.Errorf("expected request body to match, got %s", req.Body)
		}
		if req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", req.Header.Get("Content-Type"))
		}
	})

	// 4. PUT/PATCH/DELETE/HEAD/OPTIONS requests
	t.Run("PUT request", func(t *testing.T) {
		makeRequest("PUT", []string{"put", "-c", "put-body"})
		mu.Lock()
		req := lastReq
		mu.Unlock()
		if req.Method != "PUT" {
			t.Errorf("expected PUT, got %s", req.Method)
		}
	})

	t.Run("DELETE request", func(t *testing.T) {
		makeRequest("DELETE", []string{"delete"})
		mu.Lock()
		req := lastReq
		mu.Unlock()
		if req.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", req.Method)
		}
	})

	t.Run("OPTIONS request", func(t *testing.T) {
		makeRequest("OPTIONS", []string{"options"})
		mu.Lock()
		req := lastReq
		mu.Unlock()
		if req.Method != "OPTIONS" {
			t.Errorf("expected OPTIONS, got %s", req.Method)
		}
	})
}
