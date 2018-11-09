# Prow Landscape
In the following the landscape of the prow setup is documented. It outlines the different systems involved and the interconnection between each other.

![Landscape overview](assets/landscape.svg)

The "Prow cluster" is a kubernetes instance managed by the google kubernetes engine as part of the project "kyma-project". In that cluster, the prow components are installed.

# Provisioning and Secrets
The cluster gets provisioned by a script executed manually. I will read the full configuration from the test-infra repository and can be re-created at any time, all configuratin changes should be applied to the repository instead of to the runtime. With that only the operations team needs access to the runtime, all others should manipulate the repository and cause a re-provisioning.
Secrets are stored on a dedicated cloud storage leveraging the key management service.
At provisioning all secrets are read from the cloud storage and installed as kubernetes secrets to the cluster. For the access to the cloud storage a dedicated service account gets used, not present as secret at runtime.

## Internal communication
The prow components are accessing the RBAC protected API server via dedicated service accounts and are communicating without having TLS enabled.

## Anonymous Access
The _hook_ and _deck_ component are exposed via one ingress definition having TLS enabled using a certificate issued for _prow.kyma-project.io_. _deck_ is serving a UI accessible by anonymous to view the build status. 

## Github integration
The _hook_ component is listening for Github events; the URL to _hook_ gets configured in Github as webhook. _hook_ gets a hmac token configured which gets generated as part of the provisioning. That token will be used as secret for the configured webhook on Github side.
Plugins will access Github to modify labels of Pull requests using the Github API with a dedicated API token stored as secret.