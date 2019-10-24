## TestGrid

Due to **kubernetes TestGrid** not being fully Open Source yet, we're pushing our job results to the [instance](https://testgrid.k8s.io/kyma-all) that the Kubernetes Team is running. This currently has us put configuration in two separate projects. 

### Kyma configuration

One part of the configuration lives on our job definitions in the form of annotations as can be seen on the [**kind** job](https://github.com/kyma-project/test-infra/blob/60493dd61d77da363b8758b7e4c94f25d4b36501/prow/jobs/test-infra/test-infra-kind.yaml#L80-L83) for example: 
```yaml
  annotations:
    testgrid-dashboards: kyma-nightly
    # testgrid-alert-email: email-here@sap.com
    testgrid-num-failures-to-alert: '1'
```
The above definition will put the job onto the *kyma-nightly* dashboard on TestGrid and would send an email after one failed attempt to an email address if we would've specified one.

### Kubernetes configuration

The other part of the configuration is in the [Kubernetes repository](https://github.com/kubernetes/test-infra/tree/master/config/testgrids). Here we [*own*](https://github.com/kubernetes/test-infra/blob/master/config/testgrids/kyma/OWNERS) the Kyma folder which holds all of our configuration.

This configuration itself has three important things.
1) It's defining a dashboard group called **kyma**, which has *sub-dashboards*:
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
2) It holds the actual dashboard definitions. E.g. for the [nightly dashboard](https://github.com/kubernetes/test-infra/blob/8737414459c84bdefdbb279caef5c8339033da69/config/testgrids/kyma/kyma.yaml#L355), the configuration partially looks like this:
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

    Notice that in the list of dashboards, there's also a **kyma-all** dashboard. This dashboard repeats all definitions of the other dashboards and has an overview of all jobs. We actually have to repeat the configuration for the *dashboard-tab* with the only difference being und which dashboard they show.
3) At the very bottom of the file [**test_groups**](https://github.com/kubernetes/test-infra/blob/8737414459c84bdefdbb279caef5c8339033da69/config/testgrids/kyma/kyma.yaml#L422) are defined. They are used on the **dashboard_tab** and need to be defined for a job to show up. In the above example the **kind** integration test has the test group defined as `kyma-kind-integration`. You'll find the same definition of this test group in the list:
    ```yaml
    - name: kyma-kind-integration
      gcs_prefix: kyma-prow-logs/logs/kyma-kind-integration
      num_failures_to_alert: 1
    ```
    This gives the necessary information to the dashboard on where to grab the results from that are showing up on the dashboard and also gives a link to the job details if needed.

### Conclusion

To add a newly created job to the TestGrid dashboard, add the required annotations to the job. Make sure to add them to an existing dashboard or if necessary create a new one. And then coordinate changes to the kubernetes repository with one of the [codeowners](https://github.com/kubernetes/test-infra/blob/master/config/testgrids/kyma/OWNERS) on the Kyma folder.
[@tehcyx](https://github.com/tehcyx)
[@mszostok](https://github.com/mszostok)
[@piotrmiskiewicz](https://github.com/piotrmiskiewicz)