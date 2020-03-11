#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
BUCKET_ID=${1:-cluster-service-deleteme}
REGION="us-east-1"

echo "creating test bucket"
aws s3api create-bucket --bucket="$BUCKET_ID" --region="$REGION"
echo "tagging bucket with cluster id tag"
aws s3api put-bucket-tagging --bucket="$BUCKET_ID" --region="$REGION" --tagging 'TagSet=[{Key=integreatly.org/clusterID,Value=cluster-service}]'
echo "syncing test files from $DIR to bucket"
aws s3 sync "$DIR/files" "s3://$BUCKET_ID" --region="$REGION"