APP_NAME = image-syncer
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
TAG = $(DOCKER_TAG)

.PHONY: validate
validate:
	./before-commit.sh ci

.PHONY: ci-pr
ci-pr: validate

.PHONY: ci-main
ci-main: validate

.PHONY: ci-release
ci-release: validate

.PHONY: clean
clean:
	rm -f $(APP_NAME)
	rm -f logs-printer
.PHONY: dry-run
dry-run:
	go run main.go --dry-run=true --images-file external-images.yaml --target-key-file="${GOOGLE_APPLICATION_CREDENTIALS}"

.PHONY: run
run:
	go run main.go --dry-run=false --images-file external-images.yaml --target-key-file="${GOOGLE_APPLICATION_CREDENTIALS}"
