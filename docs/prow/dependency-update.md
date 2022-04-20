# Dependency updates

To help keep dependency up to date you can use dependabot together with automated testing and merging of pull requests.

Dependabot is a GitHub tool which can monitor repository and create pull requests to update dependencies. Dependabot support multiple languages. Check dependabot documentation for details.

Pull requests created by dependabot will be tested by Prow. Pull requestes with required tests passed and meeting criteria can be automatically merged by Tide without human review.

Owners of updated files can subscribe to receive notifications when pull request is merged without human review.

## How to use it

Enable dependabot in your repository. [Here](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/about-dependabot-version-updates) is detailed dependabot documentation.

To enable dependabot create dependabot.yml configuration file. To make it easier, you can use a rendertemplate tool to create dependabot.yml file. Use [kyma-project/test-infra](../../.github/dependabot.yml) configuration as an example.

Copy [dependabot.tmpl](https://github.com/kyma-project/test-infra/blob/main/templates/templates/dependabot.tmpl) file to .github directory in your repository.

Create dependabot-data.yaml file in .github directory in your repository. Use [test-infra](https://github.com/kyma-project/kyma/blob/main/.github/dependabot-config.yaml) data file as an example.

If you want to merge dependabot pull request without human review add `autoMerge` local set for your ecosystems in data file.

```
localSets:
  autoMerge:
    labels:
    - "skip-review"
```

Run rendertemplate with following flags to generate dependabot configuration file.

`go run cmd/rendertemplates/main.go --config ../../../kyma/.github/dependabot-config.yaml --skip-data-dir --appendSlice`

| flag name       | description  |
|-----------------|--------------|
| --config        | path to dependabot-data.yml |
| --skip-data-dir | skip loading data files from data directory |
| --appendSlice   | append slice instead overwrite |

You can get notification every time a file for which you are an owner is updated. To make it happen enable notifications in `aliases-map.yaml` or `users-map.yaml` files.

Update `aliases-map.yaml` file to send notification to yours team channel. If configuration for alias is not present, or `automerge.notification` is set to `false`, system will find alias members and check configuration in `users-map.yaml` for each user.

```
- com.github.aliasname: "github.com owners name"
  automerge.notification: true
  com.slack.enterprise.sap.groupsnames:
    - "slack team group name"
  com.slack.enterprise.sap.channelsnames:
    - "slack team channel name"
```

If you don't notify on yours team channel, update `users-map.yaml` file to send notifications to you directly. This file will be consulted if your github user is directly set as an owner too.

```
- com.github.username: "username"
  sap.tools.github.username: "username"
  com.slack.enterprise.sap.username: "slack ID"
  automerge.notification: true
```

## How it works

Dependabot creates pull requests according to the `dependabot.yml` configuration file located in `.github` directory in repository root. Dependabot will label pull request with labels defined in configuration file. If pull requests are labeled with `skip-review` label, it will be merged once all tests pass.

Pull requests created by dependabot are tested by Prow in the same way as any other pull request. Once all tests pass, pull request will be merged by Tide. Merging require human approval or if pull request has `skip-review` label, Tide will merge it automatically without human approval.

After dependabot pull request merging, Prow `automerge-notification` plugin will receive github webhook. Based on notification settings found in `users-map.yaml` and `aliases-map.yaml` files it will send notifications about automated merging.
