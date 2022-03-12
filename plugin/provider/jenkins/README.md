## Jenkins plugin
You can communicate with jenkins server with this plugin

### plugin config
```yaml
provider:
  jenkins:
    plugin: jenkins # name of the plugin
    config: # configuration
      server_url: http://localhost:8080 # jenkins server url
      username: admin   # jenkins server username
      password: admin   # jenkins server password or token
      insecure: false   # allow insecure login (example: jenkins server with ssl certificate)
      timeout: 1h       # default timeout
```

### provider template configuration
```yaml
function: build # name of the function
config: # configuration details
  wait_for_completion: true # waits till the job completes
  retry_count: 2 # number of retry count, if the job failed
  timeout: 30m # overrides the default timeout
data:   # data used to run the function
  - job_name: abc
    limit: 5
    parameters:
      ABC: XYZ
      ACC: QWE 
```

#### Function - build
`build` function creates a jenkins build on a specified jenkins job with the given parameters.
* if `wait_for_completion` is `false`, returns immediately
* `retry_count`: number of retry, if the job failed
* `timeout`: fail the task immediately, if it reaches the timeout duration
```yaml
function: build
config:
  wait_for_completion: true
  retry_count: 2
  timeout: 30m
data:
  - job_name: abc # job name
    limit: 5  # limit the results, used to filter by queue id
    parameters: # parameters used to build a job
      ABC: XYZ
      ACC: QWE 
```
