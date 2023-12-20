.DEFAULT_GOAL := jobs

jobs-definitions:
	 go run cmd/tools/rendertemplates/main.go --config templates/config.yaml --templates templates/templates --data templates/data

jobs: jobs-definitions ;
