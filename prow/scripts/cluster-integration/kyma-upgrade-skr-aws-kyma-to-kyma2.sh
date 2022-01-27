

# Fetch latest Kyma2 release
kyma::get_last_release_version -t "${BOT_GITHUB_TOKEN}"
export KYMA_SOURCE="${kyma_get_last_release_version_return_version:?}"
log::info "### Reading release version from RELEASE_VERSION file, got: ${KYMA_SOURCE}"

log::info "### Run make ci-skr-kyma-to-kyma2-upgrade"
make -C /home/prow/go/src/github.com/kyma-project/kyma/tests/fast-integration ci-skr-kyma-to-kyma2-upgrade