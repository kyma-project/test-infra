# Cluster

## Overview

The folder contains helper scripts used in pipelines

### Project structure

The structure of the folder looks as follows:

```
├── azure.sh # This script contains functions that interact with Azure services.
├── cli-alpha.sh # This script contains functions used for running test suite.
├── clitests.sh # This script contain function for deploying Kyma.
├── docker.sh # This script contains functions that interact with Docker.
├── gardener # This directory contains helper scripts used by Gardener pipeline jobs.
├── gcloud.sh # This script contains functions that interact with Google Cloud services.
├── github.sh # This script contains function that configure git.
├── junit.sh # This script contains functions  used for testing with JUnit.
├── kyma.sh # This script contains functions used for installing and interfacing with Kyma.
├── log.sh # This script provides unified logging functions.
├── testing-helpers.sh # This script contains functions adiding Kyma testing.
└── utils.sh # This script contains various functions that couldn't be assigned to any of the other helper scripts.
```

### Log library
`log::info`, `log::warn`, and `log::error` functions takes in one argument and prints it together with current time and message severity, for instance `log::info` could print following message:
```
2021/01/19 14:44:06 UTC [INFO] Build kcp-installer with target release
```

`log::banner` function takes in one argument and print it with easy to spot in logs border:
```
2021/01/19 14:41:25 UTC [INFO] *************************************************************************************
2021/01/19 14:41:25 UTC [INFO] * Provision cluster: "gkeint-pr-9921-o0ihxwxf07"
2021/01/19 14:41:25 UTC [INFO] *************************************************************************************
```

`log::success`  takes in one argument and print it with easy to spot in logs border:
```
2021/01/18 10:15:49 UTC [INFO] =====================================================================================
2021/01/18 10:15:49 UTC [INFO] = SUCCESS                                                                           =
2021/01/18 10:15:49 UTC [INFO] =====================================================================================
2021/01/18 10:15:49 UTC [INFO] = Cleanup Azure Eventhubs Namespaces finished successfully
2021/01/18 10:15:49 UTC [INFO] =====================================================================================
```


### Utils library
`utils::check_required_vars` funciotn takes in one argument that holds a listo of variables and check if they are set. If at least one variable is unset, it is printed in the log, and script exists.
Example usage
```bash
requiredVars=(
    REPO_OWNER
    REPO_NAME
    DOCKER_PUSH_REPOSITORY
)

utils::check_required_vars "${requiredVars[@]}"
```

`utils::generate_self_signed_cert` creates self-signed certificate for the given domain.