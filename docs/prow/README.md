# Docs

## Overview

The folder contains documents that provide an insight into Prow configuration, development, and testing.

<!-- Update the list each time you modify the document structure in this folder. -->

Read the documents to learn how to:

- [Configure Prow on a production cluster](./production-cluster-configuration.md) based on the preconfigured Google Cloud Storage (GCS) resources.
- [Create a service account](./prow-secrets-management.md) and store its encrypted key in a GCS bucket.
- [Install and configure Prow](./prow-installation-on-forks.md) on a forked repository to test and develop it on your own.
- [Install and manage monitoring](./prow-monitoring.md) on a Prow cluster.
- [Create, modify, and remove component jobs using templates](./manage-component-jobs-with-templates.md) for the Prow pipeline.
- [Update](./prow-cluster-update.md) a Prow cluster.
- [Run kind jobs manually](./kind-jobs.md) in a local environment.

Find out more about:

- [Prow architecture](./prow-architecture.md) and its setup in the Kyma project.
- [Prow jobs](./prow-jobs.md) for details on Prow jobs.
- [Naming convention of the Prow test instance](./prow-naming-convention.md)
- [Prow jobs on TestGrid](./prow-k8s-testgrid.md) for details on how to add jobs to the TestGrid dashboard.
- [Prow test clusters](./test-clusters.md) for details on permissions for tests clusters.
- [Obligatory security measures](./obligatory-security-measures.md) to take regularly for the Prow production cluster and when someone leaves the Kyma project.
- [Presets](./presets.md) you can use to define Prow jobs.
- [Authorization](./authorization.md) concepts employed in Prow.
