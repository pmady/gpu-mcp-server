# gpu-mcp-server

An [MCP](https://modelcontextprotocol.io/) server that exposes NVIDIA GPU metrics as tools.
Any MCP-compatible AI agent (Claude, Goose, Cursor, Windsurf) can query real-time GPU
utilization, memory, temperature, power, PCIe and NVLink throughput — no Prometheus
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

All tools support MIG (Multi-Instance GPU) — MIG instances appear as separate
devices with their parent GPU's shared metrics (temperature, power, PCIe).

## Quick links

- [Getting Started](getting-started.md) — build and run in under 5 minutes
- [Architecture](architecture.md) — how the server and Collector interface work
- [Tools Reference](tools.md) — detailed input/output for each tool
- [Agent Integration](agent-integration.md) — copy-paste configs for Claude, Goose, Cursor, Windsurf

## Project info

- **License:** Apache 2.0
- **Language:** Go
- **AAIF alignment:** [MCP](https://modelcontextprotocol.io/)
- **Related:** [keda-gpu-scaler](https://github.com/pmady/keda-gpu-scaler) — GPU autoscaling for Kubernetes
