# Crier

Crier reports ProwJob status changes. For now it is responsible for Slack notifications as Plank is still reporting ProwJob statuses to the Github.

## Available reporters

Crier supports multiple reporters, each reporter will become a crier controller. Reporters that can can be used:
- Github reporter
- Slack reporter
- PubSub reporter

For any reporter you want to use, you need to mount your prow configs and specify `--config-path` and `--job-config-path` flag.

### GitHub reporter

You can enable github reporter in crier by specifying `--github-workers=N` flag.

You also need to mount a github oauth token by specifying `--github-token-path` flag, which defaults to `/etc/github/oauth`.

If you have a [ghproxy] deployed, also remember to point `--github-endpoint` to your ghproxy to avoid token throttle.

### Slack reporter

> **NOTE:** when you enable Crier for the first time it will message to the Slack all ProwJobs matching the configured filtering criteria.

You can enable the Slack reporter in crier by specifying the `--slack-workers` and `--slack-token-file` flags.

The `--slack-token-file` flag takes a path to a file containing a Slack [**OAuth Access Token**](https://api.slack.com/docs/oauth).

The **OAuth Access Token** can be obtained as follows:

1. Navigate to: https://api.slack.com/apps.
1. Click **Create New App**.
1. Provide an **App Name** (e.g. Prow Slack Reporter) and **Development Slack Workspace** (e.g. Kubernetes).
1. Click **Permissions**.
1. Add the `chat:write.public` scope using the **Scopes / Bot Token Scopes** dropdown and **Save Changes**.
1. Click **Install App to Workspace**
1. Click **Allow** to authorize the Oauth scopes.
1. Copy the **OAuth Access Token**.

Once the **access token** is obtained, you can create a `secret` in the cluster using that value:

```shell
kubectl create secret generic slack-token --from-literal=token=< access token >
```

Furthermore, to make this token available to **Crier**, mount the *slack-token* `secret` using a `volume` and set the `--slack-token-file` flag in the deployment spec.

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

> **NOTE:** `slack_reporter_configs` is a map of `org`, `org/repo`, or `*` wildcard to a set of slack reporter configs.

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

The Slack `channel` can be overridden at the ProwJob level via the `reporter_config.slack.channel` field:
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

## Current Slack notification settings

Crier does not send Slack notifications at all for presubmit jobs.

Reporter config:
```
job_types_to_report:
  - postsubmit
  - periodic
  - batch
```

If you don't want to report postsubmit or periodic jobs to report to Slack channel use `skip_report:true`.
If the job is still in testing phase we can set `optional: true`.

## Migration from plank for github report

First, you need to disable GitHub reporting in Plank, add the `--skip-report=true` flag to the Plank [deployment](https://github.com/kyma-project/test-infra/blob/master/prow/cluster/components/11-plank_deployment.yaml).

Before migrating, be sure plank is setting the [PrevReportStates field](https://github.com/kubernetes/test-infra/blob/de3775a7480fe0a724baacf24a87cbf058cd9fd5/prow/apis/prowjobs/v1/types.go#L566)
by describing a finished presubmit ProwJob. Plank started to set this field after commit [2118178](https://github.com/kubernetes/test-infra/pull/10975/commits/211817826fc3c4f3315a02e46f3d6aa35573d22f), if not, you want to upgrade your plank to a version includes this commit before moving forward.

Flags required by Crier:
- Point `config-path` and `--job-config-path` to your prow config and job configs accordingly.
- Set `--github-worker` to be number of parallel github reporting threads you need.
- Point `--github-endpoint` to ghproxy, if you have set that for plank.
- Bind github oauth token as a secret and set `--github-token-path` if you've have that set for plank.

In your plank deployment, you can:
- Remove the `--github-endpoint` flag.
- Remove the github oauth secret, and `--github-token-path` flag if set.
- Add`--skip-report`, so plank will skip the reporting logic.

Both change should be deployed at the same time, if have an order preference, deploy crier first since report twice should just be a no-op.
