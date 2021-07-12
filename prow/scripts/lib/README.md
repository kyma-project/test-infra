# Cluster

## Overview

The folder contains helper scripts with commonly used functions.

### Project structure

The folder structure looks as follows:

```
├── azure.sh # This script contains functions that interact with Azure services.
├── clitests.sh # This script contain function for deploying Kyma.
├── docker.sh # This script contains functions that interact with Docker.
├── gardener # This directory contains helper scripts used by Gardener pipeline jobs.
├── gcp.sh # This script contains functions that interact with Google Cloud services.
├── github.sh # This script contains function that configure git.
├── junit.sh # This script contains functions  used for testing with JUnit.
├── kyma.sh # This script contains functions used for installing and interfacing with Kyma.
├── log.sh # This script provides unified logging functions.
├── testing-helpers.sh # This script contains functions adiding Kyma testing.
└── utils.sh # This script contains various functions that couldn't be assigned to any of the other helper scripts.
```

### Example use case
The following example use case reserves an IP address from the Google Compute Engine and creates a new DNS record that uses the IP address.

```bash
#!/usr/bin/env bash

# this script will reserve IP address and create a DNS record with this address

export TEST_INFRA_SOURCES_DIR="${KYMA_PROJECT_DIR}/test-infra"

# shellcheck source=prow/scripts/lib/log.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/utils.sh"
# shellcheck source=prow/scripts/lib/kyma.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/kyma.sh"
# shellcheck source=prow/scripts/lib/gcp.sh
source "${TEST_INFRA_SOURCES_DIR}/prow/scripts/lib/gcp.sh"

# make sure all required variables are set
requiredVars=(
    GATEWAY_IP_ADDRESS_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    GATEWAY_DNS_COMMON_NAME
    CLOUDSDK_CORE_PROJECT
    CLOUDSDK_COMPUTE_REGION
)

utils::check_required_vars "${requiredVars[@]}"

gcp::authenticate \
    -c "${GOOGLE_APPLICATION_CREDENTIALS}"

gcp::reserve_ip_address \
    -n "$GATEWAY_IP_ADDRESS_NAME" \
    -p "$CLOUDSDK_CORE_PROJECT" \
    -r "$CLOUDSDK_COMPUTE_REGION"
export GATEWAY_IP_ADDRESS="${gcp_reserve_ip_address_return_ip_address:?}"

gcp::create_dns_record \
-a "$GATEWAY_IP_ADDRESS" \
-h "*" \
-s "$GATEWAY_DNS_COMMON_NAME" \
-p "$CLOUDSDK_CORE_PROJECT" \
-z "$CLOUDSDK_DNS_ZONE_NAME"
```
