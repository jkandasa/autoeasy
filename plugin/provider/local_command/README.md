## Local command plugin
You can run commands on the localhost with this plugin

### plugin configuration
```yaml
provider:
  local_command:
    plugin: local_command # name of the plugin
    config: # configurations for this plugin
      timeout: 2m  # default timeout
      error: # if error occurs what should be the action
        record: true  # record the error
        dir: ./logs/local_command # directory location to save the error details
        filename: errors.txt # filename to save the error details
```

### provider template configuration
#### command example
executes the commands on the localhost
```yaml
command: sleep  # command to be executed
args:   # arguments to the above command
  - "10"
timeout: 11s    # terminates the tasks, if hits the timeout duration
output: # logs location
  dir: ./logs/sleep_cmd
  filename: sleep.txt
  append: false # would you like to append data with existing logs (if any)
```
#### script example
executes the script on the local host
```yaml
script: |   # multiline script sample
  echo 'hello world'
  echo 'script example'
timeout: 30s
output:
  dir: './logs/script'
  filename: url.txt
  append: false
```
