#!/usr/bin/env bash

# cli-alpha::deploy starts Kyma installation using the alpha deploy command
#
function cli-alpha::deploy {
	kyma alpha deploy --ci
}