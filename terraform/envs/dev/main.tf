resource "random_pet" "cluster_suffix" {}

module "infrastructure" {
  source = "../../modules/infrastructure"

  project_id = "venafi-afterburners"

  network_subnet_cidr        = "10.0.0.0/17"
  network_master_subnet_cidr = "10.60.0.0/17"
  secondary_pods_cidr        = "192.168.0.0/18"
  secondary_svcs_cidr        = "192.168.64.0/18"

  cluster_suffix = random_pet.cluster_suffix.id
}