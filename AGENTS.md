# AGENTS.md

Instructions for AI agents working with this repository.

## Project Overview

gpu-mcp-server is an MCP server that exposes NVIDIA GPU metrics as tools.
Built on the official Go MCP SDK and NVIDIA go-nvml.

## Build

```bash
make build    # requires CGO + NVIDIA drivers (Linux)
make test     # runs tests using mock collector (any platform)
make lint     # golangci-lint
```

## Code Layout

```
cmd/gpu-mcp-server/main.go   entry point
gpu/gpu.go                    Metrics struct + Collector interface
gpu/nvml.go                   real NVML collector (linux+cgo)
gpu/stub.go                   stub for non-Linux platforms
gpu/mock.go                   mock for tests
server/server.go              MCP server + tool handlers
server/server_test.go         tests
```

## Key Patterns

- **Collector interface** in `gpu/gpu.go` abstracts hardware access
- **Build tags**: `gpu/nvml.go` uses `//go:build cgo && linux`, `gpu/stub.go` covers everything else
- **Tool handlers** are methods on `server.Handler`, registered via `mcp.AddTool`
- Tests call handler methods directly with a `gpu.Mock` no MCP transport needed

## Testing

Tests use `gpu.NewMock()` with static device data. No GPU required.

```bash
make test
```

## Adding a New Tool

1. Define input/output structs in `server/server.go` with `json` and `jsonschema` tags
2. Add a handler method on `*Handler`
3. Register with `mcp.AddTool` in the `New()` function
4. Add tests in `server/server_test.go`

## Dependencies

Only two external dependencies:
- `github.com/modelcontextprotocol/go-sdk` MCP protocol
- `github.com/NVIDIA/go-nvml` NVML bindings

## Rules

- Do not add dependencies without maintainer approval
- Keep comments terse match existing style
- All PRs need tests
- Sign off commits (DCO)
