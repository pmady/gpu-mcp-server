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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"

	"github.com/pmady/gpu-mcp-server/gpu"
	"github.com/pmady/gpu-mcp-server/server"
)

var version = "dev"

func main() {
	transport := flag.String("transport", "stdio", "transport to serve on: stdio or http")
	port := flag.Int("port", 8080, "TCP port for the http transport")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	collector, err := gpu.NewNVML()
	if err != nil {
		log.Fatalf("nvml: %v", err)
	}
	defer collector.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv := server.New(collector, version)

	switch *transport {
	case "stdio":
		if err := srv.Run(ctx); err != nil {
			log.Fatalf("server: %v", err)
		}
	case "http":
		addr := fmt.Sprintf(":%d", *port)
		if err := srv.RunHTTP(ctx, addr); err != nil {
			log.Fatalf("server: %v", err)
		}
	default:
		log.Fatalf("unknown transport %q (want stdio or http)", *transport)
	}
}
