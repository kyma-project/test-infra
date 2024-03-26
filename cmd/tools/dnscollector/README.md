# IP Address and DNS Record Garbage Collector

## Overview

This command finds and removes orphaned IP Addresses and related DNS records created by GKE integration jobs in a Google Cloud (GCP) project.

When an integration job installs Kyma on a GKE cluster, it reserves the required IP addresses and adds necessary DNS records.
Usually, the job that provisions the cluster cleans up the resources once they are not needed.
It can happen, however, that the job cleanup process fails.
This causes a resource leak that generates unwanted costs.
The garbage collector finds and removes unused IP addresses and related DNS records.


There are three conditions used to find IP address for removal:
- The address name pattern that is specific for the given GKE integration job
- The status of the IP address shows it is not in use (this value is not configurable by the user)
- The address `creationTimestamp` value that is used to find addresses existing at least for a preconfigured number of hours

Addresses that meet all these conditions are subject to removal.
Before removal of a matching address, the command finds all DNS records associated with the IP address.
The command removes all associated DNS records first, then the IP Address.

## Usage

For safety reasons, the dry-run mode is the default one.
To run it, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} --dnsZone={gcloud managed zone name} \
    --regions={list of GCP regions}
```

To turn the dry-run mode off, use:
```bash
env GOOGLE_APPLICATION_CREDENTIALS={path to service account file} go run main.go \
    --project={gcloud project name} --dnsZone={gcloud managed zone name} \
    --regions={list of GCP regions} --dryRun=false
```

### Flags

See the list of available flags:

| Name                    | Required | Description                                                                                          |
| :---------------------- | :------: | :--------------------------------------------------------------------------------------------------- |
| **--project**           |   Yes    | GCP project name.
| **--regions**           |   Yes    | GCP region name or a comma-separated list of such values.
| **--dnsZone**           |   Yes    | GCP DNS Managed Zone name used to look up DNS records for removal.
| **--dryRun**            |    No    | The boolean value that controls the dry-run mode. It defaults to `true`.
| **--ageInHours**        |    No    | The integer value for the number of hours. It only matches addresses older than `now()-ageInHours`. It defaults to `2`.
| **--addressRegexpList** |    No    | The string value with a Golang regexp or a comma-separated list of such expressions. It is used to match addresses by their name. It defaults to `(remoteenvs-)?gkeint-(pr|commit)-.*,(remoteenvs-)?gke-upgrade-(pr|commit)-.*`.

### Environment Variables

See the list of available environment variables:

| Name                                  | Required | Description                                                                                          |
| :------------------------------------ | :------: | :--------------------------------------------------------------------------------------------------- |
| **GOOGLE_APPLICATION_CREDENTIALS**    |    Yes   | The path to the service account file. The service account requires at least `compute.addresses.list, compute.addresses.delete, dns.resourceRecordSets.list`, and `dns.changes.create` Google IAM permissions. |
