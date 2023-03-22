.DEFAULT_GOAL := jobs

jobs-definitions:
	go install github.com/kyma-project/test-infra/development/tools/cmd/rendertemplates
	go run github.com/kyma-project/test-infra/development/tools/cmd/rendertemplates --config templates/config.yaml --templates templates/templates --data templates/data

jobs: jobs-definitions ;
