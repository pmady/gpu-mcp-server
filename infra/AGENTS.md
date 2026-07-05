# AGENTS.md — `infra/` guidance

Guidance for humans and AI agents working on the Infrastructure-as-Code and
manual e2e validation in this directory. Keep changes consistent with the
conventions below.

## Purpose

`infra/` provisions **throwaway, GPU-ready Kubernetes clusters** to
end-to-end-validate `gpu-mcp-server` images and its Helm chart against **real
NVIDIA hardware**, including arm64 (Graviton + T4G). The Go unit test suite
only exercises a mock GPU collector (`gpu.NewMock()`), so these clusters are
the only place the server is validated against an actual driver, NVML, and a
real Kubernetes GPU device plugin. They are test clusters, not production
infrastructure.

## Follow the project's contribution rules

This file **supplements, it does not replace** the repository's contribution
policy. Read and follow both:

- **[CONTRIBUTING.md](../CONTRIBUTING.md)** — workflow, build/test, code style,
  pull requests.
- **[AI_GUIDELINES.md](../AI_GUIDELINES.md)** — policy for AI-assisted work.

The rules that apply to any change here:

- **Human-verify everything.** Every AI-assisted change must be reviewed,
  understood, and tested before submitting — you own what you commit.
- **Run the checks before pushing.** For Terraform run `terraform fmt` /
  `terraform validate`; for any Go changes run `make fmt` / `make test` /
  `make lint`.
- **Match the house style.** Terse, single-line comments; no verbose godoc or
  over-polished prose.
- **One logical change per PR.** Don't let tooling expand scope into adjacent
  code (one cloud / one concern per PR).
- **Disclose significant AI usage** in the PR description.
- **Sign off every commit** (`git commit -s`) — required by
  [AI_GUIDELINES.md](../AI_GUIDELINES.md) and the DCO.

## Layout

```
infra/
  terraform/
    aws/        # Amazon EKS (implemented)
    azure/      # Azure AKS  (implemented)
    gcp/        # Google GKE (implemented)
    README.md   # index + per-cloud details
  e2e/
    README.md          # manual validation procedure (nodes -> nvidia-smi -> MCP -> load)
    gpu-burn-amd64.yaml # short GPU load job for the amd64 node
    gpu-burn-arm64.yaml # short GPU load job for the arm64 node
```

Each cloud under `terraform/` is a **self-contained, independently
`apply`-able** Terraform stack (its own providers, variables, outputs, state).
They do not share a root module, so a new cloud is a sibling directory
following the same conventions — no rework of existing stacks. `e2e/` documents
the cloud-agnostic manual validation you run against a deployed stack.

## What every stack must do

One `terraform apply` produces a cluster that is immediately ready for e2e
tests, with no manual post-apply steps:

1. a cluster + a GPU node pool (AWS runs **two** — one per CPU architecture —
   so both the amd64 and arm64 `gpu-mcp-server` images get validated; Azure and
   GCP have no arm64 NVIDIA GPU SKU, so they validate linux/amd64 only — AWS is
   the only dual-arch stack);
2. the GPU made usable in Kubernetes on every GPU node — a working driver, the
   NVIDIA device plugin, the `nvidia` RuntimeClass, and the
   `nvidia.com/gpu.present` node label;
3. `gpu-mcp-server`, installed from the in-tree chart at
   `deploy/helm/gpu-mcp-server` so the cluster always runs the local version,
   not a published release.

Provide consistent outputs (`configure_kubectl`, `cluster_name`, `region` /
`location`, `mcp_namespace`, `mcp_release_name`, `mcp_http_endpoint`) and tag
every resource (`Project=gpu-mcp-server`, `Component=gpu-e2e-test`,
`ManagedBy=terraform`, `Stack=infra/terraform/<cloud>`) so a forgotten cluster
is easy to find.

## gpu-mcp-server chart requirements

Always read `deploy/helm/gpu-mcp-server/values.yaml` before wiring a cluster —
it drives exactly what a stack has to provide:

- **Unprivileged.** The container runs with `allowPrivilegeEscalation: false`,
  `readOnlyRootFilesystem: true`, and drops all capabilities. It reaches NVML
  either via **NVIDIA runtime injection** (`nvidia.runtimeInjection: true`
  + `nvidia.runtimeClassName`, the default here) or, as a fallback for
  clusters without a working RuntimeClass, via `nvidia.hostMounts` (mounts
  `libnvidia-ml.so` and the `/dev/nvidia*` device nodes from the host).
- **`nodeSelector: nvidia.com/gpu.present=true`** — the chart's default
  `nodeSelector` is `{}` (unset); every stack here overrides it via Helm `set`
  to the label the GPU operator's GPU-feature-discovery applies, otherwise the
  DaemonSet schedules everywhere.
