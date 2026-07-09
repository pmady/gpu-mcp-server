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
# GKE control plane + single untainted GPU node pool (the only pool, runs the
# whole stack). GCP has no arm64 NVIDIA SKU, so linux/amd64 only.
###############################################################################

resource "google_container_cluster" "this" {
  name     = var.cluster_name
  location = var.zone # zonal cluster, not regional — one control plane, not three

  min_master_version = var.kubernetes_version

  remove_default_node_pool = true
  initial_node_count       = 1

  # Default is true and blocks `terraform destroy`; must be false for a throwaway cluster.
  deletion_protection = false

  # Project default VPC — throwaway test cluster, no dedicated networking.
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

      # Disable GKE's own GPU device plugin so it doesn't conflict with the operator's. see infra/terraform/README.md
      "gke-no-default-nvidia-gpu-device-plugin" = "true"
    }

    guest_accelerator {
      type  = var.gpu_accelerator_type
      count = 1

      gpu_driver_installation_config {
        # GPU operator owns the driver, not GKE. See gpu_operator.tf.
        gpu_driver_version = "INSTALLATION_DISABLED"
      }
    }

    # Intentionally untainted: operator controllers and CoreDNS co-locate here.
  }

  management {
    auto_repair  = true
    auto_upgrade = false
  }
}
