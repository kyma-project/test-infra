#!/usr/bin/env bash

function log::date {
    date +"%Y/%m/%d %T %Z"
}

function log::info {
    echo "$(log::date) [INFO] ${1}"
}

function log::warn {
    echo "$(log::date) [WARN] ${1}"
}

function log::error {
    >&2 echo "$(log::date) [ERRO] ${1}"
}