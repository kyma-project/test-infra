.PHONY: resolve validate jobs-tests

resolve:
	go mod tidy

validate: resolve
	./before-commit.sh

jobs-tests: resolve
	go test -v ./jobs/...
