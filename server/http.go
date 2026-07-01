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
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const httpShutdownTimeout = 5 * time.Second

// HTTPHandler returns the HTTP handler that serves the MCP endpoint over the
// Streamable HTTP transport plus a /healthz health check.
func (h *Handler) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)

	streamable := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return h.srv },
		nil,
	)
	mux.Handle("/", streamable)

	return mux
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok\n")
}

// RunHTTP serves the MCP endpoint over Streamable HTTP on addr (e.g. ":8080").
// It shuts down gracefully when ctx is cancelled.
func (h *Handler) RunHTTP(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: h.HTTPHandler(),
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Warn("http shutdown", "err", err)
		}
	}()

	slog.Info("serving over http", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
