#!/bin/bash

echo "Running rendertemplates pre-commit hook..."
(
  cd "$(git rev-parse --show-toplevel)" || exit 1
    if files=$(go run development/tools/cmd/rendertemplates/main.go -config templates/config.yaml -show-output-dir=true); then
      if [ -z "$files" ]; then
        echo "No changes made. Continuing..."
      else
        modified=$(git diff --exit-code --name-only "$files")
        echo "Templates have been regenerated and automatically added to your commit."
        echo
        echo "Modified files:"
        echo "$modified"
        git add -u "$files"
      fi
    else
      echo "Failed to render templates. Commit aborted."
      exit 1
    fi
)