- **Tolerates `nvidia.com/gpu:NoSchedule`.** A no-op on these stacks (the GPU
  pools are deliberately left untainted so the GPU operator's own controllers
  and CoreDNS can co-locate), but harmless to leave in place.
- **`transport: http` on port 8080**, with liveness/readiness probes against
  `/healthz`. An MCP client can also drive the server over a `--transport stdio`
  `kubectl exec` session — see `infra/e2e/README.md`.

Satisfy the driver/RuntimeClass/device-plugin/label requirements with the
**NVIDIA GPU operator**. Whether the operator or the node AMI/image owns the
driver differs per cloud (AWS: AMI ships it, `driver.enabled=false`; Azure:
`gpu_driver = "None"` on the node pool, operator owns it) — state which
approach a stack uses and why in its own files.

## Conventions

- **Pin versions.** Confirm the latest Terraform release, provider versions,
  module versions, and chart versions before writing — don't rely on memory.
  Add `.terraform-version` and floor `required_version` at the current minor.
- **Current Kubernetes version.** Default to a version still in standard
  support; never default to a near-EOL minor.
- **Fixed, predictable pools.** On-demand GPU node groups sized by a `*_count`
  variable, no cluster autoscaler, no spot — predictable and cheap for tests.
  A count of `0` should cleanly omit a node group (see AWS's dual-arch pools).
- **Install add-ons explicitly.** Ensure the GPU operator brings up the device
  plugin and every GPU node reaches `Ready` and advertises
  `nvidia.com/gpu.present=true` before the chart install is attempted.
- **Order the GPU operator before the chart.** `gpu-mcp-server`'s
  `nodeSelector` and `runtimeClassName` both depend on objects the operator
  creates, so use `depends_on` to install the operator first.
- **Set a real image tag.** The chart's `appVersion` may not have a matching
  published image; set `mcp_image_tag` (and `mcp_image_repository` if needed)
  to a real, published tag of `ghcr.io/pmady/gpu-mcp-server` when testing a
  specific build.

## Scope: e2e validation, not production deployment

These stacks exist to answer one question — "does this image, on this chart,
work on real NVIDIA hardware, on every architecture we ship for" — not to
model a production deployment. There's deliberately no HA control plane
tuning, no autoscaling, no ingress/TLS, and no long-lived state. Anything a
production cluster would need but an e2e run doesn't is out of scope here.

## GPU service quota (all clouds)

Cloud GPU quota is typically **0 on fresh accounts**, and is per-region (AWS)
or per-location (Azure) and per-GPU-family. Identify the exact quota and
request an increase **before** applying, or the apply fails at node creation.
Document the quota name, running cost, and `terraform destroy` loudly in each
stack's README section.

- AWS: "Running On-Demand G and VT instances" (`L-DB2E81BA`), measured in
  vCPUs per region. The dual-arch default (`g4dn.xlarge` + `g5g.xlarge`) needs
  **8** vCPUs; zeroing one node group's count needs **4**.
- Azure: "Standard NCASv3_T4 Family vCPUs", per location. The default
  `Standard_NC4as_T4_v3` node needs **4** vCPUs.
- GCP: two separate quotas, both needed — the regional **"NVIDIA T4 GPUs"**
  quota in your `region`, and the global **"GPUs (all regions)"** quota. Both
  need to be **at least 1**.

## Cost & teardown

GPU clusters bill by the hour. Always `terraform destroy` after testing, and
remind users to do so — apply and destroy are manual steps here. Resource tags
make any leftovers findable — see the README's per-cloud teardown sections.

## Validation (before claiming a stack works)

Apply on real hardware and run the manual procedure in `infra/e2e/README.md` —
confirm:

- every GPU node reaches `Ready` and is running a `Ready` `gpu-mcp-server`
  pod (on AWS, this must hold for **both** `gpu_amd64` and `gpu_arm64` node
  groups unless one was intentionally zeroed);
- an MCP client completes a real session against each pod (`list_gpus` /
  `get_gpu_metrics` / `gpu_summary`) over both the http and stdio transports,
  and the values match `nvidia-smi` ground truth — idle and under the
  `gpu-burn-*` load, on both architectures;
- `terraform destroy` actually removes everything — check for leftovers by tag
  afterward.

## Adding a new cloud

- Create `infra/terraform/<cloud>/` mirroring `infra/terraform/aws/` —
  structure, variable naming, outputs, and a "loud" README section covering
  quota, cost, and teardown.
- Use a well-maintained community cluster module where one materially
  simplifies the stack (as AWS does); go native where it doesn't (as Azure
  does) — pin whichever you choose.
- Default to a cheap, current-generation GPU SKU, easily overridable.
- Cover it in `infra/e2e/README.md` so it gets the same manual
  nodes -> `nvidia-smi` -> MCP -> load validation as AWS and Azure.
- Follow the contribution rules above (checks, terse style, DCO sign-off, AI
  disclosure) and open one PR per cloud.
