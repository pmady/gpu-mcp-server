# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### New
- **feat:** `get_gpu_processes` tool for PID-level GPU process attribution (#21)
- **feat:** Streamable HTTP transport with `--transport` and `--port` flags (#20 — pending rebase)
- **feat:** `--version`/`-v` flag to print build version (#18)
- **feat:** Helm chart for Kubernetes DaemonSet deployment (#23)

### Improvements
- **docs:** Sample JSON output section in README (#17)
- **docs:** Cursor and Windsurf MCP client config examples (#22)
- **ci:** CGO build verification job (#21)
- **ci:** Helm chart lint, render, unit test, and OCI publish workflow (#23)
- **ci:** Upgrade golangci-lint-action to v7 for v2 config support

### Fixes
- **fix:** Handle `collector.Close()` return value (errcheck)
- **fix:** Migrate `.golangci.yml` to v2 schema

## v0.1.0 — 2026-05-15

### New
- Initial release
- MCP server exposing NVIDIA GPU metrics via stdio transport
- Tools: `list_gpus`, `get_gpu_metrics`, `gpu_summary`
- MIG (Multi-Instance GPU) support
- Mock collector for testing without GPU hardware
- Docker image and CI workflow
