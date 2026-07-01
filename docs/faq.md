# FAQ

## Do I need a GPU to develop or test?

No. Tests use a mock collector that returns deterministic data for 2 A100 GPUs.
Run `make test` on any platform.

## How does this compare to dcgm-exporter?

dcgm-exporter feeds Prometheus dashboards for humans.
gpu-mcp-server gives agents direct tool access — real-time, structured JSON,
no query language needed.

## Can an agent control or modify the GPU?

No. Every tool is read-only by design. The server observes GPUs and never
changes clocks, kills processes, or allocates resources.

## What about security and authentication?

stdio transport = local only. The server runs as a child process of the agent
with no network endpoint to authenticate. If/when the HTTP transport lands,
auth becomes an explicit design item.

## Does it support multi-GPU?

Yes. NVML enumerates all local GPUs. `list_gpus` returns every device.

## Does it support multi-node?

Not yet. Today it's per-node (stdio, local process). Multi-node aggregation
is on the Q1 2027 roadmap, likely paired with the HTTP transport.

## Does it support MIG?

Yes. MIG instances appear as separate devices with `is_mig`, `parent_gpu`,
and `mig_profile` fields. An agent can address them individually.

## What are the dependencies?

Only two external dependencies:

- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) — MCP protocol
- [NVIDIA/go-nvml](https://github.com/NVIDIA/go-nvml) — NVML bindings

See [DEPENDENCIES.md](https://github.com/pmady/gpu-mcp-server/blob/main/DEPENDENCIES.md) for details.

## What about AMD or Intel GPUs?

AMD ROCm and Intel oneAPI are on the Q2 2027 roadmap. The `Collector` interface
is the seam that makes adding vendors clean — a new backend, same tools and JSON.
