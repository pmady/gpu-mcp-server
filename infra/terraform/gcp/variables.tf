###############################################################################
# Cluster / GCP
###############################################################################

variable "project_id" {
  description = "GCP project ID to deploy into. Required — no default."
  type        = string
}

variable "region" {
  description = "GCP region for the provider default and regional resources."
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP zone the cluster runs in. The cluster is ZONAL (single zone), not regional, to keep the test cluster cheap (a regional cluster runs 3 control-plane replicas)."
  type        = string
  default     = "us-central1-a"
}

variable "cluster_name" {
  description = "Name of the GKE cluster."
  type        = string
  default     = "gpu-mcp-server-e2e"
}

variable "kubernetes_version" {
  description = "GKE Kubernetes version (<major>.<minor>; GKE selects the patch). Pick a version still supported in your release channel."
  type        = string
  default     = "1.33"
}

variable "labels" {
  description = "Extra labels merged into the default labels applied to every resource (e.g. an owner or expiry date)."
  type        = map(string)
  default     = {}
}

###############################################################################
# GPU node pool
###############################################################################

variable "gpu_machine_type" {
  description = "Machine type for the (single) GPU node pool. Default is the cheapest current-gen T4 pairing: 4 vCPUs. T4 GPUs attach to N1 machine types only."
  type        = string
  default     = "n1-standard-4"

  validation {
    condition     = can(regex("^n1-", var.gpu_machine_type))
    error_message = "gpu_machine_type must be an N1 machine type (e.g. n1-standard-4) — T4 requires N1."
  }
}

variable "gpu_accelerator_type" {
  description = "GPU accelerator type attached to each GPU node. Default is the cheapest current-gen NVIDIA GPU."
  type        = string
  default     = "nvidia-tesla-t4"

  validation {
    condition     = can(regex("^nvidia-", var.gpu_accelerator_type))
    error_message = "gpu_accelerator_type must be an NVIDIA accelerator, e.g. nvidia-tesla-t4."
  }
}

variable "gpu_node_count" {
  description = "Number of GPU nodes. Fixed-size pool (no autoscaler). Kept at 1 for predictable, low-cost end-to-end testing."
  type        = number
  default     = 1
}

variable "gpu_node_disk_size" {
  description = "Boot disk size (GiB) per GPU node. GPU container images plus the driver/CUDA layers are large, so this is generous by default."
  type        = number
  default     = 100
}

###############################################################################
# Add-ons: NVIDIA GPU operator and the in-tree gpu-mcp-server chart
###############################################################################

variable "gpu_operator_chart_version" {
  description = "NVIDIA GPU operator Helm chart version (repo https://helm.ngc.nvidia.com/nvidia)."
  type        = string
  default     = "v26.3.2"
}

variable "mcp_namespace" {
  description = "Namespace gpu-mcp-server is installed into."
  type        = string
  default     = "gpu-mcp"
}

variable "mcp_release_name" {
  description = "Helm release name for the in-tree gpu-mcp-server chart. Also determines the in-cluster service name / HTTP endpoint."
  type        = string
  default     = "gpu-mcp-server"
}

variable "mcp_image_repository" {
  description = "gpu-mcp-server container image repository. Empty string uses the chart default (ghcr.io/pmady/gpu-mcp-server)."
  type        = string
  default     = ""
}

variable "mcp_image_tag" {
  description = "gpu-mcp-server container image tag to deploy. Empty string uses the chart default (the chart appVersion)."
  type        = string
  default     = ""
}

variable "mcp_runtime_class_name" {
  description = "RuntimeClass requested by the gpu-mcp-server pod. The GPU operator's container toolkit provisions the `nvidia` RuntimeClass, which configures containerd's `nvidia` runtime handler to inject driver libraries so NVML works in the unprivileged container. null skips setting it (falls back to the chart default, which is unset)."
  type        = string
  default     = "nvidia"
}

variable "helm_timeout" {
  description = "Per-release Helm wait timeout in seconds. Generous because on GKE the GPU operator builds/installs the driver from source, which can take 10-15+ minutes."
  type        = number
  default     = 1800
}
