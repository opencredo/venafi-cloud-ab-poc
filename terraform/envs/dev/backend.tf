terraform {
  backend "gcs" {
    bucket = "ocvab-tf-state"
    prefix = "terraform/state"
  }
}

