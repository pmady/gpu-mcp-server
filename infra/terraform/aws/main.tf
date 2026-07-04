data "aws_availability_zones" "available" {
  state = "available"

  # Only zones that don't require explicit opt-in (skips Local/Wavelength zones).
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  # Spread the VPC across up to 3 AZs so the GPU node groups can land wherever
  # the chosen instance types have capacity.
  azs = slice(data.aws_availability_zones.available.names, 0, 3)

  tags = merge(
    {
      Project   = "gpu-mcp-server"
      Component = "gpu-e2e-test"
      ManagedBy = "terraform"
      Stack     = "infra/terraform/aws"
    },
    var.tags,
  )

  # Two fixed-size, ON_DEMAND managed node groups — one per image architecture
  # the gpu-mcp-server DaemonSet ships for — so both amd64 and arm64 images get
  # validated on real GPU hardware in the same cluster. Setting a count to 0
  # cleanly omits that node group (e.g. to test only one arch at a time).
  node_groups = merge(
    var.gpu_amd64_node_count > 0 ? {
      gpu_amd64 = {
        ami_type       = var.gpu_amd64_ami_type
        instance_types = [var.gpu_amd64_instance_type]
        capacity_type  = "ON_DEMAND"

        min_size     = var.gpu_amd64_node_count
        max_size     = var.gpu_amd64_node_count
        desired_size = var.gpu_amd64_node_count

        disk_size = var.gpu_node_disk_size

        labels = {
          "gpu-mcp-server.io/pool" = "gpu-amd64"
        }

        # NOTE: intentionally NOT tainted. This is a single-purpose test
        # cluster, so the GPU operator's controllers and CoreDNS must be able
        # to schedule on the GPU nodes too. The gpu-mcp-server chart tolerates
        # `nvidia.com/gpu` regardless, so adding a taint here later is safe
        # for the DaemonSet but would strand system/add-on pods unless you
        # also add a separate CPU node group.
      }
    } : {},
    var.gpu_arm64_node_count > 0 ? {
      gpu_arm64 = {
        ami_type       = var.gpu_arm64_ami_type
        instance_types = [var.gpu_arm64_instance_type]
        capacity_type  = "ON_DEMAND"

        min_size     = var.gpu_arm64_node_count
        max_size     = var.gpu_arm64_node_count
        desired_size = var.gpu_arm64_node_count

        disk_size = var.gpu_node_disk_size

        labels = {
          "gpu-mcp-server.io/pool" = "gpu-arm64"
        }

        # NOTE: intentionally NOT tainted, same rationale as gpu_amd64 above.
      }
    } : {},
  )
}

###############################################################################
# Networking
###############################################################################

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "6.6.1"

  name = var.cluster_name
  cidr = var.vpc_cidr

  azs             = local.azs
  private_subnets = [for i in range(length(local.azs)) : cidrsubnet(var.vpc_cidr, 4, i)]
  public_subnets  = [for i in range(length(local.azs)) : cidrsubnet(var.vpc_cidr, 8, i + 48)]

  # Nodes live in private subnets and reach the internet (image pulls, the
  # NVIDIA NGC helm repo, etc.) via a single shared NAT gateway. One NAT gateway
  # keeps the test cluster cheap; it is not HA, which is fine for a throwaway.
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }
}

###############################################################################
# EKS control plane + dual-arch GPU node groups
###############################################################################

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "21.23.0"

  name               = var.cluster_name
  kubernetes_version = var.kubernetes_version

  # Public API endpoint so an operator can reach the cluster with a kubeconfig
  # from their laptop. This is a throwaway test cluster; restrict this for
  # anything that outlives a test run.
  endpoint_public_access = true

  # Add the identity running `terraform apply` as a cluster admin via an EKS
  # access entry, so the kubernetes/helm providers can install the add-ons in
  # the same apply.
  enable_cluster_creator_admin_permissions = true

  # Core EKS-managed add-ons. Without these the cluster has NO pod networking,
  # so nodes never reach Ready and the node groups fail with
  # "NodeCreationFailure: Unhealthy nodes". vpc-cni (and the pod identity agent)
  # must be installed BEFORE the node groups join — hence before_compute = true;
  # coredns/kube-proxy can settle once nodes exist.
  addons = {
    vpc-cni                = { before_compute = true }
    eks-pod-identity-agent = { before_compute = true }
    kube-proxy             = {}
    coredns                = {}
  }

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = local.node_groups
}
