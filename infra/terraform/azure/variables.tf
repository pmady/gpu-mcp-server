###############################################################################
# Cluster / Azure
###############################################################################

variable "subscription_id" {
  description = "Azure subscription ID to deploy into. Leave null to use the ARM_SUBSCRIPTION_ID environment variable."
  type        = string
  default     = null
}

variable "location" {
  description = "Azure region. Pick one with NCasT4_v3 (T4) capacity where you hold the GPU vCPU quota."
  type        = string
  default     = "eastus"
}

variable "resource_group_name" {
  description = "Resource group created to hold the cluster. Destroyed by `terraform destroy`."
  type        = string
  default     = "gpu-mcp-server-e2e-rg"
}

variable "cluster_name" {
  description = "Name of the AKS cluster (also used as the DNS prefix)."
  type        = string
  default     = "gpu-mcp-server-e2e"
}

variable "kubernetes_version" {
  description = "AKS Kubernetes version (<major>.<minor>; AKS selects the latest patch). Defaults to a current in-support minor validated end-to-end; 1.34/1.35 are also in support. Never default to a near-EOL minor."
  type        = string
  default     = "1.33"
}

variable "tags" {
  description = "Extra tags merged into the default tags applied to every resource (e.g. an owner or expiry date)."
  type        = map(string)
  default     = {}
}

###############################################################################
# GPU node pool
###############################################################################

variable "gpu_vm_size" {
  description = "GPU VM size for the (single) node pool. Default is the cheapest current-gen T4. Newer/bigger: Standard_NC24ads_A100_v4 (A100), Standard_NC24ads_L40S_v4 (L40S)."
  type        = string
  default     = "Standard_NC4as_T4_v3"

  validation {
    condition     = can(regex("^Standard_N", var.gpu_vm_size))
    error_message = "gpu_vm_size must be an Azure N-series (NVIDIA GPU) VM size, e.g. Standard_NC4as_T4_v3."
  }
}

variable "gpu_node_count" {
  description = "Number of GPU nodes. Fixed-size on-demand pool (no autoscaler). Kept at 1 for predictable, low-cost end-to-end testing."
  type        = number
  default     = 1
}

variable "gpu_node_disk_size" {
  description = "OS disk size (GiB) per GPU node. GPU container images plus the driver/CUDA layers are large, so this is generous by default."
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
  description = "Per-release Helm wait timeout in seconds. Generous because on AKS the GPU operator also builds/installs the driver, which can take several minutes."
  type        = number
  default     = 900
}
