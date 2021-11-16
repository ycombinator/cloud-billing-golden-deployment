## Deployment Configurations

Deployment configurations define the overall configuration
of a golden deployment. When a [test scenario](#Test_Scenarios) starts
executing, it spins up a new deployment (or ensures that one already exists).
This deployment is configured according to the deployment configuration specified
in the test scenario.

### Create or update a deployment configuration
_Not implemented yet._
```
PUT /deployment_config/{config ID}
// Terraform deployment configuration contents
```

### List deployment configurations
```
GET /deployment_configs
```

### View a deployment configuration
```
GET /deployment_config/{config ID}
```

### Show a deployment configuration's contents
```
GET /deployment_config/{config ID}/payload
```

### Delete a deployment configuration
```
DELETE /deployment_config/{config ID}
```

## Test Scenarios

Test scenarios define the deployment to spin up (or ensure already exists),
the workload configuration to execute against the deployment, and the validations to be performed.

### Create a test scenario and start running it
```
POST /scenarios
{
  "deployment_configuration": {
    "id": "es1x1g",
    "variables": {
      "stack_version": "latest"
    }
  },
  "workload": {
    "start_offset_seconds": 0,
    "min_interval_seconds": 0,
    "max_interval_seconds": 3,
    "index_to_search_ratio": 4
  },
  "validations": {
    "frequency": "daily",
    "start_timestamp": "now-1d",
    "end_timestamp": "now",
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

