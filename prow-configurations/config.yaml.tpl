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

plank:
  allow_cancellations: true # AllowCancellations enables aborting presubmit jobs for commits that have been superseded by newer commits in Github pull requests.
