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
resource "ec_deployment" "golden_latest" {
  # Optional name.
  name = "golden-latest"

  # Mandatory fields
  region                 = "gcp-us-west1"
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-hot-warm"

  elasticsearch {
    topology {
      id = "hot_content"
    }

    topology {
      id = "master"
    }

    topology {
      id   = "ml"
      size = "1.0g"
    }

    topology {
      id = "warm"
    }
  }

  kibana {}
}
