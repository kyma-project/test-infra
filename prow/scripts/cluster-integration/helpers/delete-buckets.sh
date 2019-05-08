#!/usr/bin/env bash

#Description: Deletes buckets created by Asset Store
#
#Expected vars:
# - CLUSTER_BUCKETS: list of ClusterBuckets to remove
# - BUCKETS: list of Buckets to remove
# - UPLOADER_PRIVATE_BUCKET: Asset Uploader Private Bucket to remove
# - UPLOADER_PUBLIC_BUCKET: Asset Uploader Public Bucket to remove
#
#Permissions: In order to run this script you need to use a service account with "Storage Admin" role

set +e

function deleteBucket {
    echo "Delete Bucket: ${1}"
    gsutil rm -r "gs://${1}"
}

if [ -n "${UPLOADER_PRIVATE_BUCKET}" ]; then
    deleteBucket "${UPLOADER_PRIVATE_BUCKET}"
fi

if [ -n "${UPLOADER_PUBLIC_BUCKET}" ]; then
    deleteBucket "${UPLOADER_PUBLIC_BUCKET}"
fi

for clusterBucket in ${CLUSTER_BUCKETS}
do
    deleteBucket "${clusterBucket}"
done

for bucket in ${BUCKETS}
do
    deleteBucket "${bucket}"
done