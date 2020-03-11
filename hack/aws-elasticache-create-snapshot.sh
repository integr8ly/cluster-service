#!/bin/bash

REGION="us-east-1"
RG_ID=${1:-cluster-service-deleteme}

echo "creating snapshot replication $RG_ID group in region $REGION"
aws elasticache create-snapshot --region "$REGION" \
  --cache-cluster-id "$RG_ID-001" --snapshot-name "$RG_ID"