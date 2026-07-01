# Getting Started

## Prerequisites

- Go 1.23+
- CGO enabled
- NVIDIA drivers and NVML headers (Linux, for real GPU access)

Tests use a mock collector and run anywhere — no GPU hardware required.

## Build

```bash
make build       # compile binary (requires CGO + NVML on Linux)
make test        # run tests (no GPU needed — uses mock)
make lint        # golangci-lint
make docker      # container image
```

## Run

```bash
# stdio transport (default) — the agent spawns this as a child process
./gpu-mcp-server

# check version
./gpu-mcp-server --version
```

The server communicates over stdio using the MCP protocol.
It reads GPU state from NVML on demand — no polling, no daemon.

## Docker

```bash
make docker
docker run --gpus all --rm gpu-mcp-server:dev
```

## Helm (Kubernetes)

```bash
helm install gpu-mcp-server deploy/helm/gpu-mcp-server
```

## Verify

```bash
# run the test suite (works with no GPU)
make test

# check code quality
make lint
make vet
make fmt
```
