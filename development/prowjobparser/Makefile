APP_NAME = prowjobparser
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: validate
validate:
	./before-commit.sh ci

.PHONY: clean
clean:
	rm -f $(APP_NAME)

