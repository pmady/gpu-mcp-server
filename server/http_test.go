/*
Copyright 2026 The gpu-mcp-server Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHealthz(t *testing.T) {
	h := newTestHandler()
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHTTPHandler_MCPEndpointMounted(t *testing.T) {
	h := newTestHandler()
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	// A bare GET to the MCP endpoint is not a valid session request, but it
	// must be handled by the streamable transport rather than 404 — proving the
	// MCP handler is mounted at the root.
	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		t.Errorf("root returned 404, expected the MCP handler to be mounted")
	}
}

func TestRunHTTP_GracefulShutdown(t *testing.T) {
	h := newTestHandler()
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() { errCh <- h.RunHTTP(ctx, "127.0.0.1:0") }()

	// Give the listener a moment, then cancel and expect a clean return.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("RunHTTP returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunHTTP did not shut down within 2s")
	}
}

func TestHealthzBody(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.HTTPHandler().ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "ok") {
		t.Errorf("body = %q, want to contain ok", rec.Body.String())
	}
}
