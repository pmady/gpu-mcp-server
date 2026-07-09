output "cluster_name" {
  description = "AKS cluster name."
  value       = azurerm_kubernetes_cluster.this.name
}

output "location" {
  description = "Azure region the cluster runs in."
  value       = var.location
}

output "resource_group_name" {
  description = "Resource group holding the cluster."
  value       = azurerm_resource_group.this.name
}

output "cluster_endpoint" {
  description = "AKS Kubernetes API server FQDN."
  value       = azurerm_kubernetes_cluster.this.fqdn
}

output "configure_kubectl" {
  description = "Command to write a kubeconfig entry for the new cluster."
  value       = "az aks get-credentials --resource-group ${azurerm_resource_group.this.name} --name ${azurerm_kubernetes_cluster.this.name} --overwrite-existing"
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
