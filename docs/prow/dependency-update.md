# Dependency updates

To help keep dependency up to date you can use Dependabot together with automated testing and merging of pull requests.

Dependabot is a GitHub tool that can monitor your repository and create pull requests to update dependencies. Dependabot supports multiple languages. Check the Dependabot documentation for details.

Pull requests created by Dependabot will be tested by Prow. If the pull requests passed the required tests and meet the criteria, they can be automatically merged by Tide without human review.

Owners of the updated files can subscribe to receive notifications when a pull request is merged without human review.

## How to use it

1. Enable Dependabot in your repository. Learn more under [About Dependabot version updates](https://docs.github.com/en/code-security/dependabot/dependabot-version-updates/about-dependabot-version-updates).

2. To enable Dependabot, create `dependabot.yml` configuration file. You can use a rendertemplate tool to create the `dependabot.yml` file. Use [kyma-project/test-infra](../../.github/dependabot.yml) configuration as an example.

3. Copy the [dependabot.tmpl](https://github.com/kyma-project/test-infra/blob/main/templates/templates/dependabot.tmpl) file to the `.github` directory in your repository.

4. Create a `dependabot-data.yaml` file in the `.github` directory in your repository. Use the [test-infra](https://github.com/kyma-project/kyma/blob/main/.github/dependabot-config.yaml) data file as an example.

5. If you want to merge Dependabot pull requests without human review, add `autoMerge` local set for your ecosystems in the data file:

   localSets:
     autoMerge:
       labels:
       - "skip-review"

6. To generate the Dependabot configuration file, run rendertemplate with the following flags:

`go run cmd/rendertemplates/main.go --config ../../../kyma/.github/dependabot-config.yaml --skip-data-dir --appendSlice`

| Flag Name       | Description  |
|-----------------|--------------|
| --config        | Path to `dependabot-data.yml` |
| --skip-data-dir | Skip loading data files from data directory |
| --appendSlice   | Append slice instead overwrite |

7. If you want to get a notification every time a file for which you are an owner is updated, enable notifications in the `aliases-map.yaml` or `users-map.yaml` files.

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

Dependabot creates pull requests according to the `dependabot.yml` configuration file located in `.github` directory in repository root. Dependabot labels pull requests with labels defined in the configuration file. If pull requests are labeled with the `skip-review` label, they are merged after all tests pass.

Pull requests created by dependabot are tested by Prow in the same way as any other pull request. Once all tests pass, pull request will be merged by Tide. Merging require human approval or if pull request has `skip-review` label, Tide will merge it automatically without human approval.

After a pull request has been merged by Dependabot, Prow `automerge-notification` plugin receives a GitHub webhook and sends notifications about automated merging based on the notification settings in the `users-map.yaml` and `aliases-map.yaml` files.
