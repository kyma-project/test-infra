# TestGrid

Kyma publishes Prow job results on the [**Kubernetes TestGrid**](https://testgrid.k8s.io/kyma-all) that the Kubernetes team runs. Prow job configuration for Kyma is stored in two separate projects, in Kyma and Kubernetes `test-infra` repositories.

## Kyma configuration

One part of the configuration is based on job definitions in the form of annotations. See the example of the [`kind` job](https://github.com/kyma-project/test-infra/blob/60493dd61d77da363b8758b7e4c94f25d4b36501/prow/jobs/test-infra/test-infra-kind.yaml#L80-L83) displayed on the *kyma-nightly* dashboard on TestGrid:

```yaml
  annotations:
    testgrid-dashboards: kyma-nightly
    # testgrid-alert-email: email-here@sap.com
    testgrid-num-failures-to-alert: '1'
```

You can also specify an email address if you want to receive notifications after a predefined number of job failures.

## Kubernetes configuration

The other part of the configuration is stored in the [Kubernetes repository](https://github.com/kubernetes/test-infra/tree/master/config/testgrids). The Kyma team [owns](https://github.com/kubernetes/test-infra/blob/master/config/testgrids/kyma/OWNERS) the `kyma` folder which holds the configuration related to Kyma Prow jobs.

This configuration contains these important definitions:

1. A dashboard group called **kyma** and a number of sub-dashboards:

    ```yaml
    dashboard_groups:
    - name: kyma
      dashboard_names:
      - kyma-all
      - kyma-release
      - kyma-cleaners
      - kyma-presubmit
      - kyma-postsubmit
      - kyma-nightly
      - kyma-weekly
      - kyma-incubator
    ```

2) Actual dashboard definitions. For example, a part of the configuration for the [nightly dashboard](https://github.com/kubernetes/test-infra/blob/8737414459c84bdefdbb279caef5c8339033da69/config/testgrids/kyma/kyma.yaml#L355) looks as follows:

    ```yaml
    dashboards:
    - name: kyma-nightly
      dashboard_tab:
      - name: Kyma kind integration
        description: Tracks that Kyma is able to run on kind
        test_group_name: kyma-kind-integration
        # alert_options:
        #   alert_mail_to_addresses: email-here@sap.com
        code_search_url_template:
        url: https://github.com/kyma-project/kyma/compare/<start-custom-0>...<end-custom-0>
        open_bug_template:
        url: https://github.com/kyma-project/kyma/issues/
    ```
    The list of dashboards also includes the **kyma-all** dashboard which repeats all dashboard definitions and contains an overview of all jobs.

3) [**test_groups**](https://github.com/kubernetes/test-infra/blob/8737414459c84bdefdbb279caef5c8339033da69/config/testgrids/kyma/kyma.yaml#L422) defined at the end of the `kyma.yaml` file. These groups are used on the **dashboard_tab** and need to be defined for a job to show up on the dashboard. In the `kind` job example, the integration test has the test group defined as `kyma-kind-integration`. You can see the same definition of this test group on the list:

    ```yaml
    - name: kyma-kind-integration
      gcs_prefix: kyma-prow-logs/logs/kyma-kind-integration
      num_failures_to_alert: 1
    ```

    This definition specifies the source of information for the dashboard and gives a link to the job details if needed.

## Add a new job to the dashboard

Follow these steps to add a newly created job to the TestGrid dashboard:

1. Add the required annotations to the specific [Prow job](https://github.com/kyma-project/test-infra/tree/master/prow/jobs) in Kyma, specifying the dashboard on which the job results should appear. If you add a new dashboard, make sure to also add it to the `kyma.yaml` file in the [Kubernetes repository](https://github.com/kubernetes/test-infra/blob/8737414459c84bdefdbb279caef5c8339033da69/config/testgrids/kyma/kyma.yaml).
2. Add corresponding changes to the `kyma.yaml` file in the [Kubernetes repository](https://github.com/kubernetes/test-infra/blob/8737414459c84bdefdbb279caef5c8339033da69/config/testgrids/kyma/kyma.yaml). One of the [code owners](https://github.com/kubernetes/test-infra/blob/master/config/testgrids/kyma/OWNERS) of the `kyma` folder must approve your changes.
