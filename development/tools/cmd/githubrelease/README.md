# Github Release

## Overview

This command creates Github releases based on artifacts stored in a Google bucket. Each release requires the following set of artifacts:
- `kyma-config-cluster.yaml`
- `kyma-config-local.yaml`
- `release-changelog.md`

## Usage

To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to a service account file} go run main.go \ 
    -targetCommit={a commitish that the Github tag will refer to} \
    -githubRepoOwner={the Github repository owner} \
    -githubRepoName={the Github repository name} \
    -githubAccessToken={the Github oauth2 access token} \
    -releaseVersionFilePath={the full path to the RELEASE_VERSION file} 
```

### Flags

See the list of available flags:

| Name                           | Required | Description                                                                                          |
| :----------------------------- | :------: | :--------------------------------------------------------------------------------------------------- |
| **--targetCommit**             |   Yes    | The string value with a commitish that the Github tag will refer to.
| **--bucketName**               |    No    | The string value with a name of a Google bucket containing release artifacts. It defaults to `kyma-prow-artifacts`.
| **--kymaConfigCluster**        |    No    | The string value with a name of a Kyma cluster configuration file. It defaults to `kyma-config-cluster.yaml`.
| **--kymaConfigLocal**          |    No    | The string value with a name of a Kyma local configuration file. It defaults to `kyma-config-local.yaml`.
| **--kymaChangelog**            |    No    | The string value with a name of the release changelog file. It defaults to `release-changelog.md`.
| **--githubRepoOwner**          |   Yes    | The string value with a name of the Github repository owner.
| **--githubRepoName**           |   Yes    | The string value with a name of the Github repository name.
| **--githubAccessToken**        |   Yes    | The string value with a name of the Github oath2 access token.
| **--releaseVersionFilePath**   |   Yes    | The string value with the full path to the RELEASE_VERSION file.

### Environment variables

Available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to a service account file. The service account requires at least storage roles. |
