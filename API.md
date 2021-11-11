## Deployment Configurations

Deployment configurations define the overall configuration
of a golden deployment.

```
PUT /deployment_config/{config ID}
GET /deployment_configs
GET /deployment_config/{config ID}
DELETE /deployment_config/{config ID}
```

## Workloads

Workloads define the indexing and search workloads to be
exercised against a golden deployment as part of a test
scenario.

```
PUT /workload/{workload ID}
GET /workloads
GET /workload/{workload ID}
GET /workload/{workload ID}/payload
DELETE /workload/{workload ID}
```

## Test Scenarios

Test scenarios define the deployment configuration to
setup (potentially parameterized), the workload to
execute, and the validations to be performed.

```
POST /scenarios
GET /scenarios
GET /scenario/{scenario ID}
DELETE /scenario # Doesn't actually delete, just stops the test
```

