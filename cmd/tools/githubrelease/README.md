# GitHub Release

## Overview

This command creates GitHub releases.

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

| Name | Required | Description |
|---|:---:|---|
| **--targetCommit** | Yes | The string value which specifies the [commitish value](https://docs.github.com/en/rest/releases/releases#create-a-release) that the GitHub tag refers to.|
| **--kymaComponentsPath**  | No | The string value with the path to the Kyma `components.yaml`. <br> It defaults to `installation/resources/components.yaml`.|                        
| **--kymaChangelog** | No | The string value with the name of the release changelog file. <br> It defaults to `release-changelog.md`.|                                                      
| **--githubRepoOwner** | Yes | The string value with the name of the GitHub repository owner.|                                                                                            
| **--githubRepoName** | Yes| The string value with the name of the GitHub repository.|                                                                                                
| **--githubAccessToken** | Yes | The string value with the name of the GitHub OAuth2 access token.|                                                                                        
| **--releaseVersionFilePath** | Yes | The string value with the full path to the `RELEASE_VERSION` file.|                                                                                       

### Environment Variables

Available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least storage roles. |
