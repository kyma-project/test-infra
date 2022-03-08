#!/usr/bin/env bash

set -o errexit
set -o pipefail  # Fail a pipe if any sub-command fails.

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"
export KYMA_SOURCES_DIR="${KYMA_PROJECT_DIR}/kyma"
export COMPASS_SOURCES_DIR="/home/prow/go/src/github.com/kyma-incubator/compass"
export TEST_INFRA_CLUSTER_INTEGRATION_SCRIPTS="${TEST_INFRA_SOURCES_DIR}/prow/scripts/cluster-integration/helpers"

# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/docker.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"

requiredVars=(
    REPO_OWNER
    REPO_NAME
    DOCKER_PUSH_REPOSITORY
    KYMA_PROJECT_DIR
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
    CLOUDSDK_COMPUTE_ZONE
    CLOUDSDK_DNS_ZONE_NAME
    GOOGLE_APPLICATION_CREDENTIALS
)

utils::check_required_vars "${requiredVars[@]}"

export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"

# Enforce lowercase
readonly REPO_OWNER=${REPO_OWNER,,}
export REPO_OWNER
readonly REPO_NAME=${REPO_NAME,,}
export REPO_NAME

readonly RANDOM_NAME_SUFFIX=$(LC_ALL=C tr -dc 'a-z0-9' < /dev/urandom | head -c10)

if [[ "$BUILD_TYPE" == "pr" ]]; then
  # In case of PR, operate on PR number
  readonly COMMON_NAME_PREFIX="gkecompint-pr"
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${PULL_NUMBER}-${RANDOM_NAME_SUFFIX}")
  export JOBGUARD_TIMEOUT="30m"
else
  # Otherwise (main), operate on triggering commit id
  readonly COMMON_NAME_PREFIX="gkecompint-commit"
  readonly COMMIT_ID=$(cd "$COMPASS_SOURCES_DIR" && git rev-parse --short HEAD)
  COMMON_NAME=$(echo "${COMMON_NAME_PREFIX}-${COMMIT_ID}-${RANDOM_NAME_SUFFIX}")
fi

COMMON_NAME=$(echo "${COMMON_NAME}" | tr "[:upper:]" "[:lower:]")

gcp::set_vars_for_network \
  -n "$JOB_NAME"
export GCLOUD_NETWORK_NAME="${gcp_set_vars_for_network_return_net_name:?}"
export GCLOUD_SUBNET_NAME="${gcp_set_vars_for_network_return_subnet_name:?}"

#Local variables
DNS_SUBDOMAIN="${COMMON_NAME}"
COMPASS_SCRIPTS_DIR="${COMPASS_SOURCES_DIR}/installation/scripts"

