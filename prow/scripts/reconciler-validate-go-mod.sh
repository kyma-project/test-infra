#!/bin/bash
log::banner "Validate reconciler's go.mod file"

set -o errexit

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# shellcheck source=prow/scripts/lib/log.sh
source "${SCRIPT_DIR}/lib/log.sh"

# Configure dependencies
log::info "Configure dependencies"
curl https://bootstrap.pypa.io/pip/2.7/get-pip.py --output get-pip.py
python2 get-pip.py
pip install semver==2.10

# Execute validation script
log::info "Execute validation script"
python2 ./scripts/validate-go-mod.py && \
	([ $$? -eq 0 ] && echo "Result: go.mod is VALID") \
	|| (echo "Result: go.mod is INVALID (see log above)"; exit 1)

# TODO: Test script exit code
# if [[ $? -eq 0 ]];then
#         log::success "Tests completed"
#     else
#         log::error "Tests failed"
#         exit 1
#     fi