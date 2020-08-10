#!/usr/bin/env bash

# prepare stackdriver logging integration
# - creates secret which holds ServiceAccount with LogWritter permission (for Stackdriver access)
# - creates installation override for fluent-fit
function prepare_stackdriver_logging() {
    local filepath="${1}"
    if [[ -z "${filepath}" ]]; then
        echo "Filepath for installation override not given"
        return 1
    fi

    # create secret
    echo "creating stackdriver secret"
    kubectl create namespace kyma-system || true # in case the namespace already exists
    kubectl apply -f - << EOF
apiVersion: v1
kind: Secret
metadata:
  name:  gcp-sa-stackdriver
  namespace: kyma-system
data:
  gcp-sa-stackdriver.json: $(base64 -w0 < "${SA_GARDENER_LOGS}")
EOF

    echo "creating stackdriver installation overrides"
    cat << EOF > "${filepath}"
apiVersion: v1
kind: ConfigMap
metadata:
  name: logging-overrides-stackdriver
  namespace: kyma-installer
  labels:
    installer: overrides
    component: logging
    kyma-project.io/installation: ""
data:
  fluent-bit.extraVolumes: |
    - name: gcp-sa-stackdriver
      secret:
        defaultMode: 420
        secretName: gcp-sa-stackdriver

  fluent-bit.extraVolumeMounts: |
    - mountPath: /etc/gcp-sa-stackdriver/
      name: gcp-sa-stackdriver

  fluent-bit.conf.extra: |
    [Output]
        # see stackdriver documentation: https://docs.fluentbit.io/manual/pipeline/outputs/stackdriver
        Name stackdriver
        Match kube.*
        resource global
        google_service_credentials /etc/gcp-sa-stackdriver/gcp-sa-stackdriver.json
EOF
    return 0
}
