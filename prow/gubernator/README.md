Gubernator is a frontend for displaying Kyma test results stored in GCS.

It runs on Google App Engine, and parses JSON and junit.xml results for display.

https://kyma-project.appspot.com

# Development

- Install the Google Cloud SDK: https://cloud.google.com/sdk/
- Deploy with `make deploy` followed by `make migrate`
