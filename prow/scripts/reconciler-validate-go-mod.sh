#!/bin/bash
log::banner "Validate reconciler's go.mod file"

log::info "Preconfigure dependencies"
curl https://bootstrap.pypa.io/pip/2.7/get-pip.py --output get-pip.py
python2 get-pip.py
pip install semver==2.10

log::info "Execute validation script"
make validate-go-mod

# Execute validation script
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