function createCluster() {
  #Used to detect errors for logging purposes
  ERROR_LOGGING_GUARD="true"

  log::info "Authenticate"
  gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"
  docker::start
  DNS_DOMAIN="$(gcloud dns managed-zones describe "${CLOUDSDK_DNS_ZONE_NAME}" --format="value(dnsName)")"
  export DNS_DOMAIN
  DOMAIN="${DNS_SUBDOMAIN}.${DNS_DOMAIN%?}"
  export DOMAIN

  log::info "Reserve IP Address for Ingressgateway"
  GATEWAY_IP_ADDRESS_NAME="${COMMON_NAME}"
  gcp::reserve_ip_address \
    -n "${GATEWAY_IP_ADDRESS_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -r "$CLOUDSDK_COMPUTE_REGION"
  GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"
  CLEANUP_GATEWAY_IP_ADDRESS="true"
  echo "Created IP Address for Ingressgateway: ${GATEWAY_IP_ADDRESS}"

  log::info "Create DNS Record for Ingressgateway IP"
  gcp::create_dns_record \
    -a "$GATEWAY_IP_ADDRESS" \
    -h "*" \
    -s "$COMMON_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -z "$CLOUDSDK_DNS_ZONE_NAME"
  CLEANUP_GATEWAY_DNS_RECORD="true"

  log::info "Create ${GCLOUD_NETWORK_NAME} network with ${GCLOUD_SUBNET_NAME} subnet"
  gcp::create_network \
    -n "${GCLOUD_NETWORK_NAME}" \
    -s "${GCLOUD_SUBNET_NAME}" \
    -p "$CLOUDSDK_CORE_PROJECT"

  log::info "Provision cluster: \"${COMMON_NAME}\""
  export GCLOUD_SERVICE_KEY_PATH="${GOOGLE_APPLICATION_CREDENTIALS}"
  gcp::provision_k8s_cluster \
        -c "$COMMON_NAME" \
        -m "$MACHINE_TYPE" \
        -n "$NODES_PER_ZONE" \
        -p "$CLOUDSDK_CORE_PROJECT" \
        -v "$GKE_CLUSTER_VERSION" \
        -j "$JOB_NAME" \
        -J "$PROW_JOB_ID" \
        -z "$CLOUDSDK_COMPUTE_ZONE" \
        -R "$CLOUDSDK_COMPUTE_REGION" \
        -N "$GCLOUD_NETWORK_NAME" \
        -S "$GCLOUD_SUBNET_NAME" \
        -P "$TEST_INFRA_SOURCES_DIR"
  CLEANUP_CLUSTER="true"

  utils::generate_self_signed_cert \
      -d "$DNS_DOMAIN" \
      -s "$COMMON_NAME" \
      -v "$SELF_SIGN_CERT_VALID_DAYS"
  export TLS_CERT="${utils_generate_self_signed_cert_return_tls_cert:?}"
  export TLS_KEY="${utils_generate_self_signed_cert_return_tls_key:?}"

  # TODO
  # Prepare Docker external registry overrides
  export DOCKER_PASSWORD=""
  DOCKER_PASSWORD=$(tr -d '\n' < "${GOOGLE_APPLICATION_CREDENTIALS}")

  export DOCKER_REPOSITORY_ADDRESS=""
  DOCKER_REPOSITORY_ADDRESS=$(echo "$DOCKER_PUSH_REPOSITORY" | cut -d'/' -f1)

  export DNS_DOMAIN_TRAILING=${DNS_DOMAIN%.}
  envsubst < "${TEST_INFRA_SOURCES_DIR}/prow/scripts/resources/compass-gke-kyma-overrides.tpl.yaml" > "$PWD/kyma_overrides.yaml"

}

function prometheusMTLSPatch() {
  patchPrometheusForMTLS
  patchAlertManagerForMTLS
  enableNodeExporterMTLS
  patchDeploymentsToInjectSidecar
  patchKymaServiceMonitorsForMTLS
  removeKymaPeerAuthsForPrometheus
  #patchDexPeerAuthForPrometheus
}

function patchPrometheusForMTLS() {
  patch=$(cat <<"EOF"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: monitoring-prometheus
  namespace: kyma-system
spec:
  alerting:
    alertmanagers:
      - apiVersion: v2
        name: monitoring-alertmanager
        namespace: kyma-system
        pathPrefix: /
        port: web
        scheme: https
        tlsConfig:
          caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
          certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
          keyFile: /etc/prometheus/secrets/istio.default/key.pem
          insecureSkipVerify: true
  podMetadata:
    annotations:
      sidecar.istio.io/inject: "true"
      traffic.sidecar.istio.io/includeInboundPorts: ""   # do not intercept any inbound ports
      traffic.sidecar.istio.io/includeOutboundIPRanges: ""  # do not intercept any outbound traffic
      proxy.istio.io/config: |
        # configure an env variable OUTPUT_CERTS to write certificates to the given folder
        proxyMetadata:
          OUTPUT_CERTS: /etc/istio-output-certs
      sidecar.istio.io/userVolumeMount: '[{"name": "istio-certs", "mountPath": "/etc/istio-output-certs"}]' # mount the shared volume at sidecar proxy
  volumes:
    - emptyDir:
        medium: Memory
      name: istio-certs
  volumeMounts:
    - mountPath: /etc/prometheus/secrets/istio.default/
      name: istio-certs
EOF
  )

  echo "${patch}" > patch.yaml
  kubectl apply -f patch.yaml
  rm patch.yaml
}

function patchAlertManagerForMTLS() {
  patch=$(cat <<"EOF"
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: monitoring-alertmanager
  namespace: kyma-system
spec:
  podMetadata:
    annotations:
      sidecar.istio.io/inject: "true"
EOF
  )

  echo "${patch}" > patch.yaml
  kubectl apply -f patch.yaml
  rm patch.yaml
}

function patchDeploymentsToInjectSidecar() {
  allDeploy=(
    kiali-server
    monitoring-kube-state-metrics
    monitoring-operator
    api-gateway
  )

  resource="deployment"
  namespace="kyma-system"

  for depl in "${allDeploy[@]}"; do
    if kubectl get ${resource} -n ${namespace} "${depl}" > /dev/null; then
      kubectl get ${resource} -n ${namespace} "${depl}" -o yaml > "${depl}.yaml"

      if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' -e 's/sidecar.istio.io\/inject: "false"/sidecar.istio.io\/inject: "true"/g' "${depl}.yaml"
      else # assume Linux otherwise
        sed -i 's/sidecar.istio.io\/inject: "false"/sidecar.istio.io\/inject: "true"/g' "${depl}.yaml"
      fi

      kubectl apply -f "${depl}.yaml" || true

      rm "${depl}.yaml"
    fi
  done
}

