.DEFAULT_GOAL := jobs

jobs-definitions:
	go get github.com/kyma-project/test-infra/development/tools/cmd/rendertemplates 
	go run github.com/kyma-project/test-infra/development/tools/cmd/rendertemplates --config templates/config.yaml
jobs-tests:
	$(MAKE) -C development/tools $@

jobs: jobs-definitions jobs-tests ;
