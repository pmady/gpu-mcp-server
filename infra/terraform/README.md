> [!WARNING]
> ## GPU service quota and cost — read this first, your apply could fail, and you may spend a lot of money
>
> Real GPU hardware is not cheap. The AWS default (two node groups, one per
> architecture) runs roughly **~$1.1/hr all-in** — about **$26/day**. The
> Azure default is cheaper, roughly **~$0.55/hr (~$13/day)**. The GCP default
> is similarly cheap, roughly **~$0.65/hr (~$16/day)**.
>
> **Bring the infra up, run the e2e tests, destroy everything.** Don't leave a
> cluster running "just in case" — see [Teardown](#teardown) in each section.
>
> Fresh AWS accounts almost always have a GPU instance quota of **0**, and the
> apply will fail at node group creation with an insufficient-capacity / quota
> error.
>
> The relevant Service Quota is **"Running On-Demand G and VT instances"**
> (quota code `L-DB2E81BA`), measured in **vCPUs**, **per region**. The default
> dual-arch layout uses `g4dn.xlarge` (4 vCPUs) **and** `g5g.xlarge` (4 vCPUs),
> so it needs a quota of **at least 8** in your chosen region. Zeroing out one
> of `gpu_amd64_node_count` / `gpu_arm64_node_count` (testing a single
> architecture) drops that to **4**.
>
> Request an increase before applying: Service Quotas console → **Amazon
> EC2** → *Running On-Demand G and VT instances* → request ≥ 8 (or ≥ 4 for a
> single-arch run). Approval can take anywhere from minutes to a couple of
> days. Verify with:
>
> ```bash
> aws service-quotas get-service-quota \
>   --service-code ec2 --quota-code L-DB2E81BA --region us-west-2
> ```
>
> Azure is the same story: fresh subscriptions have a GPU vCPU quota of **0**
> for the NC / ND / NV VM families, per region. The default T4 node draws on
> **"Standard NCASv3_T4 Family vCPUs"** (`Standard_NC4as_T4_v3` = 4 vCPUs), so
> request **≥ 4** before applying — see the
> [Azure quota section](#gpu-vcpu-quota--your-apply-fails-without-it) below.
> Verify with:
>
> ```bash
> az vm list-usage --location eastus \
>   --query "[?contains(name.value, 'NCASv3_T4')]" -o table
> ```
>
> GCP tells the same story: fresh projects have a GPU quota of **0**. Two
> quotas gate the default T4 node pool — the regional **"NVIDIA T4 GPUs"**
> quota in your `region` and the global **"GPUs (all regions)"** quota —
> both need to be **at least 1** before applying. See the
> [GCP quota section](#gpu-quota--your-apply-fails-without-it) below. Verify
> with:
>
> ```bash
> gcloud compute regions describe us-central1 --format="value(quotas)"
> ```

# Infrastructure-as-Code for gpu-mcp-server

Terraform for standing up **throwaway** GPU-ready Kubernetes clusters to
end-to-end test `gpu-mcp-server` against real NVIDIA hardware — including
arm64 (Graviton + NVIDIA T4G). These are test clusters, not production
infrastructure; see [`infra/e2e`](../e2e) for the harness that drives them.

AWS (EKS), Azure (AKS), and GCP (GKE) are implemented today.

## Layout

```
infra/terraform/
  aws/        # Amazon EKS (implemented)
  azure/      # Azure AKS  (implemented)
  gcp/        # Google GKE (implemented)
```

Each cloud lives in its own self-contained, independently `apply`-able
directory (its own providers, variables, outputs, state). They deliberately do
**not** share a root module, so adding a new cloud is a matter of dropping in
a sibling that follows the same convention — no rework of the existing
stacks.

The shared contract every directory aims to honour:

- one `terraform apply` produces a cluster immediately ready for e2e tests:
  every GPU node has a working NVIDIA driver, the device plugin, the
  `nvidia.com/gpu.present` label and the `nvidia` RuntimeClass, and
  `gpu-mcp-server` is running from the in-tree chart at
  `deploy/helm/gpu-mcp-server`;
- the same `configure_kubectl` / `mcp_*` output names across clouds;
- every resource tagged so a forgotten cluster is easy to find and destroy.

## Status

| Target | Directory | Status |
|---|---|---|
| AWS EKS | [`aws/`](./aws) | ✅ Implemented (dual-arch: amd64 + arm64) |
| Azure AKS | [`azure/`](./azure) | ✅ Implemented (single GPU node, amd64 only) |
| GCP GKE | [`gcp/`](./gcp) | ✅ Implemented (single GPU node, amd64 only) |

## Conventions

- **Terraform version** is pinned per directory via `.terraform-version`
  (currently `1.15.6`); `required_version` floors at the current minor
  (`>= 1.15.0`).
- **Providers are version-pinned** (`aws ~> 6.51`, `azurerm ~> 4.79`,
  `google ~> 6.12`, `kubernetes ~> 3.2`, `helm ~> 3.2`), confirmed against the
  Terraform Registry at authoring time.
- **`.terraform.lock.hcl` is generated, not hand-written** — running
  `terraform init` in a stack directory for the first time creates it. It's
  worth committing once generated, so everyone (and CI, if it's ever added)
  resolves the exact same provider builds.
- **CI is manual only** — a human (or the `infra/e2e` harness) runs
  `terraform apply` locally. Intentionally **not** wired into a scheduled
  GitHub Actions workflow: a real GPU cluster needs GPU quota and costs money
  per run.

---

# AWS EKS GPU test cluster

One `terraform apply` provisions everything and leaves nothing manual:

- a small VPC (3 AZs, single NAT gateway),
- an EKS control plane,
- **two** fixed-size, on-demand EKS managed node groups — `gpu_amd64`
  (`g4dn.xlarge`, 1x NVIDIA T4) and `gpu_arm64` (`g5g.xlarge`, Graviton2 + 1x
  NVIDIA T4G) — so the amd64 **and** arm64 `gpu-mcp-server` images both get
  validated on real hardware in one cluster; either count can be set to `0` to
  test a single architecture,
- the **NVIDIA GPU operator** (device plugin, GPU-feature-discovery node
  labels, DCGM, and the `nvidia` RuntimeClass — multi-arch, so one release
  covers both node groups), and
- **`gpu-mcp-server`**, installed from the in-tree chart at
  `deploy/helm/gpu-mcp-server` so the cluster always runs the local version.

It uses well-maintained community modules (`terraform-aws-modules/vpc`
6.6.1, `terraform-aws-modules/eks` 21.23.0) rather than hand-rolled
networking/EKS resources.

## Prerequisites

- **Terraform 1.15.6** — pinned in [`aws/.terraform-version`](./aws/.terraform-version)
  (use `tfenv` to match it exactly).
- **awscli v2** on `PATH` with valid credentials for the target account/region.
  The Kubernetes/Helm providers call `aws eks get-token` to authenticate.
- **kubectl** (for poking at the cluster after apply; not required by
  Terraform itself).
- The **GPU service quota** above.
- Registry access from the machine running Terraform: `terraform init`
  fetches the VPC/EKS modules and the aws/kubernetes/helm providers from the
  public Terraform Registry, and the apply pulls the GPU operator chart from
  `helm.ngc.nvidia.com`.

## Usage

```bash
cd infra/terraform/aws

cp terraform.tfvars.example terraform.tfvars   # optional: override defaults

terraform init
terraform apply

# Point kubectl at the new cluster (also emitted as the `configure_kubectl` output)
aws eks update-kubeconfig --region us-west-2 --name gpu-mcp-server-e2e

# Confirm both GPU node groups are visible and gpu-mcp-server is running on each
kubectl get nodes -L nvidia.com/gpu.present,kubernetes.io/arch
kubectl -n gpu-mcp get pods -o wide
```

`gpu-mcp-server` is reachable in-cluster at the `mcp_http_endpoint` output,
e.g. `http://gpu-mcp-server.gpu-mcp.svc.cluster.local:8080` — MCP over
Streamable HTTP at `/`, health at `/healthz`.

## Common overrides

| Variable | Default | Notes |
|---|---|---|
| `region` | `us-west-2` | Choose one with GPU capacity + your quota. |
| `gpu_amd64_instance_type` | `g4dn.xlarge` (T4, 4 vCPUs) | amd64 pool. |
| `gpu_arm64_instance_type` | `g5g.xlarge` (Graviton2 + T4G, 4 vCPUs) | arm64 pool. |
| `gpu_amd64_node_count` / `gpu_arm64_node_count` | `1` / `1` | Set either to `0` to omit that node group entirely (e.g. single-arch testing). |
| `kubernetes_version` | `1.35` | EKS control plane version (latest is 1.36; keep to a version in standard support). |
| `gpu_operator_chart_version` | `v26.3.2` | NVIDIA GPU operator chart. |
| `mcp_image_repository` / `mcp_image_tag` | chart defaults | Pin the exact `gpu-mcp-server` image under test. |

```bash
# Single-architecture run: amd64 only
terraform apply -var 'gpu_arm64_node_count=0'
```

## Cost

You are paying for real GPU hardware — **destroy it when you're done.** Rough
on-demand list prices (us-west-2, USD; check current pricing for your region
— these are ballparks to verify, not quotes):

| Component | Approx. cost |
|---|---|
| EKS control plane | ~$0.10/hr |
| 1x `g4dn.xlarge` (`gpu_amd64`) | ~$0.526/hr |
| 1x `g5g.xlarge` (`gpu_arm64`) | ~$0.42/hr |
| NAT gateway | ~$0.045/hr + data processing |

Ballpark: **~$1.1/hr all-in** with the dual-arch defaults. Zeroing one node
group removes its line above.

## Teardown

```bash
terraform destroy
```

This removes everything this stack created. If a `terraform destroy` is ever
interrupted, the resource tags make leftovers easy to find:

```bash
# Every resource is tagged Project=gpu-mcp-server, Component=gpu-e2e-test, ManagedBy=terraform
aws resourcegroupstaggingapi get-resources \
  --tag-filters Key=Project,Values=gpu-mcp-server --region us-west-2
```

## How the cluster satisfies the chart

`gpu-mcp-server` runs **unprivileged** (`allowPrivilegeEscalation: false`,
`readOnlyRootFilesystem: true`, all capabilities dropped — see
`deploy/helm/gpu-mcp-server/values.yaml`) and reaches NVML through **NVIDIA
container runtime injection**, not a privileged/hostPath mount:

| Chart requirement | Provided by |
|---|---|
| `nodeSelector: nvidia.com/gpu.present=true` (set by Terraform via Helm `set`) | GPU-feature-discovery (GPU operator) labels both node groups |
| `nvidia.runtimeClassName: nvidia` (set by Terraform via `mcp_runtime_class_name`) | GPU operator creates the `nvidia` RuntimeClass |
| driver + NVML libraries injected into the unprivileged container | the AL2023 NVIDIA AMIs (`AL2023_x86_64_NVIDIA` / `AL2023_ARM_64_NVIDIA`) pre-install the driver and container toolkit; the operator runs with `driver.enabled=false`, `toolkit.enabled=false` |
| `tolerations: nvidia.com/gpu` | harmless no-op here — both GPU node groups are intentionally untainted so the GPU operator's controllers and CoreDNS can co-locate |

Because both node groups are untainted, the GPU operator's controllers and
CoreDNS schedule on GPU nodes alongside `gpu-mcp-server`. If you taint GPU
nodes, add a separate CPU node group for those system pods.

---

# Azure AKS GPU test cluster

Sibling to the AWS stack. One `terraform apply` provisions everything, no
manual steps:

- a resource group,
- an AKS control plane (Free tier — Microsoft-managed API server, no
  control-plane charge),
- **one** on-demand GPU node as the cluster's untainted default pool
  (`Standard_NC4as_T4_v3`, 1x NVIDIA T4, 4 vCPUs), created with
  `gpu_driver = "None"` so AKS installs no GPU software of its own,
- the **NVIDIA GPU operator** (host driver, container toolkit, device
  plugin, GPU-feature-discovery labels, DCGM, and the `nvidia` RuntimeClass —
  NVIDIA's [documented AKS path](https://learn.microsoft.com/azure/aks/nvidia-gpu-operator)), and
- **`gpu-mcp-server`**, installed from the in-tree chart, same as AWS.

Unlike EKS, AKS manages its own VNet, so there is no networking module — a
native `azurerm_kubernetes_cluster` resource is the whole cluster.

Azure has **no arm64 NVIDIA GPU VM sizes**, so this stack only ever validates
the linux/amd64 `gpu-mcp-server` image — arm64 coverage comes from the AWS
stack's `gpu_arm64` node group.

## Prerequisites

- **Terraform 1.15.6** — pinned in [`azure/.terraform-version`](./azure/.terraform-version).
- **Azure CLI (`az`)**, logged in (`az login`), plus a subscription ID (set
  `subscription_id` or the `ARM_SUBSCRIPTION_ID` env var — the azurerm v4
  provider requires one).
- **kubectl** (for poking at the cluster after apply; not required by
  Terraform itself).
- The **GPU vCPU quota** below.
- Registry access from the machine running Terraform: `terraform init`
  fetches the azurerm/kubernetes/helm providers from the public Terraform
  Registry, and the apply pulls the GPU operator chart from
  `helm.ngc.nvidia.com`.

## GPU vCPU quota — your apply fails without it

Fresh subscriptions have a GPU quota of **0**, per-region and per-VM-family.
The default T4 SKU draws on **"Standard NCASv3_T4 Family vCPUs"**
(`Standard_NC4as_T4_v3` = 4 vCPUs), so request **≥ 4** in your `location`
before applying (portal → **Subscriptions → Usage + quotas**). Verify:

```bash
az vm list-usage --location eastus --query "[?contains(name.value, 'NCASv3_T4')]" -o table
```

## Usage

```bash
cd infra/terraform/azure

export ARM_SUBSCRIPTION_ID=<your-subscription-id>
cp terraform.tfvars.example terraform.tfvars   # optional: all vars have defaults

terraform init
terraform apply

# Point kubectl at the new cluster (also emitted as the `configure_kubectl` output)
az aks get-credentials --resource-group gpu-mcp-server-e2e-rg \
  --name gpu-mcp-server-e2e --overwrite-existing

kubectl get nodes -L nvidia.com/gpu.present
kubectl -n gpu-mcp get pods -o wide
```

## Common overrides

| Variable | Default | Notes |
|---|---|---|
| `location` | `eastus` | A region with T4 capacity + your quota. |
| `resource_group_name` | `gpu-mcp-server-e2e-rg` | |
| `gpu_vm_size` | `Standard_NC4as_T4_v3` (T4, 4 vCPUs) | |
| `gpu_node_count` | `1` | Fixed-size pool (no autoscaler). |
| `kubernetes_version` | `1.33` | Current in-support minor; check for newer supported minors before relying on this default. |
| `gpu_operator_chart_version` | `v26.3.2` | NVIDIA GPU operator chart. |
| `mcp_image_repository` / `mcp_image_tag` | chart defaults | Pin the exact `gpu-mcp-server` image under test. |

## Cost & teardown

The AKS control plane is Free-tier ($0); you pay for the GPU VM (~$0.53/hr
for the default T4 — a rough list-price ballpark, verify for your region)
plus a managed disk — call it **~$0.55/hr**, cheaper than EKS (no paid
control plane, no NAT gateway). **Destroy when done:**

```bash
terraform destroy   # removes the resource group and everything in it
# leftovers, if a destroy is ever interrupted:
az resource list --tag Project=gpu-mcp-server -o table
```

## How the cluster satisfies the chart

| Chart requirement | Provided by |
|---|---|
| `nodeSelector: nvidia.com/gpu.present=true` (set by Terraform via Helm `set`) | GPU-feature-discovery (operator) labels the node |
| `nvidia.runtimeClassName: nvidia` (set by Terraform) | operator's container toolkit configures containerd's `nvidia` runtime and creates the RuntimeClass |
| driver + NVML libraries injected into the unprivileged container | operator's driver DaemonSet (`gpu_driver = "None"` on the node pool skips AKS's own driver install) |
| `tolerations: nvidia.com/gpu` | no-op — the GPU node is the untainted default pool |

The GPU node is the untainted default pool, so the operator's controllers and
CoreDNS co-locate with `gpu-mcp-server`.

---

# GCP GKE GPU test cluster

Sibling to the AWS and Azure stacks. One `terraform apply` provisions
everything, no manual steps:

- a **zonal** GKE cluster (native `google_container_cluster`, VPC-native, the
  default VPC, `deletion_protection = false` so `terraform destroy` actually
  works),
- **one** fixed-size, on-demand GPU node pool (`google_container_node_pool`)
  as the cluster's untainted GPU pool (`n1-standard-4`, 1x NVIDIA T4, 4
  vCPUs, Ubuntu node image), created with `gpu_driver_version = "INSTALLATION_DISABLED"` so
  GKE installs no GPU software of its own — `remove_default_node_pool = true`
  drops the generic default pool it comes with, leaving the GPU pool as the
  cluster's only pool,
- the **NVIDIA GPU operator** (host driver, container toolkit, device
  plugin, GPU-feature-discovery labels, DCGM, and the `nvidia` RuntimeClass —
  the same driver split as the Azure stack), and
- **`gpu-mcp-server`**, installed from the in-tree chart, same as AWS and
  Azure.

Like AKS, GKE manages its own networking, so there is no VPC module — a
native `google_container_cluster` resource plus a `google_container_node_pool`
is the whole cluster.

GCP has **no arm64 NVIDIA GPU machine type**, so this stack only ever
validates the linux/amd64 `gpu-mcp-server` image — arm64 coverage comes from
the AWS stack's `gpu_arm64` node group.

## Prerequisites

- **Terraform 1.15.6** — pinned in [`gcp/.terraform-version`](./gcp/.terraform-version).
- **gcloud CLI**, authenticated with Application Default Credentials
  (`gcloud auth application-default login`) for the google provider, plus a
  target project (`project_id` is required, no default).
- The **`container.googleapis.com`** and **`compute.googleapis.com`** APIs
  enabled on the project (`gcloud services enable container.googleapis.com
  compute.googleapis.com`).
- **kubectl**, and the **`gke-gcloud-auth-plugin`** — only needed for the
  post-apply `gcloud container clusters get-credentials`; the
  Kubernetes/Helm providers authenticate against the cluster with a
  short-lived google access token, so no kubeconfig is required to apply.
- The **GPU quota** below.
- Registry access from the machine running Terraform: `terraform init`
  fetches the google/kubernetes/helm providers from the public Terraform
  Registry, and the apply pulls the GPU operator chart from
  `helm.ngc.nvidia.com`.

## GPU quota — your apply fails without it

Fresh GCP projects have a GPU quota of **0**. Two separate quotas gate the
default T4 node pool, and both need to be **at least 1**:

- the regional quota, **"NVIDIA T4 GPUs"**, in your `region`; and
- the global quota, **"GPUs (all regions)"**.

Request both before applying (Console → **IAM & Admin → Quotas**, filter
"GPU"). Verify the regional one with:

```bash
gcloud compute regions describe us-central1 --format="value(quotas)"
```

## Usage

```bash
cd infra/terraform/gcp

gcloud auth application-default login
gcloud config set project <your-project-id>

cp terraform.tfvars.example terraform.tfvars   # set project_id at minimum
terraform init
terraform apply

# Point kubectl at the new cluster (also emitted as the `configure_kubectl` output)
gcloud container clusters get-credentials gpu-mcp-server-e2e \
  --zone us-central1-a --project <your-project-id>

kubectl get nodes -L nvidia.com/gpu.present
kubectl -n gpu-mcp get pods -o wide
```

## Common overrides

| Variable | Default | Notes |
|---|---|---|
| `project_id` | *(required)* | GCP project to deploy into — no default. |
| `region` / `zone` | `us-central1` / `us-central1-a` | Zonal cluster; pick a zone with T4 capacity + your quota. |
| `gpu_machine_type` | `n1-standard-4` (T4-compatible, 4 vCPUs) | |
| `gpu_accelerator_type` | `nvidia-tesla-t4` | |
| `gpu_node_count` | `1` | Fixed-size pool (no autoscaler). |
| `kubernetes_version` | `1.33` | Used as GKE's `min_master_version`. |
| `gpu_operator_chart_version` | `v26.3.2` | NVIDIA GPU operator chart. |
| `mcp_image_repository` / `mcp_image_tag` | chart defaults | Pin the exact `gpu-mcp-server` image under test. |

## Cost & teardown

You pay for the GKE cluster management fee (~$0.10/hr) plus the GPU node —
`n1-standard-4` (~$0.19/hr) and 1x T4 (~$0.35/hr) — rough list-price
ballparks (us-central1, USD; verify for your region), call it **~$0.65/hr
(~$16/day)**. **Destroy when done:**

```bash
terraform destroy   # works because deletion_protection = false on the cluster
# leftovers, if a destroy is ever interrupted:
gcloud compute instances list --filter="labels.project=gpu-mcp-server"
```

## How the cluster satisfies the chart

| Chart requirement | Provided by |
|---|---|
| `nodeSelector: nvidia.com/gpu.present=true` (set by Terraform via Helm `set`) | GPU-feature-discovery (operator) labels the node |
| `nvidia.runtimeClassName: nvidia` (set by Terraform) | operator's container toolkit configures containerd's `nvidia` runtime and creates the RuntimeClass |
| driver + NVML libraries injected into the unprivileged container | operator's driver DaemonSet (`gpu_driver_version = "INSTALLATION_DISABLED"` on the node pool skips GKE's own driver install) |
| `tolerations: nvidia.com/gpu` | no-op — the GPU node pool is untainted, and is the cluster's only pool after `remove_default_node_pool` |

The GPU node pool is untainted, so the operator's controllers and CoreDNS
co-locate with `gpu-mcp-server`.

## GKE-specific operator settings

GKE needs three things the EKS and AKS stacks don't, plus the same driver
split as Azure:

- **Driver/toolkit split.** The node pool sets
  `gpu_driver_version = "INSTALLATION_DISABLED"` and runs the Ubuntu node
  image, so GKE installs no GPU software — the NVIDIA GPU operator owns the
  whole stack (`driver.enabled=true`), the same split as the Azure stack.

- **Priority-class ResourceQuota.** GKE restricts the `system-node-critical`
  and `system-cluster-critical` priority classes to `kube-system` via an
  admission ResourceQuota. The operator's pods request those priority classes,
  so in any other namespace they're rejected ("insufficient quota to match
  these scopes") and nothing schedules. The stack therefore creates the
  `gpu-operator` namespace with its own ResourceQuota whose `scopeSelector`
  permits those priority-class scopes. AWS and AKS don't enforce this.

- **Disable GKE's managed device plugin.** The GPU node pool carries the label
  `gke-no-default-nvidia-gpu-device-plugin=true`. Without it GKE deploys its
  own NVIDIA device plugin into `kube-system`, which conflicts with the
  operator's device plugin and validator and wedges the operator rollout.

- **Toolkit `RUNTIME_CONFIG_SOURCE=file` (GKE 1.33+ / containerd 2.0 CNI
  fix).** By default the operator's container toolkit configures containerd in
  "command" mode: it runs `containerd config dump` and writes the full result
  back to `/etc/containerd/config.toml`, baking in the upstream default
  `bin_dir = "/opt/cni/bin"` and dropping GKE's real CNI bin dir
  (`/home/kubernetes/bin`). Every pod sandbox on the node then fails with
  `failed to find plugin "loopback"/"ptp" in path [/opt/cni/bin]`, so
  `nvidia-operator-validator` (and the device-plugin/GFD/dcgm behind it) never
  start. Setting `toolkit.env` `RUNTIME_CONFIG_SOURCE=file` (with
  `CONTAINERD_CONFIG=/etc/containerd/config.toml`,
  `CONTAINERD_SOCKET=/run/containerd/containerd.sock`) makes the toolkit edit
  the existing `config.toml` in place — adding only the `nvidia` runtime and
  leaving GKE's CNI settings intact. Reference:
  NVIDIA/nvidia-container-toolkit#1222.

Apply this on a **fresh** cluster. Don't try to fix an already-broken node
with `helm upgrade` — helm's pre-upgrade hook pod can't get networking once
the CNI config is clobbered, so the upgrade deadlocks; recreate the node (clean
containerd config) instead.

---

# Validating a deployed cluster

Bringing a stack up installs `gpu-mcp-server` but doesn't test it. Validation is
a **manual** procedure — there is intentionally no apply/verify/destroy
automation, because a real GPU cluster costs money and needs quota, so
`terraform apply` and `terraform destroy` stay deliberate human steps. The full
walkthrough (with the exact commands and the interactive MCP-client checks) is
in **[`infra/e2e/README.md`](../e2e/README.md)**. In short:

1. `kubectl get nodes` / `kubectl get pods -n gpu-mcp` — every GPU node `Ready`
   with a `1/1 Running` pod (two on AWS: amd64 + arm64; one on Azure and GCP,
   single amd64 pod).
2. `nvidia-smi` on each pod for ground-truth GPU metrics.
3. Connect an MCP client (e.g. Claude) to each pod and confirm the server's
   `list_gpus` / `get_gpu_metrics` / `gpu_summary` match the `nvidia-smi`
   numbers, on both architectures.
4. Apply the [`infra/e2e/gpu-burn-*.yaml`](../e2e) load jobs and confirm
   utilization climbs to ~100% on both GPUs, tracked live by the server.

Requirements: **kubectl**, the relevant cloud CLI (to write the kubeconfig), and
an MCP client for the interactive checks — plus, as above, the GPU quota for
whichever cloud you're targeting.
