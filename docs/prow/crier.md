# Crier

Crier reports the Prow Job status changes. For now, it is responsible for Slack notifications as Plank is still reporting the Prow Job statuses to GitHub.

## Available Reporters

Crier supports multiple reporters. Each reporter will become a Crier controller. Reporters that can be used:
- GitHub reporter
- Slack reporter
- PubSub reporter

For any reporter that you want to use, you must mount your Prow configs and specify the `--config-path` and `--job-config-path` flags.

### GitHub Reporter

You can enable the GitHub reporter in Crier by specifying the `--github-workers=N` flag.

You must also mount a GitHub OAuth token by specifying the `--github-token-path` flag, which defaults to `/etc/github/oauth`.

If you have a ghproxy deployed, also remember to point `--github-endpoint` to your ghproxy to avoid token throttle.

### Slack Reporter

> **NOTE:** When you enable Crier for the first time, it will sent to Slack all Prow Jobs matching the configured filtering criteria.

You can enable the Slack reporter in Crier by specifying the `--slack-workers` and `--slack-token-file` flags.

The `--slack-token-file` flag takes the path to the file containing the Slack [**OAuth Access Token**](https://api.slack.com/docs/oauth).

The **OAuth Access Token** can be obtained as follows:

1. Navigate to [`https://api.slack.com/apps`](https://api.slack.com/apps).
1. Click **Create New App**.
1. Provide the **App Name** (e.g. `Prow Slack Reporter`) and **Development Slack Workspace** (e.g. `Kubernetes`).
1. Click **Permissions**.
1. Add the `chat:write.public` scope using the **Scopes / Bot Token Scopes** dropdown and click **Save Changes**.
1. Click **Install App to Workspace**.
1. Click **Allow** to authorize the OAuth scopes.
1. Copy the **OAuth Access Token**.

Once the access token is obtained, you can create a Secret in the cluster using that value:

```shell
kubectl create secret generic slack-token --from-literal=token="{ACCESS_TOKEN}"
```

Furthermore, to make this token available to Crier, mount the `slack-token` Secret as a volume and set the `--slack-token-file` flag in the deployment spec:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: crier
  labels:
    app: crier
spec:
  selector:
    matchLabels:
      app: crier
  template:
    metadata:
      labels:
        app: crier
    spec:
      containers:
      - name: crier
        image: gcr.io/k8s-prow/crier:v20200205-656133e91
        args:
        - --slack-workers=1
        - --slack-token-file=/etc/slack/token
        - --config-path=/etc/config/config.yaml
        - --dry-run=false
        volumeMounts:
        - mountPath: /etc/config
          name: config
          readOnly: true
        - name: slack
          mountPath: /etc/slack
          readOnly: true
      volumes:
      - name: slack
        secret:
          secretName: slack-token
      - name: config
        configMap:
          name: config
```

Additionally, in order for it to work with Prow you must add the following to your `config.yaml`:

> **NOTE:** `slack_reporter_configs` is a map of the `org`, `org/repo`, or `*` wildcard to a set of Slack reporter configs.

```yaml
slack_reporter_configs:

  # Wildcard (i.e. catch-all) slack config
  "*":
    # default: None
    job_types_to_report:
      - presubmit
      - postsubmit
    # default: None
    job_states_to_report:
      - failure
      - error
    # required
    channel: my-slack-channel
    # The template shown below is the default
    report_template: "Job {{.Spec.Job}} of type {{.Spec.Type}} ended with state {{.Status.State}}. <{{.Status.URL}}|View logs>"

  # "org/repo" slack config
  istio/proxy:
    job_types_to_report:
      - presubmit
    job_states_to_report:
      - error
    channel: istio-proxy-channel

  # "org" slack config
  istio:
    job_types_to_report:
      - periodic
    job_states_to_report:
      - failure
    channel: istio-channel
```

The Slack channel can be overridden at the Prow Job level via the **reporter_config.slack.channel** field:

```yaml
postsubmits:
  some-org/some-repo:
    - name: example-job
      decorate: true
      reporter_config:
        slack:
          channel: 'override-channel-name'
      spec:
        containers:
          - image: alpine
            command:
              - echo
```

## Current Slack Notification Settings

Crier does not send any Slack notifications for presubmit jobs.

Reporter config:

```
job_types_to_report:
  - postsubmit
  - periodic
  - batch
```

If you don't want to configure postsubmit or periodic jobs to report to a Slack channel, use `skip_report:true`.
If the job is still in the testing phase, you can set `optional: true`.

## Migration from Plank to GitHub Reporter

First, you need to disable GitHub reporting in Plank. To do that, add the `--skip-report=true` flag to the Plank deployment.

Before migrating, upgrade your Plank to a version that includes the commit [`2118178`](https://github.com/kubernetes/test-infra/pull/10975/commits/211817826fc3c4f3315a02e46f3d6aa35573d22f).

Flags required by Crier:
- Point `--config-path` and `--job-config-path` to your Prow config and job configs accordingly.
- Set `--github-worker` to be the number of parallel GitHub reporting threads you need.
- Point `--github-endpoint` to ghproxy, if you have set that for Plank.
- Bind GitHub OAuth token as a Secret and set `--github-token-path` if you have that set for Plank.

In your Plank deployment, you must:
- Remove the `--github-endpoint` flag.
- Remove the GitHub OAuth Secret and the `--github-token-path` flag if set.
- Add `--skip-report`, so Plank will skip the reporting logic.

Both changes should be deployed at the same time. If you need to deploy them sequentially, deploy Crier first to avoid double-reporting.
