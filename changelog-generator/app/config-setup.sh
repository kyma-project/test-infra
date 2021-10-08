#!/usr/bin/env sh

echo "Copying configuration file..."
cp "${CONFIG_FILE}" ./package.json

removeLatestTag() {
    if [ "$(git tag -l "$LATEST_TAG")" ]; then
        echo "Temporary removing 'latest' tag..."
        git tag -d "${LATEST_TAG}"
    fi
}

if [ "$SKIP_REMOVING_LATEST" != "true" ]; then
    removeLatestTag
fi
