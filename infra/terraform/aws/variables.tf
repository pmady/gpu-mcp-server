###############################################################################
# Cluster / AWS
###############################################################################

variable "region" {
  description = "AWS region to deploy the test cluster into. Defaults to us-west-2, where the quota math in gpu_amd64_instance_type / gpu_arm64_instance_type below is documented."
  type        = string
  default     = "us-west-2"
}

variable "cluster_name" {
  description = "Name of the EKS cluster and the prefix used for the VPC and related resources."
  type        = string
  default     = "gpu-mcp-server-e2e"
}

variable "kubernetes_version" {
  description = "EKS Kubernetes control plane version (<major>.<minor>). Latest on EKS is 1.36; pick a version still in standard support."
  type        = string
  default     = "1.35"
}

variable "vpc_cidr" {
  description = "IPv4 CIDR block."
  type        = string
  default     = "10.0.0.0/16"
}

variable "tags" {
  description = "Extra tags merged into the default tags applied to every resource (e.g. an owner or expiry date)."
  type        = map(string)
  default     = {}
}

###############################################################################
# GPU node pools (dual-arch: amd64 + arm64)
###############################################################################

# Both node groups are ON_DEMAND, so they count against the AWS quota
# "Running On-Demand G and VT instances" (code L-DB2E81BA), which is measured
# in vCPUs per region. The instance type defaults below are deliberately
# 4 vCPUs each (g4dn.xlarge + g5g.xlarge = 8 vCPUs total) so the whole
# dual-arch cluster fits inside an 8-vCPU quota, which is what a fresh-ish
# account typically starts with in us-west-2. The instance types are
# variables precisely so you can trade vCPU footprint per architecture if
# your quota is higher (or lower).

variable "gpu_amd64_instance_type" {
  description = "GPU instance type for the amd64 node pool. Default g4dn.xlarge: 1x NVIDIA T4, 4 vCPUs."
  type        = string
  default     = "g4dn.xlarge"
}

variable "gpu_amd64_ami_type" {
  description = "EKS-optimized Amazon Linux 2023 with NVIDIA host driver, x86_64."
  type        = string
  default     = "AL2023_x86_64_NVIDIA"

  validation {
    condition     = can(regex("NVIDIA|GPU", var.gpu_amd64_ami_type))
    error_message = "gpu_amd64_ami_type must be an NVIDIA/GPU accelerated EKS AMI type (e.g. AL2023_x86_64_NVIDIA, BOTTLEROCKET_x86_64_NVIDIA, AL2_x86_64_GPU)."
  }
}

variable "gpu_amd64_node_count" {
  description = "Number of amd64 GPU nodes. Fixed-size on-demand pool (min = max = desired). Set to 0 to omit this node group entirely. Kept at 1 for predictable, low-cost e2e testing."
  type        = number
  default     = 1
}

variable "gpu_arm64_instance_type" {
  description = "GPU instance type for the arm64 node pool. Default g5g.xlarge: Graviton2 + 1x NVIDIA T4G, 4 vCPUs."
  type        = string
  default     = "g5g.xlarge"
}

variable "gpu_arm64_ami_type" {
  description = "EKS-optimized Amazon Linux 2023 with NVIDIA host driver, ARM_64 (Graviton)."
  type        = string
  default     = "AL2023_ARM_64_NVIDIA"

  validation {
    condition     = can(regex("NVIDIA|GPU", var.gpu_arm64_ami_type))
    error_message = "gpu_arm64_ami_type must be an NVIDIA/GPU accelerated EKS AMI type (e.g. AL2023_ARM_64_NVIDIA, BOTTLEROCKET_ARM_64_NVIDIA)."
  }
}

variable "gpu_arm64_node_count" {
  description = "Number of arm64 GPU nodes. Fixed-size on-demand pool (min = max = desired). Set to 0 to omit this node group entirely. Kept at 1 for predictable, low-cost e2e testing."
  type        = number
  default     = 1
}

variable "gpu_node_disk_size" {
  description = "Root EBS volume size (GiB) per GPU node (both pools). GPU container images plus the driver/CUDA layers are large, so this is generous by default."
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
  description = "Override the gpu-mcp-server container image repository. Empty string uses the chart default (ghcr.io/pmady/gpu-mcp-server)."
  type        = string
  default     = ""
}

variable "mcp_image_tag" {
  description = "Override the gpu-mcp-server container image tag. Empty string uses the chart default (the chart's appVersion)."
  type        = string
  default     = ""
}

variable "mcp_runtime_class_name" {
  description = "runtimeClassName set on the gpu-mcp-server pod. Defaults to 'nvidia', the RuntimeClass the GPU operator creates. Set to null to skip setting it and fall back to the chart default (empty, i.e. no runtimeClassName)."
  type        = string
  default     = "nvidia"
}

variable "helm_timeout" {
  description = "Per-release Helm wait timeout in seconds. Generous because GPU driver/device-plugin rollout and node labelling can take several minutes."
  type        = number
  default     = 900
}
