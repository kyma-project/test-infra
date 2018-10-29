plugins:
  {{ .OrganizationOrUser }}/kyma:
    - cat
    - trigger
  {{ .OrganizationOrUser }}/console:
    - cat
    - trigger
  {{ .OrganizationOrUser }}/examples:
    - cat
    - trigger
  {{ .OrganizationOrUser }}/test-infra:
    - config-updater
    - cat
    - trigger
  {{ .OrganizationOrUser }}/bundles:
    - cat
    - trigger
  {{ .OrganizationOrUser }}/community:
    - cat
    - trigger
  {{ .OrganizationOrUser }}/website:
    - cat
    - trigger
  {{ .OrganizationOrUser }}/luigi:
    - cat
    - trigger
