locals {
  cluster_name = "${var.cluster_name}-${var.cluster_suffix}"
}

module "kubernetes-engine" {
  source  = "terraform-google-modules/kubernetes-engine/google"
  version = "9.4.0"

  ip_range_pods     = local.pods_range_name
  ip_range_services = local.svc_range_name

  name = local.cluster_name

  network = local.network_name

  project_id = var.project_id

  region = var.region

  subnetwork = local.subnet_name
}
