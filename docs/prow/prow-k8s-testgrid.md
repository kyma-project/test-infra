# TestGrid

[TestGrid](https://testgrid.k8s.io) is an interactive dashboard for viewing tests results in a grid. It parses JUnit reports for generating grid view from the tests.
There is one dashboard group called `kyma` which groups dashboards from both of the organizations: *kyma-project* and *kyma-incubator*. The dashboards start with the prefix name which refers to one of the organizations: *kyma* and *kyma-incubator*.
TestGrid configuration is stored inside [kubernetes/test-infra](https://github.com/kubernetes/test-infra/tree/master/config/testgrids/kyma) repository in the `/config/testgrids/kyma/` directory.

TestGrid is being automatically updated by Prow using the tool called [transfigure.sh](https://github.com/kubernetes/test-infra/tree/master/testgrid/cmd/transfigure).

## Adding new dashboard

The dashboards' configuration is stored in the [testgrid-default.yaml](https://github.com/kyma-project/test-infra/tree/master/prow/testgrid-default.yaml) file.
This file is being automatically generated from a [template testgrid-default.yaml](https://github.com/kyma-project/test-infra/blob/master/templates/templates/testgrid-default.yaml) file.
```yaml
dashboards:
  # kyma
  - name: kyma_integration
  - name: kyma_control-plane

  # kyma-incubator
  - name: kyma-incubator_compass

dashboard_groups:
  - name: kyma
    dashboard_names:
      - kyma_integration
      - kyma_control-plane
      - kyma-incubator_compass
```
To add a new dashboard it's needed to do several steps:

1. Add a new dashboard definition in the `dashboards` field according to the example file above inside a template file. **Dashboards need to have dashboard name as prefix**.
2. Add the previously added dashboard to the corresponding `dashboard_name` inside the `dashboard_groups` field.
3. Generate a new config file using [rendertemplates](https://github.com/kyma-project/test-infra/tree/master/development/tools/cmd/rendertemplates) tool and check if the config file generated correctly.

## Adding ProwJob to the TestGrid

After adding a desired dashboard in the `testgrid-default.yaml` file you need to add ProwJobs to the dashboard. It can be done by defining new `annotations` field to a ProwJob definition.
Add the below line to the ProwJob and edit it to your needs.

The only required field is `testgrid-dashboards`. It needs to have a name that is defined in the `testgrid-default.yaml` file. It is also good to define `description` field to add short description about the job.
The rest of the fields is optional and can be omitted.
```yaml
annotations:
  testgrid-dashboards: dashboard-name      # Required. a dashboard already defined in a config.yaml.
  testgrid-tab-name: some-short-name       # optionally, a shorter name for the tab. If omitted, just uses the job name.
  testgrid-alert-email: me@me.com          # optionally, an alert email that will be applied to the tab created in the
                                           # first dashboard specified in testgrid-dashboards.
  description: Words about your job.       # optionally, a description of your job. If omitted, just uses the job name.

  testgrid-num-columns-recent: "10"        # optionally, the number of runs a row can be omitted from before it is
                                           # considered stale. Currently defaults to 10.
  testgrid-num-failures-to-alert: "3"      # optionally, the number of continuous failures before sending an email.
                                           # Currently defaults to 3.
  testgrid-alert-stale-results-hours: "12" # optionally, send an email if this many hours pass with no results at all.
```

If the job does not have to be on a TestGrid use the annotation below to disable the generation of the TestGrid test group:
```yaml
annotations:
  testgrid-create-test-group: "false"
```

This configuration generally applies for postsumbits and periodics. The presubmits jobs can be added to the TestGrid, however if there is no `annnotations` field defined the job will be omitted in config file generation.
