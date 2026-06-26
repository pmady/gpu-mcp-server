# gpu-mcp-server

An [MCP](https://modelcontextprotocol.io/) server that exposes NVIDIA GPU metrics as tools.
Any MCP-compatible AI agent (Claude, Goose, Cursor, etc.) can query real-time GPU
utilization, memory, temperature, power, PCIe and NVLink throughput no Prometheus
or dcgm-exporter required.

Built on the [official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
and [NVIDIA go-nvml](https://github.com/NVIDIA/go-nvml).

## Tools

| Tool | Description |
|------|-------------|
| `list_gpus` | List all GPUs with utilization and memory info |
| `get_gpu_metrics` | Detailed metrics for a GPU by index or UUID |
| `gpu_summary` | Aggregate stats across all devices |

All tools support MIG (Multi-Instance GPU) - MIG instances appear as separate
devices with their parent GPU's shared metrics (temperature, power, PCIe).

## Quick start

```bash
# build (requires CGO + NVML headers on Linux)
make build

# run the server communicates over stdio
./gpu-mcp-server
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gpu": {
      "command": "/path/to/gpu-mcp-server"
    }
  }
}
```

### Goose

```yaml
extensions:
  gpu-metrics:
    type: stdio
    cmd: /path/to/gpu-mcp-server
```

## Build

Requires Go 1.23+, CGO, and NVIDIA drivers on the target machine.

```bash
make build       # compile binary
make test        # run tests (no GPU needed uses mock)
make lint        # golangci-lint
make docker      # container image
```

Tests use a mock collector, so they run anywhere no GPU hardware required.

## Architecture

```
Agent (Claude/Goose) ─── MCP (stdio) ──→ gpu-mcp-server ──→ NVML ──→ GPU
                                              │
                                         Tools:
                                         • list_gpus
                                         • get_gpu_metrics
                                         • gpu_summary
```

The server runs as a local process alongside the agent. It calls NVML directly
through cgo — no sidecar, no network hops, no metric pipeline to configure.

## Project info

- **License:** Apache 2.0
- **Language:** Go
- **AAIF project alignment:** [MCP](https://modelcontextprotocol.io/)
- **Related:** [keda-gpu-scaler](https://github.com/pmady/keda-gpu-scaler) (GPU autoscaling for Kubernetes)

## Roadmap

See [ROADMAP.md](ROADMAP.md) for the 12-month public roadmap.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to get involved.

## Governance

This project follows [Linux Foundation Minimum Viable Governance](GOVERNANCE.md).

## Documentation

- [ROADMAP.md](ROADMAP.md) - public roadmap
- [GOVERNANCE.md](GOVERNANCE.md) - decision-making process
- [DEPENDENCIES.md](DEPENDENCIES.md) - external dependencies and licenses
- [SECURITY.md](SECURITY.md) - vulnerability reporting
- [AGENTS.md](AGENTS.md) - instructions for AI agents working on this repo
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) - community standards
