output "cluster_name" {
  description = "GKE cluster name."
  value       = google_container_cluster.this.name
}

output "location" {
  description = "GCP zone the cluster runs in."
  value       = var.zone
}

output "project" {
  description = "GCP project the cluster runs in."
  value       = var.project_id
}

output "cluster_endpoint" {
  description = "GKE Kubernetes API server endpoint."
  value       = "https://${google_container_cluster.this.endpoint}"
}

output "configure_kubectl" {
  description = "Command to write a kubeconfig entry for the new cluster."
  value       = "gcloud container clusters get-credentials ${var.cluster_name} --zone ${var.zone} --project ${var.project_id}"
}

output "mcp_namespace" {
  description = "Namespace gpu-mcp-server is installed in."
  value       = var.mcp_namespace
}

output "mcp_release_name" {
  description = "Helm release name for gpu-mcp-server."
  value       = var.mcp_release_name
}

output "mcp_http_endpoint" {
  description = "In-cluster MCP Streamable HTTP endpoint (MCP at /, health at /healthz)."
  value       = "http://${var.mcp_release_name}.${var.mcp_namespace}.svc.cluster.local:8080"
}
