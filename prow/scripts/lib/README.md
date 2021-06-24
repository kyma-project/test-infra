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
├── gcloud.sh # This script contains functions that interact with Google Cloud services.
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

# make sure all required variables are set
requiredVars=(
    GATEWAY_IP_ADDRESS_NAME
    GOOGLE_APPLICATION_CREDENTIALS
    GATEWAY_DNS_FULL_NAME
)

utils::check_required_vars "${requiredVars[@]}"

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"

log::info "Reserving IP address"
GATEWAY_IP_ADDRESS=$(gcloud::reserve_ip_address "${GATEWAY_IP_ADDRESS_NAME}")

gcloud::create_dns_record "${GATEWAY_IP_ADDRESS}" "${GATEWAY_DNS_FULL_NAME}"

log::success "Created DNS record for ${GATEWAY_IP_ADDRESS} IP address"
```
