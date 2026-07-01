# Changelog

All notable changes to this project will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/).

## [v0.1.0] - 2026

### Added

- Core MCP server with stdio transport
- `list_gpus` tool — enumerate all GPUs with utilization and memory
- `get_gpu_metrics` tool — detailed per-GPU metrics (temp, power, PCIe, NVLink)
- `get_gpu_processes` tool — PID-level GPU process attribution
- `gpu_summary` tool — aggregate stats across all devices
- NVML-based collector with MIG support
- Mock collector for testing without GPU hardware
- `--version` flag
- GitHub Actions CI (build, test, lint)
- Helm chart workflow
- Dockerfile (multi-stage build)
- Helm chart under `deploy/helm/gpu-mcp-server`
- `server.json` MCP registry manifest
- Apache 2.0 license
- GOVERNANCE.md (LF Minimum Viable Governance)
- SECURITY.md (private disclosure, 48-hour response)
- AI_GUIDELINES.md (AI contribution policy)
- AGENTS.md (instructions for AI agents)
