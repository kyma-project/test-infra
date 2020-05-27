#!/usr/bin/env bash

# Adjust below values and execute

helm install deploy/chart/stability-checker \
  --set clusterName="TBD" \
  --set slackClientWebhookUrl="TBD" \
  --set slackClientChannelId="TBD" \
  --set slackClientToken="TBD" \
  --set testThrottle="1m" \
  --set testResultWindowTime="30m" \
  --set stats.enabled="false" \
  --set stats.failingTestRegexp="Test status: ([0-9A-Za-z_-]+) - Failed" \
  --set stats.successfulTestRegexp="Test status: ([0-9A-Za-z_-]+) - Succeeded" \
  --set pathToTestingScript="/data/input/testing.sh" \
  --namespace="kyma-system" \
  --name="stability-checker" \
  --tls
