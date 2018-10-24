plank:
  allow_cancellations: true # AllowCancellations enables aborting presubmit jobs for commits that have been superseded by newer commits in Github pull requests.
  pod_pending_timeout: 60m
  default_decoration_config:
    timeout: 7200000000000 # 2h
    grace_period: 15000000000 # 15s
    utility_images:
      clonerefs: "gcr.io/k8s-prow/clonerefs:v20181019-08e9d55c9"
      initupload: "gcr.io/k8s-prow/initupload:v20181019-08e9d55c9"
      entrypoint: "gcr.io/k8s-prow/entrypoint:v20181019-08e9d55c9"
      sidecar: "gcr.io/k8s-prow/sidecar:v20181019-08e9d55c9"
    gcs_configuration:
      bucket: {{ .Bucket }}
      path_strategy: "explicit"
    gcs_credentials_secret: "service-account"

triggers:
- repos:
  - {{ .OrganizationOrUser }}/kyma
  - {{ .OrganizationOrUser }}/console

presets:
  - labels:
      preset-service-account: "true"
    env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /etc/service-account/service-account.json
    volumes:
      - name: service
        secret:
          secretName: service-account
    volumeMounts:
      - name: service
        mountPath: /etc/service-account
        readOnly: true
  - labels:
      preset-dind-enabled: "true"
    env:
      - name: DOCKER_IN_DOCKER_ENABLED
        value: "true"
    volumes:
      - name: docker-graph
        emptyDir: {}
    volumeMounts:
      - name: docker-graph
        mountPath: /docker-graph
  - labels:
      preset-docker-push-repository: "true"
    env:
      - name: DOCKER_PUSH_REPOSITORY
        value: "eu.gcr.io/kyma-project/prow/test"
  - labels:
      preset-docker-pr-directory: "true"
    env:
      - name: DOCKER_PUSH_DIRECTORY
        value: "pr"

presets:
- labels:
    preset-compute-service-account: "true" # Service account with "Compute Admin" and "Compute OS Admin Login" roles
  env:
    - name: GOOGLE_APPLICATION_CREDENTIALS
      value: /etc/service-account/compute-service-account.json
  volumes:
  - name: compute-service-account
    secret:
      secretName: compute-service-account
  volumeMounts:
  - name: compute-service-account
    mountPath: /etc/service-account
    readOnly: true
  - labels:
      preset-service-account: "true"
    env:
      - name: GOOGLE_APPLICATION_CREDENTIALS
        value: /etc/service-account/service-account.json
    volumes:
      - name: service
        secret:
          secretName: service-account
    volumeMounts:
      - name: service
        mountPath: /etc/service-account
        readOnly: true
  - labels:
      preset-dind-enabled: "true"
    env:
      - name: DOCKER_IN_DOCKER_ENABLED
        value: "true"
    volumes:
      - name: docker-graph
        emptyDir: {}
    volumeMounts:
      - name: docker-graph
        mountPath: /docker-graph
  - labels:
      preset-docker-push-repository: "true"
    env:
      - name: DOCKER_PUSH_REPOSITORY
        value: "eu.gcr.io/kyma-project/prow/test"
  - labels:
      preset-docker-pr-directory: "true"
    env:
      - name: DOCKER_PUSH_DIRECTORY
        value: "pr"

presubmits: # runs on PRs
  {{ .OrganizationOrUser }}/kyma:
  - name: kyma-integration
    trigger: "(?m)^/test kyma-integration"
    rerun_command: "/test kyma-integration"
    context: kyma-integration
    skip_report: false # from documentation: SkipReport skips commenting and setting status on GitHub.
    max_concurrency: 10
    labels:
      preset-compute-service-account: "true"
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/snapshot/test/integration:0.0.1 # created by running `docker build -t <image> .` in the integration-job directory.
  - name: prow/components/ui-api-layer
    run_if_changed: "components/ui-api-layer/"
    context: prow/components/ui-api-layer
    skip_report: false # from documentation: SkipReport skips commenting and setting status on GitHub.
    labels:
        preset-dind-enabled: "true"
        preset-service-account: "true"
        preset-docker-push-repository: "true"
        preset-docker-pr-directory: "true"
    decorate: true
    path_alias: github.com/kyma-project/kyma
    extra_refs:
        - org: {{ .OrganizationOrUser }}
          repo: test-infra
          base_ref: master
          path_alias: github.com/kyma-project/test-infra
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/prow/buildpack-golang:0.0.1
        securityContext:
          privileged: true
        command:
          - "prow/presubmit-component.sh"
        args:
          - "--resolve-tests"
          - "--build-tests"
          - "--build-image-tests"
          - "--push-image-tests"
  {{ .OrganizationOrUser }}/console:
  - name: prow/content
    run_if_changed: "content/"
    context: prow/content
    skip_report: false # from documentation: SkipReport skips commenting and setting status on GitHub.
    labels:
        preset-dind-enabled: "true"
        preset-service-account: "true"
        preset-docker-push-repository: "true"
        preset-docker-pr-directory: "true"
    decorate: true
    path_alias: github.com/kyma-project/console
    extra_refs:
        - org: {{ .OrganizationOrUser }}
          repo: test-infra
          base_ref: master
          path_alias: github.com/kyma-project/test-infra
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/prow/buildpack-node:0.0.1
        securityContext:
          privileged: true
        command:
          - "prow/presubmit.sh"
        args:
          - "--resolve-tests"
          - "--validate-tests"
          - "--build-tests"
          - "--unit-tests"
          - "--build-image-tests"
          - "--push-image-tests"
  - name: prow/catalog
    run_if_changed: "catalog/"
    context: prow/catalog
    skip_report: false # from documentation: SkipReport skips commenting and setting status on GitHub.
    labels:
        preset-dind-enabled: "true"
        preset-service-account: "true"
        preset-docker-push-repository: "true"
        preset-docker-pr-directory: "true"
    decorate: true
    path_alias: github.com/kyma-project/console
    extra_refs:
        - org: {{ .OrganizationOrUser }}
          repo: test-infra
          base_ref: master
          path_alias: github.com/kyma-project/test-infra
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/prow/buildpack-node:0.0.1
        securityContext:
          privileged: true
        command:
          - "prow/presubmit.sh"
        args:
          - "--resolve-tests"
          - "--validate-tests"
          - "--build-tests"
          - "--unit-tests"
          - "--build-image-tests"
          - "--push-image-tests"
  - name: prow/instances
    run_if_changed: "instances/"
    context: prow/instances
    skip_report: false # from documentation: SkipReport skips commenting and setting status on GitHub.
    labels:
        preset-dind-enabled: "true"
        preset-service-account: "true"
        preset-docker-push-repository: "true"
        preset-docker-pr-directory: "true"
    decorate: true
    path_alias: github.com/kyma-project/console
    extra_refs:
        - org: {{ .OrganizationOrUser }}
          repo: test-infra
          base_ref: master
          path_alias: github.com/kyma-project/test-infra
    spec:
      containers:
      - image: eu.gcr.io/kyma-project/prow/buildpack-node:0.0.1
        securityContext:
          privileged: true
        command:
          - "prow/presubmit.sh"
        args:
          - "--resolve-tests"
          - "--validate-tests"
          - "--build-tests"
          - "--unit-tests"
          - "--build-image-tests"
          - "--push-image-tests"