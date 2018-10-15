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
- There is no possibility to test Prow configuration locally, on minikube. All tests needs to be done on clusters. 
- Avoid provisioning long-running clusters.
- Test Prow configuration against your kyma fork repository.
- Disable build on internal CI only if all CI functionality are provided by Prow, not only for master branch, but also for releases branches.
