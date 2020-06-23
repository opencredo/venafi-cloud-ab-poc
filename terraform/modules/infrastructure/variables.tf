variable "network_name" {
    type = string
    description = "(optional) Name of network to create"
    default = "ocvab-network"
}

variable "project_id" {
    type = string
    description = "(required) GCP Project ID"
}

variable "region" {
    type = string
    description = "(optional) Region network deployed in"
    default = "europe-west2"
}