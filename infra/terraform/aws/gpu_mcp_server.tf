locals {
  # Only override chart values that were explicitly set; otherwise fall back
  # to the in-tree chart defaults (nodeSelector {}, runtimeClassName: "",
  # image ghcr.io/pmady/gpu-mcp-server:<appVersion>).
  mcp_set = concat(
    [
      # The chart's nodeSelector default is {} (unset), so the DaemonSet would
      # otherwise schedule on every node. Pin it to the GFD label the GPU
      # operator applies. type = "string" keeps the label value the literal
      # string "true" — Kubernetes label values must be strings.
      {
        name  = "nodeSelector.nvidia\\.com/gpu\\.present"
        value = "true"
        type  = "string"
      },
    ],
    var.mcp_image_repository != "" ? [{ name = "image.repository", value = var.mcp_image_repository, type = "auto" }] : [],
    var.mcp_image_tag != "" ? [{ name = "image.tag", value = var.mcp_image_tag, type = "auto" }] : [],
    # The NVIDIA runtime (requested via runtimeClassName) injects the driver
    # libraries into the otherwise-unprivileged container, which is how NVML
    # reaches the host driver. This is the RuntimeClass the GPU operator
    # creates.
    var.mcp_runtime_class_name != null ? [{ name = "nvidia.runtimeClassName", value = var.mcp_runtime_class_name, type = "auto" }] : [],
  )
}

# gpu-mcp-server, installed FROM the in-tree Helm chart so the test cluster
# always runs the local version of the server rather than a published release.
resource "helm_release" "gpu_mcp_server" {
  name             = var.mcp_release_name
  namespace        = var.mcp_namespace
  create_namespace = true

  chart = "${path.module}/../../../deploy/helm/gpu-mcp-server"

  set = local.mcp_set

  wait    = true
  timeout = var.helm_timeout

  # The DaemonSet's nodeSelector targets the `nvidia.com/gpu.present` label
  # and its pod requests the `nvidia` RuntimeClass — both are provisioned by
  # the GPU operator, so it must be in place first or the pod stays Pending
  # and the release times out.
  depends_on = [
    module.eks,
    helm_release.gpu_operator,
  ]
}
