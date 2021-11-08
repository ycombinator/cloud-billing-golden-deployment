# Usage
1. Set Elastic Cloud API Key in environment.
   ```
   export EC_API_KEY=<your Elastic Cloud API Key>
   ```

2. Create deployment for desired stack version.
   ```
   cd deployments/<version>   # latest = latest stable release
   terraform apply --auto-approve=true
   ```

# TODO
* Add workloads + runner
* Add more deployments for other stack versions
