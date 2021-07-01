#!/usr/bin/env bash

# This script is designed to provision a new vm and start kyma with cli. It takes the following optional positional parameters:
# custom VM image --image flag: Use this flag to specify the custom image for provisioning vms. If no flag is provided, the latest custom image is used.
# Kyma version to install --kyma-version flag: Use this flag to indicate which Kyma version the CLI should install (default: same as the CLI)

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
KYMA_PROJECT_DIR=${KYMA_PROJECT_DIR:-"/home/prow/go/src/github.com/kyma-project"}

# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPT_DIR}/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${SCRIPT_DIR}/lib/utils.sh"

cleanup() {
    ARG=$?
    log::info "Removing instance cli-integration-test-${RANDOM_ID}"
    date
    gcloud compute instances delete --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" || true ### Workaround: not failing the job regardless of the vm deletion result
    exit $ARG
}

function testCustomImage() {
    CUSTOM_IMAGE="$1"
    IMAGE_EXISTS=$(gcloud compute images list --filter "name:${CUSTOM_IMAGE}" | tail -n +2 | awk '{print $1}')
    if [[ -z "$IMAGE_EXISTS" ]]; then
        log::error "${CUSTOM_IMAGE} is invalid, it is not available in GCP images list, the script will terminate ..." && exit 1
    fi
}

# Run a command remote on GCP host.
# Arguments:
# required:
# c - command to execute
# z - GCP zone, defaults to $ZONE
# h - GCP host, defaults to $HOST
# optional:
# a - assert
# j - jq filter
function assertRemoteCommand() {

    local OPTIND
    local cmd
    local assert
    local jqFilter
    local zone
    local host

    while getopts ":c:a:j:z:h:" opt; do
        case $opt in
            c)
                cmd="$OPTARG" ;;
            a)
                assert="$OPTARG" ;;
            j)
                jqFilter="$OPTARG" ;;
            z)
                zone="$OPTARG" ;;
            h)
                host="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$cmd" "Command was not provided. Exiting..."
    utils::check_empty_arg "$zone" "GCP zone was not provided. Exiting..."
    utils::check_empty_arg "$host" "GCP host was not provided. Exiting..."

    # config values
    local interval=15
    local retries=10

    local output
    for loopCount in $(seq 1 $retries); do
        output=$(gcloud compute ssh --quiet --zone="${zone}" "${host}" -- "$cmd")
        cmdExitCode=$?

        # check return code and apply assertion (if defined)
        if [ $cmdExitCode -eq 0 ]; then
            if [ -n "$assert" ]; then
                # apply JQ filter (if defined)
                if [ -n "$jqFilter" ]; then
                    local jqOutput
                    jqOutput=$(echo "$output" | jq -r "$jqFilter")
                    jqExitCode=$?

                    # show JQ failures
                    if [ $jqExitCode -eq 0 ]; then
                        output="$jqOutput"
                    else
                        echo "JQFilter '${jqFilter}' failed with exit code '${jqExitCode}' (JSON input: ${output})"
                    fi
                fi
                # assert output
                if [ "$output" = "$assert" ]; then
                    break
                else
                    echo "Assertion failed: expected '$assert' but got '$output'"
                fi
            else
                # no assertion required
                break
            fi
        fi

        # abort if max-retires are reached
        if [ "$loopCount" -ge $retries ]; then
            echo "Failed after $loopCount attempts."
            exit 1
        fi

        echo "Retrying in $interval seconds.."
        sleep $interval
    done;
}

function testVersion {
    log::info "Checking the versions"
assertRemoteCommand \
    -c "sudo kyma version" \
    -z "${ZONE}" \
    -h "cli-integration-test-${RANDOM_ID}" \
    -c "$SOURCE"
}

function testFunction {
    log::info "Create local resources for a sample Function"
    assertRemoteCommand \
        -c "sudo kyma init function --name first-function --runtime nodejs12" \
        -z "${ZONE}" \
        -h "cli-integration-test-${RANDOM_ID}" \
        -c "$SOURCE"

    log::info "Apply local resources for the Function to the Kyma cluster"
    assertRemoteCommand \
        -c "sudo kyma apply function" \
        -z "${ZONE}" \
        -h "cli-integration-test-${RANDOM_ID}" \
        -c "$SOURCE"

    sleep 30

    log::info "Check if the Function is running"
    assertRemoteCommand \
        -c "sudo kubectl get pods -lserverless.kyma-project.io/function-name=first-function,serverless.kyma-project.io/resource=deployment -o jsonpath='{.items[0].status.phase}'" \
        -a 'Running' \
        -z "${ZONE}" \
        -h "cli-integration-test-${RANDOM_ID}" \
        -c "$SOURCE"
}

