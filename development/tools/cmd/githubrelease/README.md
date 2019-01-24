# Github Release

## Overview

This command creates GitHub releases based on artifacts stored in a Google bucket. Each release requires the following set of artifacts:
- `kyma-config-cluster.yaml`
- `kyma-installer-cluster.yaml`
- `kyma-config-local.yaml`
- `kyma-installer-local.yaml`
- `release-changelog.md`

## Usage

To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to the service account file} go run main.go \ 
    -targetCommit={commitish value that the GitHub tag refers to} \
    -githubRepoOwner={GitHub repository owner} \
    -githubRepoName={GitHub repository name} \
    -githubAccessToken={GitHub OAuth2 access token} \
    -releaseVersionFilePath={full path to the RELEASE_VERSION file} 
```

### Flags

See the list of available flags:

| Name                           | Required | Description                                                                                          |
| :----------------------------- | :------: | :--------------------------------------------------------------------------------------------------- |
| **--targetCommit**             |   Yes    | The string which specifies the [commitish value](https://developer.github.com/v3/repos/releases/#create-a-release) that the GitHub tag refers to.
| **--bucketName**               |    No    | The string value with the name of the Google bucket containing release artifacts. It defaults to `kyma-prow-artifacts`.
| **--kymaConfigCluster**        |    No    | The string value with the name of the Kyma cluster configuration file. It defaults to `kyma-config-cluster.yaml`.
| **--kymaInstallerCluster**     |    No    | The string value with the name of the Kyma cluster configuration file. It defaults to `kyma-installer-cluster.yaml`.
| **--kymaConfigLocal**          |    No    | The string value with the name of the Kyma local configuration file. It defaults to `kyma-config-local.yaml`.
| **--kymaInstallerLocal**       |    No    | The string value with the name of the Kyma local configuration file. It defaults to `kyma-installer-local.yaml`.
| **--kymaChangelog**            |    No    | The string value with the name of the release changelog file. It defaults to `release-changelog.md`.
| **--githubRepoOwner**          |   Yes    | The string value with the name of the GitHub repository owner.
| **--githubRepoName**           |   Yes    | The string value with the name of the GitHub repository.
| **--githubAccessToken**        |   Yes    | The string value with the name of the GitHub OAuth2 access token.
| **--releaseVersionFilePath**   |   Yes    | The string value with the full path to the `RELEASE_VERSION` file.

### Environment variables

Available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least storage roles. |
