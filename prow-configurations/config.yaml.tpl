triggers:
- repos:
  - {{ .OrganizationOrUser }}/kyma

presubmits: # runs on PRs
  {{ .OrganizationOrUser }}/kyma:
  - name: prow/components/ui-api-layer
    run_if_changed: "components/ui-api-layer"
    context: prow/components/ui-api-layer
    skip_report: false # from documentation: SkipReport skips commenting and setting status on GitHub.
    spec:
      containers:
      - image: alpine
        command: ["/bin/printenv"]
