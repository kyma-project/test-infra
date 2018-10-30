presets:
- labels:
    preset-sa-vm-kyma-integration: "true" # Service account with "Compute Admin" and "Compute OS Admin Login" roles
  env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /etc/service-account/sa-vm-kyma-integration
  volumes:
  - name: sa-vm-kyma-integration
    secret:
      secretName: sa-vm-kyma-integration
  volumeMounts:
  - name: sa-vm-kyma-integration
    mountPath: /etc/service-account
    readOnly: true

presubmits: # runs on PRs
  {{ .OrganizationOrUser }}/kyma:
  - name: prow/components/ui-api-layer
    optional: true
    run_if_changed: "components/ui-api-layer"
    context: prow/components/ui-api-layer
    skip_report: true # from documentation: SkipReport skips commenting and setting status on GitHub.
    spec:
      containers:
      - image: alpine
        command: ["/bin/printenv"]
  - name: kyma-integration
    optional: true
    run_if_changed: "^(resources|installation)"
    trigger: "(?m)^/test kyma-integration"
    rerun_command: "/test kyma-integration"
    context: kyma-integration
    skip_report: true # from documentation: SkipReport skips commenting and setting status on GitHub.
    max_concurrency: 10
    labels:
      preset-sa-vm-kyma-integration: "true"
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/snapshot/test/integration:0.0.1 # created by running `docker build -t <image> .` in the integration-job directory.
  - name: kyma-gke-integration
    run_if_changed: "^(resources|installation)"
    trigger: "(?m)^/test kyma-gke-integration"
    rerun_command: "/test kyma-gke-integration"
    context: kyma-gke-integration
    optional: true
    skip_report: true # from documentation: SkipReport skips commenting and setting status on GitHub.
    max_concurrency: 10
    labels:
      preset-sa-vm-kyma-integration: "true"
    spec:
      containers:
      - image: alpine
        command: ["/bin/echo"]
        args: ["starting fake gke test integration job"]

postsubmits:
  {{ .OrganizationOrUser }}/kyma:
  - name: kyma-integration-master
    branches:
    - master
    max_concurrency: 10
    labels:
      preset-sa-vm-kyma-integration: "true"
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/snapshot/test/integration:0.0.1 # created by running `docker build -t <image> .` in the integration-job directory.

plank:
  allow_cancellations: true # AllowCancellations enables aborting presubmit jobs for commits that have been superseded by newer commits in Github pull requests.
