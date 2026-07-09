locals {
  # Only override chart values that were explicitly set; otherwise fall back to
  # the in-tree chart defaults (nodeSelector {}, nvidia.runtimeClassName: "",
  # image ghcr.io/pmady/gpu-mcp-server:<appVersion>). The nodeSelector entry is
  # always set: this stack needs the DaemonSet pinned to GPU nodes, and Helm
  # would otherwise treat the "true" value as a boolean, so it is typed as a
  # string explicitly (label values must be strings).
  mcp_set = concat(
    var.mcp_image_repository != "" ? [{ name = "image.repository", value = var.mcp_image_repository, type = "auto" }] : [],
    var.mcp_image_tag != "" ? [{ name = "image.tag", value = var.mcp_image_tag, type = "auto" }] : [],
    var.mcp_runtime_class_name != null ? [{ name = "nvidia.runtimeClassName", value = var.mcp_runtime_class_name, type = "auto" }] : [],
    [
      {
        name  = "nodeSelector.nvidia\\.com/gpu\\.present"
        value = "true"
        type  = "string"
      },
    ],
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

  # The DaemonSet pod selects the `nvidia.com/gpu.present` label and the
  # `nvidia` RuntimeClass that the GPU operator provisions — both must be in
  # place first or the pod stays Pending and the release times out.
  depends_on = [helm_release.gpu_operator]
}
