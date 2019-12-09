.DEFAULT_GOAL := jobs

rendertemplates:
	$(MAKE) -C ../development/tools ARGS=--config=../../templates/config.yaml $@-run

jobs-tests:
	$(MAKE) -C ../development/tools $@

jobs: rendertemplates jobs-tests ;



