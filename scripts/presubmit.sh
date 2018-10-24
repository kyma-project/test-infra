#!/usr/bin/env bash

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPT_DIR}/library.sh"

function main() {
    init || exit 1

    local all_parameters=$@
    [[ -z $1 ]] && all_parameters="--all-tests"

    for parameter in ${all_parameters}; do
        case ${parameter} in
          --all-tests)
            RUN_RESOLVE_TESTS=1
            RUN_BUILD_TESTS=1
            RUN_VALIDATE_TESTS=1
            RUN_UNIT_TESTS=1
            RUN_INTEGRATION_TESTS=1
            RUN_BUILD_IMAGE_TESTS=1
            RUN_PUSH_IMAGE_TESTS=1
            shift
            ;;
          --resolve-tests)
            RUN_RESOLVE_TESTS=1
            shift
            ;;
          --build-tests)
            RUN_BUILD_TESTS=1
            shift
            ;;
          --validate-tests)
            RUN_VALIDATE_TESTS=1
            shift
            ;;
          --unit-tests)
            RUN_UNIT_TESTS=1
            shift
            ;;
          --integration-tests)
            RUN_INTEGRATION_TESTS=1
            shift
            ;;
          --build-image-tests)
            RUN_BUILD_IMAGE_TESTS=1
            shift
            ;;
          --push-image-tests)
            RUN_PUSH_IMAGE_TESTS=1
            shift
            ;;
          *)
            echo "error: unknown option ${parameter}"
            exit 1
            ;;
        esac
    done

    readonly RUN_RESOLVE_TESTS
    readonly RUN_BUILD_TESTS
    readonly RUN_UNIT_TESTS
    readonly RUN_INTEGRATION_TESTS
    readonly RUN_BUILD_IMAGE_TESTS
    readonly RUN_PUSH_IMAGE_TESTS
    readonly RUN_VALIDATE_TESTS

    if [[ ! -z "${SUBDIRECTORY}" ]]; then
        cd "${SUBDIRECTORY}" || exit 1
    fi

    if (( RUN_RESOLVE_TESTS )); then
        resolve_tests || exit 1
    fi

    if (( RUN_VALIDATE_TESTS )); then
        validate_tests || exit 1
    fi

    if (( RUN_BUILD_TESTS )); then
        build_tests || exit 1
    fi

    if (( RUN_UNIT_TESTS )); then
        unit_tests || exit 1
    fi

    if (( RUN_INTEGRATION_TESTS )); then
        integration_tests || exit 1
    fi

    if (( RUN_BUILD_IMAGE_TESTS )); then
        build_image_tests || exit 1
    fi

    if (( RUN_PUSH_IMAGE_TESTS )); then
        push_image_tests || exit 1
    fi

    exit 0
}