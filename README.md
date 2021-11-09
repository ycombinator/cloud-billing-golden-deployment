# Elastic Cloud Billing: Golden Deployment Service

This service allows running pre-defined test scenarios for validating that the Elastic Cloud Billing metering and billing calculations are working as expected.

## Structure

The `scenarios` folder contains the various test scenarios, each of varying complexity with regards to the deployment configuration, exercise steps, or other factors. Each scenario is defined in it's own sub-folder with a short and mildly descriptive identifier for that scenario.

Conceptually, each scenario folder describes:
- the setup for the deployment to be used in the test scenario
- steps to exercise the scenario
- expected values for all metering and billing calculations

### Scenario lifecycle

Each scenario is governed by the following lifecycle stages:

* `setup`: This stage sets up the Elastic Cloud deployment, any Elastic Stack resources (e.g. index templates, ILM policies, etc.), and any initial data.

* `exercise`: This stage exercises the deployment. [TODO] Can this be executed multiple times?

* `validate`: This stage performs metering and billing calculations on the deployment and compares their results against expected values. [TODO] Tolerances? Pro-ration?

* `teardown`: This stage tears down everything created during the `setup` lifecycle stage.

## Usage

1. Set Elastic Cloud API Key in environment.
   ```
   export EC_API_KEY=<your Elastic Cloud API Key>
   ```

2. Set the desired scenario in the environment.
   ```
   export EC_BILLING_GDS_SCENARIO=<scenario>
   ```

3. Setup the scenario. This step should be idempotent.
   ```
   make setup
   ```

4. Exercise the deployment.
   ```
   make exercise
   ```

5. Validate the results.
   ```
   make validate
   ```

6. Teardown the scenario.
   ```
   make teardown
   ```
