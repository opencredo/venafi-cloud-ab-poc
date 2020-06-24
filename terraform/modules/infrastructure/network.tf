locals {
  network_name       = "${var.network_name}-${var.cluster_suffix}"
  subnet_name        = "${local.network_name}-subnet"
  master_auth_subnet = "${local.network_name}-master-subnet"
  pods_range_name    = "ip-range-pods-${var.cluster_suffix}"
  svc_range_name     = "ip-range-svc-${var.cluster_suffix}"
}

module "network" {
  source  = "terraform-google-modules/network/google"
  version = "2.3.0"

  network_name = local.network_name
  project_id   = var.project_id

  subnets = [
    {
      subnet_name   = local.subnet_name
      subnet_ip     = var.network_subnet_cidr
      subnet_region = var.region
    },
    {
      subnet_name   = local.master_auth_subnet
      subnet_ip     = var.network_master_subnet_cidr
      subnet_region = var.region
    },
  ]

  secondary_ranges = {
    "${local.subnet_name}" = [
      {
        range_name    = local.pods_range_name
        ip_cidr_range = var.secondary_pods_cidr
      },
      {
        range_name    = local.svc_range_name
        ip_cidr_range = var.secondary_svcs_cidr
      },
    ]
  }
}
