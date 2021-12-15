## Deployment Templates

Deployment templates define the overall template
of a golden deployment. When a [test scenario](#Test_Scenarios) starts
executing, it spins up a new deployment (or ensures that one already exists).
This deployment is templateured according to the deployment template specified
in the test scenario.

### Create or update a deployment template
_Not implemented yet._
```
PUT /deployment_template/{template ID}
{
  "vars": {
    "stack_version": {
      "type": "string",
      "default": "7.15.2"
    },
    "region": {
      "type": "string",
      "default": "gcp-us-central1"
    }
  },
  "template": {
    "resources": {
      "elasticsearch": [
        {
          "region": "{{ vars.region }}",
          "ref_id": "main-elasticsearch",
          "plan": {
            "cluster_topology": [
              {
                "zone_count": 1,
                "elasticsearch": {
                  "node_attributes": {
                    "data": "hot"
                  }
                },
                "instance_configuration_id": "gcp.es.datahot.n2.68x10x45",
                "node_roles": [
                  "master",
                  "ingest",
                  "data_hot",
                  "data_content"
                ],
                "id": "hot_content",
                "size": {
                  "resource": "memory",
                  "value": 1024
                }
              }
            ],
            "elasticsearch": {
              "version": "{{ vars.stack_version }}"
            },
            "deployment_template": {
              "id": "gcp-storage-optimized"
            }
          }
        }
      ]
    }
  }
}
```

### List deployment templates
```
GET /deployment_templates
```

### View a deployment template
```
GET /deployment_template/{template ID}
```

### Show a deployment template's contents
```
GET /deployment_template/{template ID}/payload
```

### Delete a deployment template
```
DELETE /deployment_template/{template ID}
```

## Test Scenarios

Test scenarios define the deployment to spin up (or ensure already exists),
the workload template to execute against the deployment, and the validations to be performed.

### Create a test scenario and start running it
```
POST /scenarios
{
  "deployment_template": {
    "id": "es1x1g",
    "vars": {
      "stack_version": "7.14.0"
    }
  },
  "workload": {
    "start_offset_seconds": 0,
    "min_interval_seconds": 0,
    "max_interval_seconds": 3,
    "max_requests_per_second": 3,
    "index_to_search_ratio": 4
  },
  "validations": {
    "frequency_seconds": 86400,
    "query": {
      "start_timestamp": "now-1d",
      "end_timestamp": "now"
    },
    "expectations": {
      "instance_capacity_gb_hours": { "min": 12345, "max": 23456 },
      "data_out_gb": { "min": 345678, "max": 456789 },
      "data_internode_gb": { "min": 5678, "max": 6789 },
      "snapshot_storage_size_gb": { "min": 678, "max": 789 },
      "snapshot_api_requests_count": { "min": 7890, "max": 8901 }
    } 
  }
}
```

### List test scenarios
```
GET /scenarios
```

### Show a test scenario's configuration
```
GET /scenario/{scenario ID}
```

### Stop running a test scenario
```
DELETE /scenario/{scenario ID} 
```

