# autoeasy
autoeasy is inspried by Ansible.
You can automate manual steps into automated way with YAML files.
completely written on GoLang. Just download the binary to your platform and start using it.

## supported plugins
* OpenShift
* Local Command
* Jenkins

## overview
### plugin.yaml file
this file holds all the configuration for your plugins.
you can have multiple entries for the same plugin. Example I want to access different jenkins server

#### file format
```yaml
provider:   # indicates provider plugins configuration
  jenkins:  # you can give any name, you have to use this name in the template file
    plugin: jenkins # name of the plugin
    config: # configuration details for the plugin
      server_url: http://localhost:8080
      username: admin
      password: admin
      insecure: false
      timeout: 1h
  jenkins_eng:
    plugin: jenkins
    config:
      ...
  openshift:
    plugin: openshift
    config:
      ...
```
