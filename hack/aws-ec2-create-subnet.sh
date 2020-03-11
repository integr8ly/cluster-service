#!/bin/bash

REGION="us-east-1"
VPC_ID=$(aws ec2 describe-vpcs --filters Name=isDefault,Values=true --region "$REGION" | jq -r '.Vpcs[0].VpcId')

echo "creating subnet in vpc $VPC_ID in region $REGION"
SUBNET_ID=$(aws ec2 create-subnet --region "$REGION" --cidr-block 172.31.128.0/20 --vpc-id "$VPC_ID" | jq -r '.Subnet.SubnetId')
aws ec2 create-tags --resources "$SUBNET_ID" --tags Key=integreatly.org/clusterID,Value=cluster-service --region us-east-1