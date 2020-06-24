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

variable "network_subnet_cidr" {
    type = string
    description = "(required) A CIDR range to use for the main subnet"
}

variable "network_master_subnet_cidr" {
    type = string
    description = "(required) A CIDR range to use for the k8s master subnet"
}

variable "secondary_pods_cidr" {
    type = string
    description = "(required) A secondary CIDR range to use for the k8s pods"
}

variable "secondary_svcs_cidr" {
    type = string
    description = "(required) A secondary CIDR range to use for the k8s services"
}

variable "cluster_suffix" {
    type = string
    description = "(reqiuired) An identifying suffix for the cluster"
}