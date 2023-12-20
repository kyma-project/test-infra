#!/bin/bash

echo "Running rendertemplates pre-commit hook..."
(
  cd "$(git rev-parse --show-toplevel)" || exit 1
    if files=$(go run cmd/tools/rendertemplates/main.go -config templates/config.yaml -show-output-dir=true -data templates/data -templates templates/templates); then
      # shellcheck disable=SC2086
      modified=$(git diff --exit-code --name-only $files) && echo "No changes made. Continuing..." || echo "Templates have been regenerated and automatically added to your commit.

  Modified files:
  $modified"
    else
      echo "Failed to render templates. Commit aborted."
      exit 1
    fi
    # shellcheck disable=SC2086
    git add -u $files
)
