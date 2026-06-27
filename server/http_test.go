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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

// TestHTTP_EndToEnd drives a real MCP client over the HTTP transport: it
// connects, lists tools, and calls list_gpus, asserting the structured result
// matches the mock fixture. This exercises the full request/response path
// through the Streamable HTTP transport, not just routing.
func TestHTTP_EndToEnd(t *testing.T) {
	h := newTestHandler()
	srv := httptest.NewServer(h.HTTPHandler())
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: srv.URL}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"list_gpus", "get_gpu_metrics", "gpu_summary"} {
		if !names[want] {
			t.Errorf("tool %q not advertised over http; got %v", want, names)
		}
	}

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "list_gpus"})
	if err != nil {
		t.Fatalf("call list_gpus: %v", err)
	}
	if res.IsError {
		t.Fatalf("list_gpus returned tool error: %+v", res.Content)
	}

	var out ListGPUsOutput
	data, _ := json.Marshal(res.StructuredContent)
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode structured content: %v", err)
	}
	if out.Count != 2 {
		t.Errorf("count = %d, want 2", out.Count)
	}
	if len(out.Devices) != 2 || out.Devices[0].UUID != "GPU-aaaa-1111" {
		t.Errorf("devices = %+v, want first UUID GPU-aaaa-1111", out.Devices)
	}
}
