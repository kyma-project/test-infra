.DEFAULT_GOAL := jobs

jobs-definitions:
	 go run development/tools/cmd/rendertemplates/main.go --config templates/config.yaml --templates templates/templates --data templates/data

jobs: jobs-definitions ;
