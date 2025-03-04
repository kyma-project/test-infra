#!/usr/bin/env bash
#
# This is a helper script for validating bash scripts inside the test-infra repository.
# It uses shellcheck as a validator.
set -e
set -o pipefail

export LC_ALL=C.UTF-8

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"
# Scripts were checked with shellcheck 0.4.4, but the newer versions adds additional checks that blocks development, so we had to disable them for now
# unknown version
export SHELLCHECK_OPTS="-e SC2034 -e SC2181 -e SC2155"
# 0.4.5
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC2185 -e SC2184 -e SC2183 -e SC2182 -e SC2181 -e SC1106"
# 0.4.6
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC2204 -e SC2205 -e SC2200 -e SC2201 -e SC2198 -e SC2199 -e SC2196 -e SC2197 -e SC2195 -e SC2194 -e SC2193 -e SC2188 -e SC2189 -e SC2186 -e SC1109 -e SC1108 -e SC2191"
# 0.4.7
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC2221 -e SC2222 -e SC2220 -e SC2218 -e SC2216 -e SC2217 -e SC2215 -e SC2214 -e SC2213 -e SC2212 -e SC2211 -e SC2210 -e SC2206 -e SC2207 -e SC1117 -e SC1113 -e SC1114 -e SC1115"
# 0.5.0
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC2233 -e SC2234 -e SC2235 -e SC2232 -e SC2231 -e SC2229 -e SC2227 -e SC2224 -e SC2225 -e SC2226 -e SC2223 -e SC1131 -e SC1128 -e SC1127"
# 0.6.0
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC2152 -e SC2151 -e SC2236 -e SC2237 -e SC2238 -e SC2239 -e SC2240 -e SC1133"
# 0.7.0
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC2154 -e SC2252 -e SC2251 -e SC2250 -e SC2249 -e SC2248 -e SC2247 -e SC2246 -e SC2245 -e SC2243 -e SC2244 -e SC1135"
# 0.7.1
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC1136 -e SC2254 -e SC2255 -e SC2256 -e SC2257 -e SC2258"
# 0.7.2
# export SHELLCHECK_OPTS="${SHELLCHECK_OPTS} -e SC1090 -e SC2154 -e SC2270 -e SC2285"


find "./prow" -type f -name "*.sh" -exec "shellcheck" -x {} +

log::info "No issues detected!"

log::success "Validate scripts's all done"
# (2025-03-04)