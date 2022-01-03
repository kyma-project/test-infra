#! /bin/bash

env_file=/etc/profile

temp_cmd="export ISTIOCTL_PATH=/usr/local/bin/istioctl"

# command to write $temp_cmd to file if $temp_cmd doesnt exist w/in it
perm_cmd="grep -q -F '$temp_cmd' $env_file || echo '$temp_cmd' >> $env_file"

# set the env var permanently for any SUBSEQUENT shell logins
eval $perm_cmd