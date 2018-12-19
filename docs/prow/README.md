# Docs

## Overview

The folder contains documents that provide an insight into Prow configuration, development, and testing.

<!-- Update the list each time you modify the document structure in this folder. -->

Read the documents to learn how to:

- [Configure Prow on a production cluster](./production-cluster-configuration.md) based on the preconfigured Google Cloud Storage (GCS) resources.
- [Create a service account](./prow-secrets-management.md) and store its encrypted key in a GCS bucket.
- [Install and configure Prow](./prow-installation-on-forks.md) on a forked repository to test and develop it on your own.
- [Install and manage monitoring](./prow-monitoring.md) on a Prow cluster.
- [Prepare your component for the migration](./migration-guide.md) from the internal CI to Prow.
- [Define a release job for your component](./migration-guide-release.md)

Read also:

 - [Prow architecture](./prow-architecture.md) and its set up in the Kyma project.
 - The proposal for the new [release process](./kyma-release-process.md) in Kyma that uses Prow. See how it differs from the release process with the internal CI.
