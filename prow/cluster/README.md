# Cluster

## Overview

This folder contains files related to the configuration of the Prow production cluster that are used during the cluster provisioning.

### Project Structure

<!-- Update the folder structure each time you modify it. -->

The structure of the folder looks as follows:

```
  ├── components            # Definitions of Prow components and cluster configuration.
  ├── resources             # Helm charts used by the Prow cluster.
  └── static-files          # Files that will be uploaded to the nginx web server.
```

###  Adding Static Files
All files added to the `static-files` folder are automatically uploaded by the Prow `config_updater` plugin to the cluster in a Config Map. Uploaded files are mounted by the web server in the web root directory. To route traffic for a specific path to the NGINX web server, in order to serve these files, update the Ingress `tls-ing` configuration in `tls-ing_ingress.yaml`.
