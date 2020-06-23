resource "random_pet" "name_randomiser" {}

locals {
  network_name = "${var.network_name}-${random_pet.name_randomiser.id}"
}

module "network" {
  source  = "terraform-google-modules/network/google"
  version = "2.3.0"

  network_name = local.network_name
  project_id = var.project_id

  subnets = [
        {
            subnet_name           = "${var.network_name}-sn-01"
            subnet_ip             = "10.10.10.0/24"
            subnet_region         = var.region
        }
  ]
}
