---
title: 'gpu-mcp-server: NVIDIA GPU metrics for AI agents over the Model Context Protocol'
tags:
  - GPU
  - observability
  - Model Context Protocol
  - AI agents
  - NVML
  - Go
authors:
  - name: Pavan Madduri
    orcid: 0009-0007-3795-7593
    affiliation: 1
affiliations:
  - name: Independent Researcher, USA
    index: 1
date: 20 September 2026
bibliography: paper.bib
---

# Summary

`gpu-mcp-server` exposes NVIDIA GPU telemetry to AI agents through the Model
Context Protocol (MCP) [@mcp]. MCP is an open protocol that lets large language
model applications call external tools through a standard interface;
`gpu-mcp-server` implements that interface and provides GPU state as a small set
of tools. It reads metrics directly from the NVIDIA Management Library (NVML)
[@nvml] and reports per-device utilization, memory use, temperature, power draw,
and, where available, per-process usage and Multi-Instance GPU (MIG) partitions.
The server is built on the official Go MCP SDK and the Go NVML bindings, runs
over stdio or streamable HTTP, and can be deployed as a container or a Kubernetes
DaemonSet.

# Statement of need

AI agents are increasingly used to observe and operate infrastructure, but they
have no standard way to see GPU state. In practice an agent must shell out to
`nvidia-smi` and parse free-form text, which is brittle, inconsistent across
driver versions, and awkward to expose safely. At the same time, GPUs are the
scarce and expensive resource in machine learning infrastructure, so accurate,
structured, real-time visibility into utilization and memory is exactly what an
agent needs to reason about capacity, cost, and placement.

`gpu-mcp-server` addresses this gap by presenting GPU telemetry as first-class
MCP tools with typed inputs and outputs. Any MCP-compatible agent (for example,
assistants integrated with editors or agent runtimes) can query GPU state
directly, without hardware-specific glue code. Because it reads NVML rather than
scraping command output, the data is structured and stable, and the same
telemetry engine underlies the author's related Kubernetes GPU autoscaler,
keeping metric semantics consistent across tools.

# Functionality

The server registers a small set of MCP tools, including:

- `list_gpus`: enumerate GPUs with utilization and memory.
- `get_gpu_metrics`: detailed metrics for a specific GPU or MIG instance.
- `gpu_summary`: aggregate statistics across all devices.
- `get_gpu_processes`: per-process (PID-level) GPU usage where supported.

It supports stdio and streamable HTTP transports, ships a Helm chart for
Kubernetes deployment as a DaemonSet, and reports a build version for
reproducibility. Tests run against a mock collector, so the server can be
exercised on machines without a GPU.

# Acknowledgements

The project builds on the Model Context Protocol [@mcp] and the NVIDIA Management
Library [@nvml].

# References
