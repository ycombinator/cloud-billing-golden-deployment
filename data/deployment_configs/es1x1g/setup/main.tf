terraform {
  required_providers {
    ec = {
      source  = "elastic/ec"
    }
  }
}

variable "stack_version" {
  type    = string
  default = "latest"
}

variable "region" {
  type    = string
  default = "gcp-us-west1"
}

variable "deployment_template_id" {
  type    = string
  default = "gcp-io-optimized-v2"
}

data "ec_stack" "info" {
  version_regex = var.stack_version
  region        = var.region
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "golden_es1x1g" {
  # Optional name.
  name = "golden-es1x1g"

  # Mandatory fields
  region                 = var.region
  version                = data.ec_stack.info.version
  deployment_template_id = var.deployment_template_id

  elasticsearch {}
}
