#!/bin/bash

REGION="us-east-1"
RG_ID=${1:-cluster-service-deleteme}

echo "creating replication group in region $REGION"
aws elasticache create-replication-group --region "$REGION" \
  --replication-group-id "$RG_ID" --tags "Key=integreatly.org/clusterID,Value=cluster-service" \
  --replication-group-description "$RG_ID" --engine redis \
  --cache-node-type cache.t2.micro