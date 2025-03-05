
WORKFLOW_REF_REGEX=^kyma-project/test-infra/.github/workflows/image-builder-test.yml@refs/pull/[0-9]+$/m

# Allow only the test workflow to use the image-builder image from the input
if [[ ! "kyma-project/test-infra/.github/workflows/image-builder-test.yml@refs/pull/142" =~ $WORKFLOW_REF_REGEX ]] && [[ -n "" ]]; then
    echo "Only main branch is allowed to trigger this workflow."
    exit 1
fi