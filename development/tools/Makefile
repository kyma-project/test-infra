resolve:
	go mod verify

validate: resolve
	./before-commit.sh

jobs-tests: resolve
	go test ./jobs/...

