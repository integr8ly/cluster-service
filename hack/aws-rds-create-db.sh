#!/bin/bash

DB_ID=${1:-cluster-service-deleteme}
REGION="us-east-1"
PASSWORD=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

echo "creating test rds instance"
aws rds create-db-instance --db-instance-identifier "$DB_ID" --db-instance-class db.t2.micro \
  --engine postgres --allocated-storage=20 --region "$REGION" --master-username postgres \
  --master-user-password "$PASSWORD" --copy-tags-to-snapshot --tags Key=integreatly.org/clusterID,Value=cluster-service