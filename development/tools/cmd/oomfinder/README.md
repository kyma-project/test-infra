# oomfinder

## Overview

oomfinder is a small tool designed to run in a Pod on each k8s worker node as a privileged container. It will check if docker or containerd is used and attach to its socket to listen for oom events. On oom event, oomfinder will print message to os stdout with following details.

* time when oom event occur
* namespace where it happened
* pod name which had this event
* container name which had this event
* image used for impacted container


Example log message
>OOM event received time: 18 May 21 14:19 +0000 , namespace: kyma-system , pod: ory-mechanism-migration-5r9fh ,container: ory-mechanism-migration , image: sha256:b12613fec0c853a73bf27df1bcf051bb6f91e0c1960f0a60ad973f10cc7bdc1c

## Usage
