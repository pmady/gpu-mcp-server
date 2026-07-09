output "cluster_name" {
  description = "EKS cluster name."
  value       = module.eks.cluster_name
}

output "region" {
  description = "AWS region the cluster runs in."
  value       = var.region
}

output "cluster_endpoint" {
  description = "EKS Kubernetes API server endpoint."
  value       = module.eks.cluster_endpoint
}

output "configure_kubectl" {
  description = "Command to write a kubeconfig entry for the new cluster."
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}

output "mcp_namespace" {
  description = "Namespace gpu-mcp-server is installed in."
  value       = var.mcp_namespace
}

output "mcp_release_name" {
  description = "Helm release name of the in-cluster gpu-mcp-server DaemonSet."
  value       = var.mcp_release_name
}

output "mcp_http_endpoint" {
  description = "In-cluster MCP Streamable HTTP endpoint (path /, health at /healthz)."
  value       = "http://${var.mcp_release_name}.${var.mcp_namespace}.svc.cluster.local:8080"
}
