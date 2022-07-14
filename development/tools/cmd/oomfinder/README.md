# oomfinder

## Overview

oomfinder is a small tool designed to run in a Pod on each k8s worker node as a privileged container. It will check if Docker or Containerd is used and attach to its socket to listen for oom events. If an oom event occurs, oomfinder will print a message to `os stdout` with the following details:

* Time when the oom event occurred
* Namespace where it happened
* Pod name which had this event
* Container name which had this event
* Image used for the impacted container


This is an example log message:
>OOM event received time: 18 May 21 14:19 +0000 , namespace: kyma-system , pod: ory-mechanism-migration-5r9fh ,container: ory-mechanism-migration , image: sha256:b12613fec0c853a73bf27df1bcf051bb6f91e0c1960f0a60ad973f10cc7bdc1c
