# test-infra

## Overview

Test infrastructure for the Kyma project.

## Prerequisites

No prerequisites at the moment.

## Installation

No installation instructions at the moment.

## Usage

No usage instructions at the moment.

## Development
- You cannot test Prow configuration locally on Minikube. Perform all the tests on the cluster. 
- Avoid provisioning long-running clusters.
- Test Prow configuration against your `kyma` fork repository.
- Disable builds on the internal CI only after all CI functionalities are provided by Prow. This applies not only for the `master` branch but also for release branches.
