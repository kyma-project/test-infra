.DEFAULT_GOAL := jobs

jobs-definitions:
	go install github.com/kyma-project/test-infra/development/tools/cmd/rendertemplates
	go run github.com/kyma-project/test-infra/development/tools/cmd/rendertemplates --config templates/config.yaml
jobs-tests:
	$(MAKE) -C development/tools $@

jobs: jobs-definitions jobs-tests ;