function enableNodeExporterMTLS() {
  # Note: The two CRDs described in the two variables below are left as they are with all their properties
  # since it's risky to omit some properties due to different strategic merge patch strategies.
  # https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#notes-on-the-strategic-merge-patch

  monitor=$(cat <<"EOF"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  annotations:
    meta.helm.sh/release-name: monitoring
    meta.helm.sh/release-namespace: kyma-system
  labels:
    app: monitoring-node-exporter
    app.kubernetes.io/instance: monitoring
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: monitoring
    chart: monitoring-1.0.0
    helm.sh/chart: monitoring-1.0.0
    release: monitoring
  name: monitoring-node-exporter
  namespace: kyma-system
spec:
  endpoints:
  - metricRelabelings:
    - action: keep
      regex: ^(go_goroutines|go_memstats_alloc_bytes|go_memstats_heap_alloc_bytes|go_memstats_heap_inuse_bytes|go_memstats_heap_sys_bytes|go_memstats_stack_inuse_bytes|node_.*|process_cpu_seconds_total|process_max_fds|process_open_fds|process_resident_memory_bytes|process_start_time_seconds|process_virtual_memory_bytes)$
      sourceLabels:
      - __name__
    port: metrics
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
      certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
      keyFile: /etc/prometheus/secrets/istio.default/key.pem
      insecureSkipVerify: true
  jobLabel: jobLabel
  selector:
    matchLabels:
      app: prometheus-node-exporter
      release: monitoring

EOF
  )
  echo "$monitor" > monitor.yaml

  # The patches around the DaemonSet involve an addition of two init containers that together setup certificates
  # for the node-exporter application to use. There are also two new mounts - a shared directory (node-certs)
  # and the Istio CA secret (istio-certs).
 
  daemonset=$(cat <<"EOF"
apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    deprecated.daemonset.template.generation: "1"
    meta.helm.sh/release-name: monitoring
    meta.helm.sh/release-namespace: kyma-system
  labels:
    app: prometheus-node-exporter
    app.kubernetes.io/instance: monitoring
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: prometheus-node-exporter
    chart: prometheus-node-exporter-1.12.0
    helm.sh/chart: prometheus-node-exporter-1.12.0
    jobLabel: node-exporter
    release: monitoring
  name: monitoring-prometheus-node-exporter
  namespace: kyma-system
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: prometheus-node-exporter
      release: monitoring
  template:
    metadata:
      labels:
        app: prometheus-node-exporter
        app.kubernetes.io/instance: monitoring
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/name: prometheus-node-exporter
        chart: prometheus-node-exporter-1.12.0
        helm.sh/chart: prometheus-node-exporter-1.12.0
        jobLabel: node-exporter
        release: monitoring
    spec:
      initContainers:
      - name: certs-init
        image: emberstack/openssl:alpine-latest
        command: ['sh', '-c', 'openssl req -newkey rsa:2048 -nodes -days 365000 -subj "/CN=$(NODE_NAME)" -keyout /etc/certs/node.key -out /etc/certs/node.csr && openssl x509 -req -days 365000 -set_serial 01 -in /etc/certs/node.csr -out /etc/certs/node.crt -CA /etc/istio/certs/ca-cert.pem -CAkey /etc/istio/certs/ca-key.pem']
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/istio/certs
          readOnly: true
        - name: node-certs
          mountPath: /etc/certs
      - name: web-config-init
        image: busybox:1.34.1
        command: ['sh', '-c', 'printf "tls_server_config:\\n  cert_file: /etc/certs/node.crt\\n  key_file: /etc/certs/node.key\\n  client_auth_type: \"RequireAndVerifyClientCert\"\\n  client_ca_file: /etc/istio/certs/ca-cert.pem" > /etc/certs/web.yaml']
        volumeMounts:
        - name: node-certs
          mountPath: /etc/certs
      containers:
      - args:
        - --path.procfs=/host/proc
        - --path.sysfs=/host/sys
        - --path.rootfs=/host/root
        - --web.listen-address=$(HOST_IP):9100
        - --web.config=/etc/certs/web.yaml
        - --collector.filesystem.ignored-mount-points=^/(dev|proc|sys|var/lib/docker/.+)($|/)
        - --collector.filesystem.ignored-fs-types=^(autofs|binfmt_misc|cgroup|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|sysfs|tracefs)$
        env:
        - name: HOST_IP
          value: 0.0.0.0
        image: eu.gcr.io/kyma-project/tpi/node-exporter:1.0.1-1de56388
        imagePullPolicy: IfNotPresent
        name: node-exporter
        livenessProbe: null
        readinessProbe: null
        ports:
        - containerPort: 9100
          hostPort: 9100
          name: metrics
          protocol: TCP
        resources: {}
        securityContext:
          allowPrivilegeEscalation: false
          privileged: false
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/certs
          name: node-certs
        - name: istio-certs
          mountPath: /etc/istio/certs
        - mountPath: /host/proc
          name: proc
          readOnly: true
        - mountPath: /host/sys
          name: sys
          readOnly: true
        - mountPath: /host/root
          mountPropagation: HostToContainer
          name: root
          readOnly: true
      dnsPolicy: ClusterFirst
      hostNetwork: true
      hostPID: true
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext:
        fsGroup: 65534
        runAsGroup: 65534
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccount: monitoring-prometheus-node-exporter
      serviceAccountName: monitoring-prometheus-node-exporter
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        operator: Exists
      volumes:
      - name: istio-certs
        secret:
          secretName: istio-ca-secret
      - name: node-certs
        emptyDir: {}
      - hostPath:
          path: /proc
          type: ""
        name: proc
      - hostPath:
          path: /sys
          type: ""
        name: sys
      - hostPath:
          path: /
          type: ""
        name: root
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate

EOF
  )
  echo "$daemonset" > daemonset.yaml

  kubectl get secret istio-ca-secret --namespace=istio-system -o yaml | grep -v '^\s*namespace:\s' | kubectl replace --force --namespace=kyma-system -f -

  kubectl apply -f monitor.yaml
  kubectl apply -f daemonset.yaml

  rm monitor.yaml
  rm daemonset.yaml
} 

