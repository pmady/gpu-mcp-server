# External Dependencies

| Dependency | License | Purpose |
|------------|---------|---------|
| [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) | MIT | Official MCP Go SDK server, tools, transport |
| [github.com/NVIDIA/go-nvml](https://github.com/NVIDIA/go-nvml) | Apache 2.0 | Go bindings for the NVIDIA Management Library (NVML) |

## Standard Library

All other imports are from the Go standard library (`context`, `encoding/json`,
`fmt`, `log`, `log/slog`, `os`, `os/signal`, `sync`, `strings`, `errors`).

## Build Requirements

- **Go 1.23+**
- **CGO** (required for go-nvml on Linux)
- **NVIDIA drivers** with NVML headers (for production builds)
- Tests run without CGO or GPU hardware using the mock collector
