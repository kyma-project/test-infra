#!/usr/bin/env sh
set -e
## Add curl and jq
apk --no-cache add curl jq

## Execute HTTP PUT to mothership-reconciler to http://mothership-reconciler-host:port/v1/clusters
statusURL=$(curl --request POST -sL \
     --url 'http://reconciler-mothership-reconciler.reconciler/v1/clusters'\
     --data @/tmp/body.json | jq -r .statusUrl)

## Wait until Kyma is installed
timeout=20 # in secs
delay=2 # in secs
iterationsLeft=$(( timeout/delay ))
while : ; do
  status=$(curl http://$statusURL | jq -r .status)
  if [ "${status}" = "Installed" ]; then
    echo "kyma is installed"
    break
  fi

  if [ "$timeout" -ne 0 ] && [ "$iterationsLeft" -le 0 ]; then
    echo "timeout reached on kyma installation error. Exiting"
    exit 1
  fi

  sleep $delay
  echo "waiting to get Kyma installed...."
  iterationsLeft=$(( iterationsLeft-1 ))
done