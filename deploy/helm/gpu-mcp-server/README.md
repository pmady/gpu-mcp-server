# gpu-mcp-server Helm chart

Deploys [gpu-mcp-server](https://github.com/pmady/gpu-mcp-server) as a
`DaemonSet` on Kubernetes GPU nodes, so every GPU node exposes its NVIDIA
metrics over MCP.

## Prerequisites

- Kubernetes 1.23+
- NVIDIA drivers on the GPU nodes and the
  [NVIDIA container toolkit](https://github.com/NVIDIA/k8s-device-plugin)
  (for runtime driver injection)
- Helm 3.8+

## Install

```bash
helm install gpu-mcp deploy/helm/gpu-mcp-server \
  --namespace gpu-monitoring --create-namespace
```

By default the chart runs the **http** transport and creates a `Service`, so
the server is reachable in-cluster at
`gpu-mcp-server.<namespace>.svc:8080`.

## How it reaches the GPU

The pod gets NVML access through the NVIDIA container runtime. The chart sets:

- `NVIDIA_VISIBLE_DEVICES=all`
- `NVIDIA_DRIVER_CAPABILITIES=utility` (metrics only — no compute/graphics)

`utility` exposes NVML without reserving a GPU, so the DaemonSet does not
consume allocatable `nvidia.com/gpu` capacity.

If you do not use the NVIDIA runtime, mount the driver library **and the
NVIDIA device nodes** from the host instead:

```bash
helm install gpu-mcp deploy/helm/gpu-mcp-server \
  --set nvidia.runtimeInjection=false \
  --set nvidia.hostMounts.enabled=true \
  --set nvidia.hostMounts.libDir=/usr/lib/x86_64-linux-gnu
```

This mounts `libnvidia-ml.so` plus the device nodes NVML needs
(`/dev/nvidiactl`, `/dev/nvidia-uvm`, `/dev/nvidia0`). Add a
`/dev/nvidia<N>` entry under `nvidia.hostMounts.devices` for every GPU on the
node. Accessing the device nodes usually requires relaxing the default
`securityContext` (e.g. `--set securityContext.readOnlyRootFilesystem=false`)
and may need the pod to run privileged depending on your node setup.

## Scheduling onto GPU nodes

The chart tolerates the `nvidia.com/gpu` taint by default. Pin it to GPU nodes
with a `nodeSelector` matching your cluster's labels, e.g.:

```bash
helm install gpu-mcp deploy/helm/gpu-mcp-server \
  --set nodeSelector."nvidia\.com/gpu\.present"=true
```

## Sidecar (stdio) mode

To run alongside an MCP client instead of as a network service:

```bash
helm install gpu-mcp deploy/helm/gpu-mcp-server --set transport=stdio
```

No `Service` or `ServiceMonitor` is created in stdio mode.

## Prometheus

A `ServiceMonitor` (Prometheus Operator) can be enabled with
`--set serviceMonitor.enabled=true`. It requires the http transport. The
`/metrics` path is reserved for a future metrics endpoint.

## Key values

| Key | Default | Description |
|-----|---------|-------------|
| `image.repository` | `ghcr.io/pmady/gpu-mcp-server` | Container image |
| `image.tag` | chart `appVersion` | Image tag |
| `transport` | `http` | `http` (standalone) or `stdio` (sidecar) |
| `port` | `8080` | http listen port |
| `nodeSelector` | `{}` | GPU node selector |
| `tolerations` | `nvidia.com/gpu:NoSchedule` | GPU node taint tolerations |
| `nvidia.runtimeInjection` | `true` | Use NVIDIA runtime to inject driver libs |
| `nvidia.hostMounts.enabled` | `false` | Mount NVML library + device nodes from the host instead |
| `nvidia.hostMounts.devices` | `nvidiactl, nvidia-uvm, nvidia0` | Host NVIDIA device nodes to expose |
| `resources` | 50m/64Mi … 200m/128Mi | Container resource requests/limits |
| `service.enabled` | `true` | Create a `Service` (http only) |
| `serviceMonitor.enabled` | `false` | Create a Prometheus `ServiceMonitor` |

See [values.yaml](values.yaml) for the full list.
