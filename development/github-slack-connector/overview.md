# Overview


This Add-On allows you to communicate with GitHub. You can handle Events incoming from GitHub repositories or manage repositories through the GitHub API. You must provision an instance for every repository you want to communicate with.
#### Prerequisites

- {GITHUB_TOKEN} - it can be generated [here](https://github.com/settings/tokens/new).
> **NOTE**: In case of generating token for an organization the [OAuth or GitHub App](https://developer.github.com/apps/) is needed.


#### Installation

1. Provision the GitHub Connector Add-On. Plans' and fields' meaning is explained below.
2. Go to `Service Management > Catalog > Services`. Find a service named `github-{REPOSITORY-NAME}` and add it.

Now you can start using the GitHub Connector. Add a new Event trigger to react to chosen GitHub notifications or bind this service in a lambda to send authorized request to the GitHub API.

## Provisioning

### Default plan

This plan allows to both handle Events incoming from connected GitHub repositories to an exposed endpoint and POST jsons to the GitHub API through Application Gateway, which automatically adds all necessary informations needed to communicate with GitHub.

### Fields

| PARAMETER NAME | DISPLAY NAME | TYPE | DESCRIPTION | REQUIRED |
| -------------- | ------------ | ---- | ----------- | :------: |
| `githubToken` | Token | `string` | {GITHUB_TOKEN} | yes |
| `githubEndpoint` | GitHub Endpoint (organization or repository) | `string` | Link to a GitHub repository in the proper format: repos/{OWNER}/{REPOSITORY} or orgs/{ORGANIZATION}. For example, "repos/kyma-incubator/github-slack-connectors". | yes |
| `kymaAddress` | Kyma Domain name | `string` | Kyma domain address in the proper format. For example, "domain.sap.com". | yes |