function testRuntest {
    log::info "Running a simple test on Kyma"
    assertRemoteCommand \
        -c "sudo kyma test run dex-connection" \
        -z "${ZONE}" \
        -h "cli-integration-test-${RANDOM_ID}" \
        -c "$SOURCE"

    echo "Check if the test succeeds"
    assertRemoteCommand \
        -c "sudo kyma test status -o json" \
        -a 'Succeeded' \
        -j '.status.results[0].status' \
        -z "${ZONE}" \
        -h "cli-integration-test-${RANDOM_ID}" \
        -c "$SOURCE"
}

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

RANDOM_ID=$(openssl rand -hex 4)

LABELS=""
if [[ -z "${PULL_NUMBER}" ]]; then
    LABELS=(--labels "branch=$PULL_BASE_REF,job-name=cli-integration")
else
    LABELS=(--labels "pull-number=$PULL_NUMBER,job-name=cli-integration")
fi

# Support configuration via ENV vars (can be be overwritten by CLI args)
KUBERNETES_RUNTIME="${KUBERNETES_RUNTIME:=minikube}"
# Either use the default Kyma install command or the new alpha command.
INSTALLATION="${INSTALLATION:=default}"

POSITIONAL=()
while [[ $# -gt 0 ]]
do

    key="$1"

    case ${key} in
        --image|-i)
            IMAGE="$2"
            testCustomImage "${IMAGE}"
            shift 2
            ;;
        --kyma-version|-kv)
            SOURCE="--source $2"
            shift 2
            ;;
        --kubernetes-runtime|-kr)
            KUBERNETES_RUNTIME="$2"
            shift 2
            ;;
        --installation)
            INSTALLATION="$2"
            shift 2
            ;;
        --*)
            echo "Unknown flag ${1}"
            exit 1
            ;;
        *)    # unknown option
            POSITIONAL+=("$1") # save it in an array for later
            shift # past argument
            ;;
    esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters


if [[ -z "$IMAGE" ]]; then
    log::info "Provisioning vm using the latest default custom image ..."
    date
    IMAGE=$(gcloud compute images list --sort-by "~creationTimestamp" \
         --filter "family:custom images AND labels.default:yes" --limit=1 | tail -n +2 | awk '{print $1}')

    if [[ -z "$IMAGE" ]]; then
       log::error "There are no default custom images, the script will exit ..." && exit 1
    fi
 fi

ZONE_LIMIT=${ZONE_LIMIT:-5}
EU_ZONES=$(gcloud compute zones list --filter="name~europe" --limit="${ZONE_LIMIT}" | tail -n +2 | awk '{print $1}')

for ZONE in ${EU_ZONES}; do
    log::info "Attempting to create a new instance named cli-integration-test-${RANDOM_ID} in zone ${ZONE} using image ${IMAGE}"
    date
    gcloud compute instances create "cli-integration-test-${RANDOM_ID}" \
        --metadata enable-oslogin=TRUE \
        --image "${IMAGE}" \
        --machine-type n1-standard-4 \
        --zone "${ZONE}" \
        --boot-disk-size 200 "${LABELS[@]}" &&\
    log::info "Created cli-integration-test-${RANDOM_ID} in zone ${ZONE}" && break
    log::error "Could not create machine in zone ${ZONE}"
done || exit 1

trap cleanup exit INT

log::info "Building Kyma CLI"
date
cd "${KYMA_PROJECT_DIR}/cli"
make build-linux
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "mkdir \$HOME/bin"

log::info "Copying Kyma CLI to the instance"
#shellcheck disable=SC2088
utils::send_to_vm "${ZONE}" "cli-integration-test-${RANDOM_ID}" "${KYMA_PROJECT_DIR}/cli/bin/kyma-linux" "~/bin/kyma"
gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "sudo cp \$HOME/bin/kyma /usr/local/bin/kyma"

# Provision Kubernetes runtime
log::info "Provisioning Kubernetes runtime '$KUBERNETES_RUNTIME'"
date
if [ "$KUBERNETES_RUNTIME" = 'minikube' ]; then
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma provision minikube --non-interactive"
else
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "curl -s -o install-k3d.sh https://raw.githubusercontent.com/rancher/k3d/main/install.sh && chmod +x ./install-k3d.sh && ./install-k3d.sh"
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma alpha provision k3s --ci"
fi

# Install kyma
log::info "Installing Kyma"
date
if [ "$INSTALLATION" = 'alpha' ]; then
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma alpha deploy --ci ${SOURCE}"
else
    gcloud compute ssh --quiet --zone="${ZONE}" "cli-integration-test-${RANDOM_ID}" -- "yes | sudo kyma install --non-interactive ${SOURCE}"
fi

# Run test suite
testVersion

testFunction


# ON alpha installation there is no dex, therefore skipping the test
if [ "$INSTALLATION" != 'alpha' ]; then
    testRuntest
fi
