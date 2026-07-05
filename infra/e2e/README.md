# Manual e2e validation

`infra/terraform/<cloud>` stands up a throwaway GPU cluster with `gpu-mcp-server`
already installed from the in-tree chart. This directory holds the procedure —
and the load manifests — to **validate a deployed cluster by hand**: confirm the
DaemonSet runs on every GPU node and that the server reports GPU metrics that
match `nvidia-smi` ground truth, on both architectures, idle and under load.

There is intentionally **no apply/verify/destroy automation**. A real GPU
cluster costs money and needs quota, so `terraform apply` and `terraform
destroy` stay deliberate human steps (see the stack READMEs). Nothing here
creates or destroys infrastructure.

## Prerequisites

- The stack applied (`cd infra/terraform/aws && terraform apply`) and `kubectl`
  pointed at the cluster (run the `configure_kubectl` output, e.g.
  `aws eks update-kubeconfig --region us-west-2 --name gpu-mcp-server-e2e`).
- `kubectl`, and — for the interactive MCP checks — an MCP client (Claude
  Desktop or Claude Code) that can run `kubectl`.

## 1. Cluster & DaemonSet health

```bash
kubectl get nodes
kubectl get pods -n gpu-mcp
```

Expect every GPU node `Ready` with one `1/1 Running` `gpu-mcp-server-*` pod. On
AWS that is two nodes / two pods — one amd64 (Tesla T4) and one arm64 (NVIDIA
T4G); on Azure and GCP a single amd64 node (Tesla T4) each. Grab the pod names
and their arch (the pod names change every rollout — don't hard-code them):

```bash
kubectl -n gpu-mcp get pods \
  -o custom-columns='POD:.metadata.name,NODE:.spec.nodeName,ARCH:.metadata.labels.kubernetes\.io/arch'
```

## 2. Ground truth: `nvidia-smi`

Read the GPU directly on each pod (substitute the pod name from step 1):

```bash
kubectl -n gpu-mcp exec <pod> -- nvidia-smi \
  --query-gpu=name,uuid,utilization.gpu,memory.used,memory.reserved,memory.total,temperature.gpu,power.draw,power.limit \
  --format=csv,noheader,nounits | \
  column -t -s ',' \
  -N "Name,UUID,GPU%,MemUsed,MemReserved,MemTotal,Temp,PowerDraw,PowerLimit"
```

> **On memory:** `memory.used` counts application allocations only; the
> ~264 MiB the driver reserves shows up under `memory.reserved`. The server
> reads NVML's `GetMemoryInfo` (`used = total − free`), which folds those
> together — so the server's "memory used" ≈ `MemUsed + MemReserved`. Not a
> discrepancy, just two definitions of "used".

## 3. Cross-check via the MCP server

Connect an MCP client to each pod's `gpu-mcp-server` over stdio and compare its
numbers to step 2. Claude Desktop `claude_desktop_config.json` (one entry per
arch; on Windows + WSL route through `wsl` so it uses your working kubeconfig
and AWS credentials):

```json
{
  "mcpServers": {
    "gpu-mcp-amd64": {
      "command": "wsl",
      "args": ["-e","kubectl","exec","-i","-n","gpu-mcp","<amd64-pod>","--","gpu-mcp-server","--transport","stdio"]
    },
    "gpu-mcp-arm64": {
      "command": "wsl",
      "args": ["-e","kubectl","exec","-i","-n","gpu-mcp","<arm64-pod>","--","gpu-mcp-server","--transport","stdio"]
    }
  }
}
```

Claude Code equivalent:

```bash
claude mcp add gpu-mcp-amd64 -- kubectl exec -i -n gpu-mcp <amd64-pod> -- gpu-mcp-server --transport stdio
claude mcp add gpu-mcp-arm64 -- kubectl exec -i -n gpu-mcp <arm64-pod> -- gpu-mcp-server --transport stdio
```

Then ask Claude to confirm connectivity and render a metrics table, one row per
pod/GPU, for example:

> 1. Confirm you are connected to both MCP servers `gpu-mcp-amd64` and
>    `gpu-mcp-arm64`.
> 2. If connected, query each and output a table (one row per GPU): name, uuid,
>    utilization.gpu, memory.used, memory.total, temperature.gpu, power.draw,
>    power.limit.

The server's values should match your step-2 `nvidia-smi` output — exactly for
the static fields (name, memory total, power limit, uuid) and within sampling
skew for utilization/temperature/power.

## 4. Put the GPUs under load

Idle GPUs read 0% utilization, so drive a short burn to see the numbers move.
The manifests here each target one architecture via `nodeSelector`:

```bash
kubectl apply -f infra/e2e/gpu-burn-amd64.yaml   # oguzpastirmaci/gpu-burn (amd64 only)
kubectl apply -f infra/e2e/gpu-burn-arm64.yaml   # compiles a burn in nvidia/cuda:*-devel (multi-arch)
```

While they run (~60s each; the arm64 CUDA image is a few GB, so its pod may sit
`ContainerCreating` on the first pull), re-run the step-2 `nvidia-smi` command
and ask Claude to *"refresh the GPU metrics table"*. Both GPUs should climb to
~100% utilization, pinned at the 70 W T4 power cap, with the server tracking the
change. Clean up:

```bash
kubectl delete job gpu-burn-amd64 gpu-burn-arm64
```

> On Azure and GCP (single amd64 node) only `gpu-burn-amd64.yaml` applies —
> there is no arm64 GPU node to load.

## Done

Teardown is a separate, deliberate step:

```bash
cd infra/terraform/<cloud>
terraform destroy
```
