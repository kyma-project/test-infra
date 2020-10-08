# Image-guard
This directory contains source code for the Image-guard admission controller.

After installing the chart Image-guard performs the following tasks:
* Monitor and log used docker images in the pods
* Check and allow usage only for whitelisted registries

## Requirements
To deploy the application you need the following:
* helm 3
* K8s cluster version at least 1.16

## Installation
To install image-guard use the following command using your k8s cluster context:
```shell script
$ helm install image-guard ./image-guard
```

## Flags
```
--host -h [host]                The hostname for the service
--port -p [port]                The port to listen on (HTTPS).
--http-only                     Only listen on unencrypted HTTP (e.g. for proxied environments)
--key-path [path-to-file]       The path to the unencrypted TLS key
--cert-path [path-to-file]      The path to the PEM-encoded TLS certificate
--allowed-registry -r [string]  Name of allowed registry
```
