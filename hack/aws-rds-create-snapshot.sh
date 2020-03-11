#!/bin/bash

DB_ID=${1:-cluster-service-deleteme}
SNAPSHOT_ID=${1:-cluster-service-deleteme}
REGION="us-east-1"
SNAPSHOT_SUFFIX=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 4 | head -n 1)

echo "creating test rds instance"
aws rds create-db-snapshot --db-instance-identifier "$DB_ID" --db-snapshot-identifier "$SNAPSHOT_ID-$SNAPSHOT_SUFFIX" --region "$REGION"