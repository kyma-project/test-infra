# Image-guard
This directory contains source code for the image-guard admission controller.

Image-guard performs the following tasks:
* Monitor and log used Docker images in the Pods.
* Optionally, check and allow the usage of the whitelisted registries only.

## Prerequisites
To deploy the application, you need the following:
* Helm 3
* Kubernetes cluster (at least 1.16 version)

## Installation
To install image-guard, use the following command using your Kubernetes cluster context:
```shell script
$ helm install image-guard ./image-guard
```
If you want to enable enforcing image registries, use the `--set enforcedRegistry.enabled=true` Helm flag and define the registries.

## Flags
These are the flags you can use for this application:
```
--host -h [host]                The hostname for the service
--port -p [port]                The HTTPS port to listen on
--http-only                     The flag that enables listening on unencrypted HTTP only (for example: for proxied environments)
--key-path [path-to-file]       The path to the unencrypted TLS key
--cert-path [path-to-file]      The path to the PEM-encoded TLS certificate
--allowed-registry -r [string]  The name of the allowed registry
```
