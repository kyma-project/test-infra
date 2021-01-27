#!/usr/bin/env bash


# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"

# prepare stackdriver logging integration
# - creates secret which holds ServiceAccount with LogWritter permission (for Stackdriver access)
# - creates installation override for fluent-fit
function prepare_stackdriver_logging() {
    local filepath="${1}"
    if [[ -z "${filepath}" ]]; then
        log::error "Filepath for installation override not given"
        return 1
    fi
    local kyma_system_namespace="kyma-system"

    # the namespace kyma-system is required in order to create the following secret
    # `kyma install` would also create it, however this is to late in the installation process
    log::info "creating kyma-system namespace"
    if errorMessage=$(kubectl create namespace "${kyma_system_namespace}" 2>&1); then
        log::info "namespace kyma-system created"
    else
        if [[ ${?} == 1 && ${errorMessage} == *"AlreadyExists"* ]]; then
            log::warn "namespace already exists"
        else
            log::error "namespace kyma-system could not be created"
            return 1
        fi
    fi

    # create secret
    log::info "creating stackdriver secret"
    kubectl apply -f - << EOF
apiVersion: v1
kind: Secret
metadata:
  name:  gcp-sa-stackdriver
  namespace: kyma-system
data:
  gcp-sa-stackdriver.json: $(base64 -w0 < "${SA_GARDENER_LOGS}")
EOF

    log::info "creating stackdriver installation overrides"
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

  fluent-bit.config.extra: |
    [Output]
        # see stackdriver documentation: https://docs.fluentbit.io/manual/pipeline/outputs/stackdriver
        Name stackdriver
        Match kube.*
        resource global
        google_service_credentials /etc/gcp-sa-stackdriver/gcp-sa-stackdriver.json
EOF
    return 0
}
