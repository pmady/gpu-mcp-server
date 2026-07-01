# Roadmap

Public roadmap for gpu-mcp-server. Updated quarterly.

## Q3 2026 (Current)

- [x] Core MCP server with stdio transport
- [x] `list_gpus`, `get_gpu_metrics`, `gpu_summary` tools
- [x] `get_gpu_processes` tool
- [x] NVML-based collector with MIG support
- [x] Mock collector for testing without GPU hardware
- [x] GitHub Actions CI
- [x] Publish to MCP servers registry
- [x] First tagged release (v0.1.0)
- [x] Container image on GitHub Container Registry (ghcr.io)

## Q4 2026

- [ ] Streamable HTTP transport (run as a network service)
- [ ] MCP resources expose GPU info as context for agents
- [ ] Per-process GPU usage (PID-level memory and compute attribution)
- [ ] Prometheus metrics endpoint (optional, alongside MCP)
- [ ] Integration tests with Claude Desktop and Goose

## Q1 2027

- [ ] Multi-node support — aggregate metrics from remote hosts
- [ ] GPU event notifications (thermal throttling, ECC errors, XID events)
- [ ] OpenSSF Best Practices badge
- [x] Helm chart for Kubernetes deployment

## Q2 2027

- [ ] AMD ROCm support (via rocm-smi)
- [ ] Intel GPU support (via oneAPI Level Zero)
- [ ] MCP sampling integration — let agents request GPU snapshots at intervals
- [ ] AAIF project proposal submission (Growth stage)

## How to contribute

Pick an item from the roadmap, open an issue to discuss the approach, and submit
a PR. See [Contributing](contributing.md) for details.
