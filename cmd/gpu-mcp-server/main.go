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
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version", "-v":
			fmt.Println(version)
			return
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	collector, err := gpu.NewNVML()
	if err != nil {
		log.Fatalf("nvml: %v", err)
	}
	defer collector.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	srv := server.New(collector, version)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}