function patchKymaServiceMonitorsForMTLS() {
  kymaSvcMonitors=(
    kiali
    logging-fluent-bit
    logging-loki
    ory-oathkeeper-maester
    ory-hydra-maester
    tracing-jaeger-operator
    tracing-jaeger
    monitoring-grafana
    monitoring-alertmanager
    monitoring-operator 
    monitoring-kube-state-metrics 
    dex
    api-gateway
    monitoring-prometheus-pushgateway
  )

  crd="servicemonitors.monitoring.coreos.com"
  namespace="kyma-system"
  patchContent=$(cat <<"EOF"
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
      certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
      keyFile: /etc/prometheus/secrets/istio.default/key.pem
      insecureSkipVerify: true

EOF
  )

  echo "$patchContent" > tmp_patch_content.yaml

  for sm in "${kymaSvcMonitors[@]}"; do
    if kubectl get ${crd} -n ${namespace} "${sm}" > /dev/null; then
      kubectl get ${crd} -n ${namespace} "${sm}" -o yaml > "${sm}.yaml"

      if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' -e '/ port:/r tmp_patch_content.yaml' "${sm}.yaml"
        sed -i '' -e '/ path:/r tmp_patch_content.yaml' "${sm}.yaml"
      else # assume Linux otherwise
        sed -i '/ port:/r tmp_patch_content.yaml' "${sm}.yaml"
        sed -i '/ path:/r tmp_patch_content.yaml' "${sm}.yaml"
      fi

      kubectl apply -f "${sm}.yaml" || true

      rm "${sm}.yaml"
    fi
  done

  rm tmp_patch_content.yaml
}

function removeKymaPeerAuthsForPrometheus() {
  crd="peerauthentications.security.istio.io"
  namespace="kyma-system"

  allPAs=(
    kiali
    logging-fluent-bit-metrics
    logging-loki
    monitoring-grafana-policy
    ory-oathkeeper-maester-metrics
    ory-hydra-maester-metrics
    tracing-jaeger-operator-metrics
    tracing-jaeger-metrics
    monitoring-prometheus-pushgateway
  )

  for pa in "${allPAs[@]}"; do
    kubectl delete ${crd} -n ${namespace} "${pa}" || true
  done
}

