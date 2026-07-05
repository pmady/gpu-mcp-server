# NVIDIA GPU operator.
#
# The GPU node pool sets gpu_driver_version = "INSTALLATION_DISABLED" and runs the Ubuntu
# node image (see main.tf), so GKE installs NO GPU software. The operator
# therefore owns the entire stack:
#   - the NVIDIA host driver (driver.enabled = true),
#   - the NVIDIA container toolkit, which configures containerd's `nvidia`
#     runtime handler the `nvidia` RuntimeClass points at (toolkit.enabled = true),
#   - the NVIDIA k8s device plugin (advertises nvidia.com/gpu),
#   - node-feature-discovery + GPU-feature-discovery, which apply the
#     `nvidia.com/gpu.present=true` node label gpu-mcp-server's nodeSelector
#     targets,
#   - DCGM / dcgm-exporter,
#   - the `nvidia` RuntimeClass referenced by the gpu-mcp-server pod template.
#
# This is the same driver/operator split as the Azure sibling, and NVIDIA's
# documented approach for GKE (skip the GKE driver, run the operator):
# https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/google-gke.html

# GKE-specific: GKE restricts the system-node-critical / system-cluster-critical
# priority classes to kube-system via an admission ResourceQuota, so the GPU
# operator's pods (which request those priority classes) are rejected in any
# other namespace with "insufficient quota to match these scopes" — the operator
# never schedules a single pod and every helm install times out. Create the
# operator namespace ourselves with a ResourceQuota that permits those
# priority-class scopes so the pods can be admitted. The AWS/Azure siblings
# don't need this (only GKE enforces the restriction).
resource "kubernetes_namespace_v1" "gpu_operator" {
  metadata {
    name = "gpu-operator"
  }

  depends_on = [google_container_node_pool.gpu]
}

resource "kubernetes_resource_quota_v1" "gpu_operator_priority" {
  metadata {
    name      = "gpu-operator-critical"
    namespace = kubernetes_namespace_v1.gpu_operator.metadata[0].name
  }

  spec {
    hard = {
      pods = "100"
    }

    scope_selector {
      match_expression {
        scope_name = "PriorityClass"
        operator   = "In"
        values     = ["system-node-critical", "system-cluster-critical"]
      }
    }
  }
}

resource "helm_release" "gpu_operator" {
  name      = "gpu-operator"
  namespace = kubernetes_namespace_v1.gpu_operator.metadata[0].name

  # Namespace is created above (with the priority-class ResourceQuota), so the
  # helm release must not try to create it.
  create_namespace = false

  repository = "https://helm.ngc.nvidia.com/nvidia"
  chart      = "gpu-operator"
  version    = var.gpu_operator_chart_version

  # On GKE the operator (not the node image) builds and installs the driver and
  # toolkit — driver.enabled=true. Do NOT set hostPaths.driverInstallDir /
  # toolkit.installDir here: those point the driver/toolkit at GKE's
  # /home/kubernetes/bin/nvidia path, which is only correct for the GKE-managed
  # driver flow (driver.enabled=false). With driver.enabled=true the operator
  # builds the driver at its own default path, so overriding the install dir
  # makes the toolkit's driver-validation look in the wrong place and loop
  # forever, stalling the whole stack.
  set = [
    {
      name  = "driver.enabled"
      value = "true"
    },
    {
      name  = "toolkit.enabled"
      value = "true"
    },

    # Inject GPUs via the Container Device Interface (CDI) instead of the legacy
    # containerd `nvidia` runtime handler. On GKE 1.33+ (containerd 2.0) the
    # legacy handler rewrites /etc/containerd/config.toml and drops GKE's CNI
    # bin_dir, breaking pod networking ("plugin ptp not found in /opt/cni/bin").
    # CDI is NVIDIA's documented GKE path and avoids that rewrite. See:
    # https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/google-gke.html
    {
      name  = "cdi.enabled"
      value = "true"
    },
    {
      name  = "cdi.default"
      value = "true"
    },
  ]

  # Driver build + device-plugin/GFD rollout and node labelling can take several
  # minutes after the node joins.
  wait    = true
  timeout = var.helm_timeout

  depends_on = [
    google_container_node_pool.gpu,
    kubernetes_resource_quota_v1.gpu_operator_priority,
  ]
}
