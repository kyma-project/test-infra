# E2E K3d task

This task runs a custom script with automatic provisioning of the K3d cluster with the registry running in Docker-in-Docker.

### Parameters

* **image**: script base image
* **script**: runtime script
* **provisionRegistry**: enables registry provisioning (default: `true`)
* **registryPort**: port used during registry provisioning (default: `5000`)
* **k3dVersion**: defines which k3d version to use (default: `5.5.1`)
* **trace** enables k3d trace output (default: `false`)

### Workspaces

* **source**: a workspace containing the application source

## Using internal registry

To use the internal registry, the PipelineRun spec needs to contain the following `podTemplate`:
```yaml
  podTemplate:
    hostAliases:
      - ip: 127.0.0.1
        hostnames:
          - "k3d-registry.localhost"
```

## Usage

For usage examples, see the files in [samples](./samples).