terraform {
  required_providers {
    zoom = {
      source  = "folio-sec/zoom"
      version = "~> 0.0.0"
    }
  }
}

provider "zoom" {
  account_id    = var.zoom_account_id
  client_id     = var.zoom_client_id
  client_secret = var.zoom_client_secret
}

variable "zoom_account_id" {}

variable "zoom_client_id" {}

variable "zoom_client_secret" {
  sensitive = true
}
