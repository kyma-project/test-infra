#!/usr/bin/env bash

# host::os returns host operating system
#
# Returns:
#   Name of the operating system - linux or darwin
host::os() {
    local host_os
    case "$(uname -s)" in
        Darwin)
            host_os=darwin
            ;;
        Linux)
            host_os=linux
            ;;
        *)
            echo "Unsupported host OS. Must be Linux or Mac OS X."
            return 1
            ;;
    esac
    echo "${host_os}"
}
