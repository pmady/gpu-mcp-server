# NVIDIA GPU operator. Owns the full driver + toolkit stack; the GPU node pool
# sets INSTALLATION_DISABLED (see main.tf). Same split as the Azure sibling.

# GKE restricts the system-node-critical/system-cluster-critical priority classes;
# this quota lets the operator's pods be admitted. see infra/terraform/README.md
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

  # Namespace created above with the priority-class quota.
  create_namespace = false

  repository = "https://helm.ngc.nvidia.com/nvidia"
  chart      = "gpu-operator"
  version    = var.gpu_operator_chart_version

  # Operator builds the driver and toolkit; don't override the install dirs. see infra/terraform/README.md
  set = [
    {
      name  = "driver.enabled"
      value = "true"
    },
    {
      name  = "toolkit.enabled"
      value = "true"
    },

    # Force the toolkit to edit containerd's config in place so GKE's CNI bin_dir
    # survives (GKE 1.33+, containerd 2.0). see infra/terraform/README.md
    {
      name  = "toolkit.env[0].name"
      value = "CONTAINERD_CONFIG"
    },
    {
      name  = "toolkit.env[0].value"
      value = "/etc/containerd/config.toml"
    },
    {
      name  = "toolkit.env[1].name"
      value = "CONTAINERD_SOCKET"
    },
    {
      name  = "toolkit.env[1].value"
      value = "/run/containerd/containerd.sock"
    },
    {
      name  = "toolkit.env[2].name"
      value = "RUNTIME_CONFIG_SOURCE"
    },
    {
      name  = "toolkit.env[2].value"
      value = "file"
    },

    # Inject GPUs via CDI (NVIDIA's recommended mode on GKE).
    {
      name  = "cdi.enabled"
      value = "true"
    },
    {
      name  = "cdi.default"
      value = "true"
    },
  ]

  # Driver build + device-plugin rollout can take several minutes after the node joins.
  wait    = true
  timeout = var.helm_timeout

  depends_on = [
    google_container_node_pool.gpu,
    kubernetes_resource_quota_v1.gpu_operator_priority,
  ]
}
