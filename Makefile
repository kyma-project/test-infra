.DEFAULT_GOAL := jobs

jobs-definitions:
	go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml
jobs-tests:
	$(MAKE) -C development/tools $@

jobs: jobs-definitions jobs-tests ;



