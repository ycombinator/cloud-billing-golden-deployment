terraform {
#   required_version = ">= 0.12.29"

  required_providers {
    ec = {
      source  = "elastic/ec"
#       version = "0.4.0"
    }
  }
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "gcp-us-west1"
}

# Create an Elastic Cloud deployment
resource "ec_deployment" "example_minimal" {
  # Optional name.
  name = "golden-latest"

  # Mandatory fields
  region                 = "gcp-us-weat1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "aws-io-optimized-v2"

  # Use the deployment template defaults
  elasticsearch {}

  kibana {}
}
