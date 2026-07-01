# gpu-mcp-server

[![CI](https://github.com/pmady/gpu-mcp-server/actions/workflows/ci.yml/badge.svg)](https://github.com/pmady/gpu-mcp-server/actions/workflows/ci.yml)
[![Helm](https://github.com/pmady/gpu-mcp-server/actions/workflows/helm.yaml/badge.svg)](https://github.com/pmady/gpu-mcp-server/actions/workflows/helm.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pmady/gpu-mcp-server)](https://goreportcard.com/report/github.com/pmady/gpu-mcp-server)
[![Go Reference](https://pkg.go.dev/badge/github.com/pmady/gpu-mcp-server.svg)](https://pkg.go.dev/github.com/pmady/gpu-mcp-server)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/pmady/gpu-mcp-server/badge)](https://securityscorecards.dev/viewer/?uri=github.com/pmady/gpu-mcp-server)

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
| `get_gpu_processes` | PID-level GPU process attribution |
| `gpu_summary` | Aggregate stats across all devices |

All tools support MIG (Multi-Instance GPU) - MIG instances appear as separate
devices with their parent GPU's shared metrics (temperature, power, PCIe).

## Sample output

Each tool returns structured JSON. The examples below show the shape of the
data an agent receives from a node with two NVIDIA A100 GPUs.

`list_gpus`:

```json
{
  "count": 2,
  "devices": [
    {
      "index": 0,
      "uuid": "GPU-aaaa-1111",
      "name": "NVIDIA A100-SXM4-80GB",
      "gpu_utilization_percent": 85,
      "memory_used_mib": 57344,
      "memory_total_mib": 81920
    },
    {
      "index": 1,
      "uuid": "GPU-bbbb-2222",
      "name": "NVIDIA A100-SXM4-80GB",
      "gpu_utilization_percent": 20,
      "memory_used_mib": 12288,
      "memory_total_mib": 81920
    }
  ]
}
```

`get_gpu_metrics` (with `{"index": 0}` or `{"uuid": "GPU-aaaa-1111"}`):

```json
{
  "index": 0,
  "uuid": "GPU-aaaa-1111",
  "name": "NVIDIA A100-SXM4-80GB",
  "gpu_utilization_percent": 85,
  "memory_utilization_percent": 70,
  "memory_used_mib": 57344,
  "memory_total_mib": 81920,
  "temperature_celsius": 72,
  "power_draw_watts": 300,
  "power_limit_watts": 400,
  "pcie_tx_kbps": 0,
  "pcie_rx_kbps": 0,
  "nvlink_tx_mbps": 0,
  "nvlink_rx_mbps": 0
}
```

`gpu_summary`:

```json
{
  "device_count": 2,
  "avg_gpu_utilization": 52.5,
  "avg_memory_utilization": 42.5,
  "total_memory_used_mib": 69632,
  "total_memory_total_mib": 163840,
  "max_temperature_celsius": 72,
  "total_power_draw_watts": 375
}
```

MIG instances add `is_mig`, `parent_gpu`, and `mig_profile` fields to the
`get_gpu_metrics` and `list_gpus` payloads.

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

### Cursor

Add to `.cursor/mcp.json` for a project, or `~/.cursor/mcp.json` for all
projects:

```json
{
  "mcpServers": {
    "gpu": {
      "type": "stdio",
      "command": "/path/to/gpu-mcp-server"
    }
  }
}
```

### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "gpu": {
      "command": "/path/to/gpu-mcp-server"
    }
  }
}
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

## Contributors

Thanks to all our [contributors](CONTRIBUTORS.md)! Add yourself via PR.

## Governance

This project follows [Linux Foundation Minimum Viable Governance](GOVERNANCE.md).

## Documentation

- [Full documentation](https://gpu-mcp-server.readthedocs.io) - hosted on Read the Docs
- [ROADMAP.md](ROADMAP.md) - public roadmap
- [GOVERNANCE.md](GOVERNANCE.md) - decision-making process
- [DEPENDENCIES.md](DEPENDENCIES.md) - external dependencies and licenses
- [SECURITY.md](SECURITY.md) - vulnerability reporting
- [AGENTS.md](AGENTS.md) - instructions for AI agents working on this repo
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) - community standards

## Star History

<a href="https://www.star-history.com/?repos=pmady%2Fgpu-mcp-server&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=pmady/gpu-mcp-server&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=pmady/gpu-mcp-server&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=pmady/gpu-mcp-server&type=date&legend=top-left" />
 </picture>
</a>
