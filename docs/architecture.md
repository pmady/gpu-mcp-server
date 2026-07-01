# Architecture

## Overview

```
Agent (Claude/Goose/Cursor) ── MCP (stdio) ──> gpu-mcp-server ──> NVML ──> GPU
```

The server runs as a local process alongside the agent. It calls NVML directly
through CGO — no sidecar, no network hops, no metric pipeline to configure.

## Components

### Entry point

`cmd/gpu-mcp-server/main.go` — parses flags, initializes the NVML collector,
creates the MCP server, and runs the stdio transport.

### Collector interface

`gpu/gpu.go` defines the `Collector` interface:

```go
type Collector interface {
    All() ([]Metrics, error)
    ByIndex(index int) (Metrics, error)
    ByUUID(uuid string) (Metrics, error)
    Count() (int, error)
    Processes() ([]ProcessInfo, error)
    Close() error
}
```

Two implementations:

- **`gpu/nvml.go`** — real NVML collector (Linux + CGO only, build tag `cgo && linux`)
- **`gpu/mock.go`** — deterministic fake with 2 A100s, used in tests

The `gpu/stub.go` file provides a compile-time stub for non-Linux platforms.

### MCP server

`server/server.go` — registers the four tools with the MCP SDK and implements
handlers that delegate to the `Collector`.

### Tests

`server/server_test.go` — tests call handler methods directly with `gpu.NewMock()`.
No MCP transport needed, no GPU required.

## Design decisions

- **Read-only** — every tool observes; nothing mutates GPU state
- **stdio transport** — local, private, no network surface
- **Swappable Collector** — the interface seam makes the entire server testable without hardware
- **Minimal dependencies** — only the official MCP Go SDK and go-nvml