function patchDexPeerAuthForPrometheus() {
  crd="peerauthentication"
  namespace="kyma-system"
  name="dex-service"

  patchDex=$(cat <<"EOF"
  portLevelMtls:
    5558:
      mode: STRICT
EOF
  )

  if kubectl get ns ${namespace} > /dev/null; then
    echo "${patchDex}" > patch-dex.yaml
    kubectl get ${crd} -n ${namespace} ${name} -o yaml > dex-pa.yaml

    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' -e '/^spec:/r patch-dex.yaml' dex-pa.yaml
    else # assume Linux otherwise
      sed -i '/^spec:/r patch-dex.yaml' dex-pa.yaml
    fi

    kubectl apply -f dex-pa.yaml

    rm dex-pa.yaml
    rm patch-dex.yaml
  fi
}

function installKyma() {
  kyma::install_cli

  KYMA_VERSION=$(<"${COMPASS_SOURCES_DIR}/installation/resources/KYMA_VERSION")
  kyma deploy --ci --source="${KYMA_VERSION}" --workspace "$KYMA_SOURCES_DIR" --verbose --values-file "$PWD/kyma_overrides.yaml"

  # TODO this function does not patch dex
  prometheusMTLSPatch
}

function installCompass() {
  compassUnsetVar=false

  # shellcheck disable=SC2043
  for var in GATEWAY_IP_ADDRESS ; do
    if [ -z "${!var}" ] ; then
      echo "ERROR: $var is not set"
      compassUnsetVar=true
    fi
  done
  if [ "${compassUnsetVar}" = true ] ; then
    exit 1
  fi

  COMPASS_OVERRIDES="${COMPASS_SOURCES_DIR}/installation/resources/compass-overrides-gke-integration.yaml"
  bash "${COMPASS_SCRIPTS_DIR}"/install-compass.sh --overrides-file "${COMPASS_OVERRIDES}" --timeout 30m0s
  STATUS=$(helm status compass -n compass-system -o json | jq .info.status)
  echo "Compass installation status ${STATUS}"

  if [ -n "$(kubectl get service -n kyma-system apiserver-proxy-ssl --ignore-not-found)" ]; then
    log::info "Create DNS Record for Apiserver proxy IP"
    APISERVER_IP_ADDRESS=$(kubectl get service -n kyma-system apiserver-proxy-ssl -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    gcp::create_dns_record \
        -a "$APISERVER_IP_ADDRESS" \
        -h "apiserver" \
        -s "$COMMON_NAME" \
        -p "$CLOUDSDK_CORE_PROJECT" \
        -z "$CLOUDSDK_DNS_ZONE_NAME"
    CLEANUP_APISERVER_DNS_RECORD="true"
  fi
}

# Using set -f to prevent path globing in post_hook arguments.
# utils::post_hook call set +f at the beginning.
trap 'EXIT_STATUS=$?; set -f; utils::post_hook -n "$COMMON_NAME" -p "$CLOUDSDK_CORE_PROJECT" -c "$CLEANUP_CLUSTER" -g "$CLEANUP_GATEWAY_DNS_RECORD" -G "$INGRESS_GATEWAY_HOSTNAME" -a "$CLEANUP_APISERVER_DNS_RECORD" -A "$APISERVER_HOSTNAME" -I "$CLEANUP_GATEWAY_IP_ADDRESS" -l "$ERROR_LOGGING_GUARD" -z "$CLOUDSDK_COMPUTE_ZONE" -R "$CLOUDSDK_COMPUTE_REGION" -r "$PROVISION_REGIONAL_CLUSTER" -d "$DISABLE_ASYNC_DEPROVISION" -s "$COMMON_NAME" -e "$GATEWAY_IP_ADDRESS" -f "$APISERVER_IP_ADDRESS" -N "$COMMON_NAME" -Z "$CLOUDSDK_DNS_ZONE_NAME" -E "$EXIT_STATUS" -j "$JOB_NAME"; ' EXIT INT


if [[ "${BUILD_TYPE}" == "pr" ]]; then
    log::info "Execute Job Guard"
    export JOB_NAME_PATTERN="(pre-compass-components-.*)|(^pre-compass-tests$)"
    export JOBGUARD_TIMEOUT="30m"
    "${TEST_INFRA_SOURCES_DIR}/development/jobguard/scripts/run.sh"
fi

log::info "Create new cluster"
createCluster

log::info "Install Kyma"
installKyma

log::info "Install Compass"
installCompass

log::info "Test Kyma with Compass"
CONCURRENCY=1 "${TEST_INFRA_SOURCES_DIR}"/prow/scripts/kyma-testing.sh -l "!benchmark"

log::success "Success"

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"
