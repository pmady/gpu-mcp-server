locals {
  labels = merge(
    {
      project    = "gpu-mcp-server"
      component  = "gpu-e2e-test"
      managed-by = "terraform"
      stack      = "infra-terraform-gcp"
    },
    var.labels,
  )
}

###############################################################################
# GKE control plane + single GPU node pool
#
# Hand-rolled on the native `google_container_cluster` resource plus a
# separate `google_container_node_pool`, with `remove_default_node_pool` so
# the GPU pool is the ONLY pool — the cheapest single-pool layout, mirroring
# the EKS/AKS siblings. The single GPU node runs the whole stack (the GPU
# operator, CoreDNS, and gpu-mcp-server) and is left untainted.
#
# GCP has no arm64 NVIDIA GPU machine type, so unlike the AWS sibling (dual
# amd64+arm64 node groups) this stack only ever validates the linux/amd64
# image.
###############################################################################

resource "google_container_cluster" "this" {
  name     = var.cluster_name
  location = var.zone # zonal cluster, not regional — one control plane, not three

  min_master_version = var.kubernetes_version

  remove_default_node_pool = true
  initial_node_count       = 1

  # IMPORTANT: the provider default for deletion_protection is true, which
  # blocks `terraform destroy`. Must be false for a throwaway test cluster.
  deletion_protection = false

  # Use the project's default VPC — throwaway test cluster, no dedicated networking.
  network    = "default"
  subnetwork = "default"

  networking_mode = "VPC_NATIVE"
  ip_allocation_policy {} # VPC-native with GKE-managed secondary ranges

  resource_labels = local.labels
}

resource "google_container_node_pool" "gpu" {
  name       = "gpu"
  cluster    = google_container_cluster.this.name
  location   = var.zone
  node_count = var.gpu_node_count

  node_config {
    machine_type = var.gpu_machine_type

    # Ubuntu, not COS — the GPU operator's driver container needs it.
    image_type   = "UBUNTU_CONTAINERD"
    disk_size_gb = var.gpu_node_disk_size

    oauth_scopes = ["https://www.googleapis.com/auth/cloud-platform"]

    labels = {
      "gpu-mcp-server.io/pool" = "gpu"

      # Disable GKE's own managed NVIDIA GPU device plugin so it doesn't run
      # alongside (and conflict with) the one the NVIDIA GPU operator installs.
      # Required whenever the GPU operator manages the stack on GKE — without it
      # GKE deploys its device plugin into kube-system, which competes with the
      # operator's device plugin / validator and wedges the operator's rollout.
      # https://docs.cloud.google.com/kubernetes-engine/docs/how-to/gpu-operator
      "gke-no-default-nvidia-gpu-device-plugin" = "true"
    }

    guest_accelerator {
      type  = var.gpu_accelerator_type
      count = 1

      gpu_driver_installation_config {
        # Do NOT let GKE install the driver; the GPU operator owns it
        # (mirrors the Azure sibling's gpu_driver = "None"). See gpu_operator.tf.
        gpu_driver_version = "INSTALLATION_DISABLED"
      }
    }

    # NOTE: intentionally the only (untainted) pool. The GPU operator
    # controllers and CoreDNS schedule on the GPU node alongside
    # gpu-mcp-server. The DaemonSet tolerates `nvidia.com/gpu` regardless, so
    # tainting later is safe for gpu-mcp-server but would strand system pods
    # unless you add a separate CPU pool.
  }

  management {
    auto_repair  = true
    auto_upgrade = false
  }
}
