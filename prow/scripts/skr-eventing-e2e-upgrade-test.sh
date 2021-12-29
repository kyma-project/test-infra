# execute eventing e2e pre upgrade tests

export SKR_SUFFIX="test"
export SKR_INSTANCE_ID=$(uuidgen)


log::info "Running skr eventing e2e upgrade with testst with SKR_SUFFIX: ${SKR_SUFFIX} and SKR_INSTANCE_ID: ${SKR_INSTANCE_ID}"

make provision-eventing-skr

# test the default Eventing backend which comes with Kyma
eventing::pre_upgrade_test_fast_integration_eventing

# upgrade the kyma to the current PR/commit state
KYMA_SOURCE="PR-${PULL_NUMBER}"
export KYMA_SOURCE

make upgrade-eventing-skr

# test the eventing fi tests after the upgrade
eventing::post_upgrade_test_fast_integration_eventing

#!!! Must be at the end of the script !!!
ERROR_LOGGING_GUARD="false"