#!/usr/bin/env bash

utils::checkRequiredVars() {
    local discoverUnsetVar=false
    for var in $@; do
      if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
      fi
    done
    if [ "${discoverUnsetVar}" = true ] ; then
      exit 1
    fi
}
