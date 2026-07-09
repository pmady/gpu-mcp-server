# NVIDIA GPU operator.
#
# The AL2023 NVIDIA AMIs (var.gpu_amd64_ami_type / var.gpu_arm64_ami_type
# defaults) already ship the host driver, CUDA user-mode driver and the NVIDIA
# container toolkit, and containerd is preconfigured with the `nvidia`
# runtime — for both x86_64 and ARM_64. So we disable the operator's driver
# and toolkit components and let it provide only what the AMIs do not:
#   - the NVIDIA k8s device plugin (advertises nvidia.com/gpu),
#   - node-feature-discovery + GPU-feature-discovery, which apply the
#     `nvidia.com/gpu.present=true` node label the gpu-mcp-server DaemonSet's
#     nodeSelector targets,
#   - DCGM / dcgm-exporter,
#   - the `nvidia` RuntimeClass referenced by the gpu-mcp-server pod template.
#
# These components are multi-arch, so a single operator release covers both
# the gpu_amd64 and gpu_arm64 node groups.
#
# If you switch to a non-NVIDIA AMI (e.g. plain AL2023), set driver.enabled=true
# so the operator installs the driver itself.
resource "helm_release" "gpu_operator" {
  name             = "gpu-operator"
  namespace        = "gpu-operator"
  create_namespace = true

  repository = "https://helm.ngc.nvidia.com/nvidia"
  chart      = "gpu-operator"
  version    = var.gpu_operator_chart_version

  set = [
    {
      name  = "driver.enabled"
      value = "false"
    },
    {
      name  = "toolkit.enabled"
      value = "false"
    },
  ]

  # Device-plugin/GFD rollout and node labelling can take a few minutes after
  # the node joins.
  wait    = true
  timeout = var.helm_timeout

  depends_on = [module.eks]
}
