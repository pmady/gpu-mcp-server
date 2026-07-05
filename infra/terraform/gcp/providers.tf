# The google provider authenticates via Application Default Credentials (run
# `gcloud auth application-default login`) and needs the container.googleapis.com
# and compute.googleapis.com APIs enabled on the project.
provider "google" {
  project = var.project_id
  region  = var.region
}

# Supplies a short-lived OAuth access token for the Kubernetes/Helm providers.
data "google_client_config" "default" {}

# The Kubernetes and Helm providers authenticate to the freshly created GKE
# cluster using the same access token as the google provider, refreshed on
# every Terraform operation, so nothing needs to be written to ~/.kube/config
# for `apply` to work. gke-gcloud-auth-plugin is NOT required by Terraform
# itself — only for the post-apply `gcloud container clusters get-credentials`
# (see the configure_kubectl output).
provider "kubernetes" {
  host                   = "https://${google_container_cluster.this.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(google_container_cluster.this.master_auth[0].cluster_ca_certificate)
}

# Helm provider v3 takes its Kubernetes connection settings as an attribute
# object (`kubernetes = { ... }`) rather than a nested block — see the v2 -> v3
# upgrade guide.
provider "helm" {
  kubernetes = {
    host                   = "https://${google_container_cluster.this.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(google_container_cluster.this.master_auth[0].cluster_ca_certificate)
  }
}
