# Github Release

## Overview

This command collects artifacts and create github release. Application expects artifacts to be stored in Google bucket:
- `kyma-config-cluster.yaml`
- `kyma-config-local.yaml`
- `release-changelog.md`

## Usage

To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \ 
    -targetCommit={commitish where the Github tag will be created} \
    -githubRepoOwner={the Github repository owner} \
    -githubRepoName={the Github repository name} \
    -githubAccessToken={the Github oauth2 access token} \
    -releaseVersionFilePath={the full path to the RELEASE_VERSION file} 
```

### Flags

See the list of available flags:

| Name                           | Required | Description                                                                                          |
| :----------------------------- | :------: | :--------------------------------------------------------------------------------------------------- |
| **--targetCommit**             |   Yes    | The string value with commitish where the Github tag will be created
| **--bucketName**               |    No    | The string value with a name of Google bucket containing release artifacts. It defaults to `kyma-prow-artifacts`.
| **--kymaConfigCluster**        |    No    | The string value with a name of a kyma cluster configuration file. It defaults to `kyma-config-cluster.yaml`.
| **--kymaConfigLocal**          |    No    | The string value with a name of a kyma local configuration file. It defaults to `kyma-config-local.yaml`.
| **--kymaChangelog**            |    No    | The string value with a name of a release changelog file. It defaults to `release-changelog.md`.
| **--githubRepoOwner**          |   Yes    | The string value with a name of the Github repository owner.
| **--githubRepoName**           |   Yes    | The string value with a name of the Github repository name.
| **--githubAccessToken**        |   Yes    | The string value with a name of the Github oath2 access token.
| **--releaseVersionFilePath**   |   Yes    | The string value with a full path to the RELEASE_VERSION file.

### Environment variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least storage roles. |
