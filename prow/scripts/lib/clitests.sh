#!/usr/bin/env bash

LIBDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"
# shellcheck source=prow/scripts/lib/log.sh
source "${LIBDIR}/log.sh"
# shellcheck source=prow/scripts/lib/utils.sh
source "${LIBDIR}/utils.sh"

#
# Run a command remote on GCP host.
# Arguments:
# required:
# c - command to execute
# z - GCP zone, defaults to $ZONE
# h - GCP host, defaults to $HOST
# optional:
# a - assert
# j - jq filter

#
clitests::assertRemoteCommand() {

    local OPTIND
    local cmd
    local assert
    local jqFilter
    local zone
    local host

    while getopts ":c:a:j:z:h:" opt; do
        case $opt in
            c)
                cmd="$OPTARG" ;;
            a)
                assert="$OPTARG" ;;
            j)
                jqFilter="$OPTARG" ;;
            z)
                zone="$OPTARG" ;;
            h)
                host="$OPTARG" ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
            :)
                echo "Option -$OPTARG argument not provided" >&2 ;;
        esac
    done

    utils::check_empty_arg "$cmd" "Command was not provided. Exiting..."
    utils::check_empty_arg "$zone" "GCP zone was not provided. Exiting..."
    utils::check_empty_arg "$host" "GCP host was not provided. Exiting..."

    # config values
    local interval=15
    local retries=10

    local output
    for loopCount in $(seq 1 $retries); do
        output=$(gcloud compute ssh --quiet --zone="${zone}" "${host}" -- "$cmd")
        cmdExitCode=$?

        # check return code and apply assertion (if defined)
        if [ $cmdExitCode -eq 0 ]; then
            if [ -n "$assert" ]; then
                # apply JQ filter (if defined)
                if [ -n "$jqFilter" ]; then
                    local jqOutput
                    jqOutput=$(echo "$output" | jq -r "$jqFilter")
                    jqExitCode=$?

                    # show JQ failures
                    if [ $jqExitCode -eq 0 ]; then
                        output="$jqOutput"
                    else
                        echo "JQFilter '${jqFilter}' failed with exit code '${jqExitCode}' (JSON input: ${output})"
                    fi
                fi
                # assert output
                if [ "$output" = "$assert" ]; then
                    break
                else
                    echo "Assertion failed: expected '$assert' but got '$output'"
                fi
            else
                # no assertion required
                break
            fi
        fi

        # abort if max-retires are reached
        if [ "$loopCount" -ge $retries ]; then
            echo "Failed after $loopCount attempts."
            exit 1
        fi

        echo "Retrying in $interval seconds.."
        sleep $interval
    done;
}
