package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/elasticache"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
)

const (
	//generic variables
	fakeARN                = "arn:fake:testIdentifier"
	fakeArnWithSlash       = "arn:fake:resourceType/testIdentifier"
	fakeResourceIdentifier = "testIdentifier"
	fakeClusterId          = "clusterId"

	//ec2-specific
	fakeEc2ClientInstanceArn = fakeArnWithSlash

	//rds-specific
	fakeRDSClientTagKey                     = tagKeyClusterId
	fakeRDSClientTagVal                     = "fakeVal"
	fakeRDSClientInstanceIdentifier         = fakeResourceIdentifier
	fakeRDSClientInstanceARN                = fakeARN
	fakeRDSClientInstanceDeletionProtection = true
	fakeRDSClientDBSubnetGroupARN           = fakeARN

	//ELasticache-specific
	fakeElasticacheClientName               = "elasticache Replication group"
	fakeElasticacheClientReplicationGroupId = "testRepGroupID"
	fakeElasticacheClientDescription        = "TestDescription"
	fakeElasticacheClientEngine             = "redis"
	fakeElasticacheClientCacheNodeType      = "cache.t2.micro"
	fakeElasticacheClientStatusAvailable    = "available"
	fakeClusterID                           = "testClusterID"
	fakeCacheClusterStatus                  = "available"
	fakeElasticacheSnapshotName             = "elasticache snapshot"
	fakeElasticacheSnapshotStatus           = "available"
	fakeElasticacheSubnetGroupName          = "elasticache subnet group"
	fakeElasticacheSubnetGroupNameValue     = "testCacheSubnetGroup"
	fakeElasticacheSubnetGroupID            = "subnetgroup:testCacheSubnetGroup"

	//resource tagging-specific
	fakeResourceTagMappingARN = fakeARN

	//resource manager-specific
	fakeResourceManagerName = "Fake Action Engine"

	// db snapshot
	fakeSnapshotType    = "manual"
	fakeSnapshotStatus  = "available"
	fakeRDSSnapshotName = "rds-snapshot"
)

func mockReportItem(modifyFn func(*clusterservice.ReportItem)) *clusterservice.ReportItem {
	mock := &clusterservice.ReportItem{}
	if modifyFn != nil {
		modifyFn(mock)
	}
	return mock
}

func fakeRDSClientTag() *rds.Tag {
	return &rds.Tag{
		Key:   aws.String(fakeRDSClientTagKey),
		Value: aws.String(fakeRDSClientTagVal),
	}
}

func fakeRDSClientDBInstance() *rds.DBInstance {
	return &rds.DBInstance{
		DBInstanceIdentifier: aws.String(fakeRDSClientInstanceIdentifier),
		DBInstanceArn:        aws.String(fakeRDSClientInstanceARN),
		DeletionProtection:   aws.Bool(fakeRDSClientInstanceDeletionProtection),
	}
}

func fakeResourceTagMappingTag() *resourcegroupstaggingapi.Tag {
	return &resourcegroupstaggingapi.Tag{
		Key:   aws.String(tagKeyClusterId),
		Value: aws.String(fakeClusterId),
	}
}

func fakeResourceTagMapping(modifyFn func(*resourcegroupstaggingapi.ResourceTagMapping)) *resourcegroupstaggingapi.ResourceTagMapping {
	mock := &resourcegroupstaggingapi.ResourceTagMapping{
		ComplianceDetails: nil,
		ResourceARN:       aws.String(fakeResourceTagMappingARN),
		Tags: []*resourcegroupstaggingapi.Tag{
			fakeResourceTagMappingTag(),
		},
	}
	if modifyFn != nil {
		modifyFn(mock)
	}
	return mock
}

func fakeRDSClientDBSnapshots() []*rds.DBSnapshot {
	return []*rds.DBSnapshot{
		fakeRDSSnapshot(),
	}
}

func fakeRDSSnapshot() *rds.DBSnapshot {
	return &rds.DBSnapshot{
		Engine:               aws.String(fakeElasticacheClientEngine),
		DBInstanceIdentifier: aws.String(fakeResourceIdentifier),
		DBSnapshotIdentifier: aws.String(fakeRDSSnapshotName),
		Status:               aws.String(fakeSnapshotStatus),
		SnapshotType:         aws.String(fakeSnapshotType),
	}
}

func fakeRDSSubnetGroup() *rds.DBSubnetGroup {
	return &rds.DBSubnetGroup{
		DBSubnetGroupArn:  aws.String(fakeRDSClientDBSubnetGroupARN),
		DBSubnetGroupName: aws.String(fakeResourceIdentifier),
	}
}

func fakeRDSClient(modifyFn func(c *rdsClientMock) error) (*rdsClientMock, error) {
	if modifyFn == nil {
		return nil, errorMustBeDefined("modifyFn")
	}
	client := &rdsClientMock{
		DescribeDBInstancesFunc: func(in1 *rds.DescribeDBInstancesInput) (output *rds.DescribeDBInstancesOutput, e error) {
			return &rds.DescribeDBInstancesOutput{
				DBInstances: []*rds.DBInstance{
					fakeRDSClientDBInstance(),
				},
			}, nil
		},
		ListTagsForResourceFunc: func(in1 *rds.ListTagsForResourceInput) (output *rds.ListTagsForResourceOutput, e error) {
			return &rds.ListTagsForResourceOutput{
				TagList: []*rds.Tag{
					fakeRDSClientTag(),
				},
			}, nil
		},
		ModifyDBInstanceFunc: func(in1 *rds.ModifyDBInstanceInput) (output *rds.ModifyDBInstanceOutput, e error) {
			return &rds.ModifyDBInstanceOutput{
				DBInstance: fakeRDSClientDBInstance(),
			}, nil
		},
		DeleteDBInstanceFunc: func(in1 *rds.DeleteDBInstanceInput) (output *rds.DeleteDBInstanceOutput, e error) {
			return &rds.DeleteDBInstanceOutput{
				DBInstance: fakeRDSClientDBInstance(),
			}, nil
		},
		DescribeDBSnapshotsFunc: func(in1 *rds.DescribeDBSnapshotsInput) (*rds.DescribeDBSnapshotsOutput, error) {
			return &rds.DescribeDBSnapshotsOutput{
				DBSnapshots: fakeRDSClientDBSnapshots(),
			}, nil
		},
		DeleteDBSnapshotFunc: func(in1 *rds.DeleteDBSnapshotInput) (*rds.DeleteDBSnapshotOutput, error) {
			return &rds.DeleteDBSnapshotOutput{}, nil
		},
		DescribeDBSubnetGroupsFunc: func(in1 *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error) {
			return &rds.DescribeDBSubnetGroupsOutput{
				DBSubnetGroups: []*rds.DBSubnetGroup{
					fakeRDSSubnetGroup(),
				},
			}, nil
		},
	}
	if err := modifyFn(client); err != nil {
		return nil, errorModifyFailed(err)
	}
	return client, nil
}

type mockEc2Client struct {
	ec2iface.EC2API
	deleteVpcFn                  func(*ec2.DeleteVpcInput) (*ec2.DeleteVpcOutput, error)
	deleteVpcPeeringConnectionFn func(*ec2.DeleteVpcPeeringConnectionInput) (*ec2.DeleteVpcPeeringConnectionOutput, error)
	deleteSubnetFn               func(*ec2.DeleteSubnetInput) (*ec2.DeleteSubnetOutput, error)
	deleteSecurityGroupFn        func(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error)
	deleteRouteTableFn           func(*ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error)
}

func buildMockEc2Client(modifyFn func(*mockEc2Client)) *mockEc2Client {
	mock := &mockEc2Client{}
	if modifyFn != nil {
		modifyFn(mock)
	}
	return mock
}

func (m *mockEc2Client) DeleteVpc(input *ec2.DeleteVpcInput) (*ec2.DeleteVpcOutput, error) {
	return m.deleteVpcFn(input)
}

func (m *mockEc2Client) DeleteVpcPeeringConnection(input *ec2.DeleteVpcPeeringConnectionInput) (*ec2.DeleteVpcPeeringConnectionOutput, error) {
	return m.deleteVpcPeeringConnectionFn(input)
}

func (m *mockEc2Client) DeleteSubnet(input *ec2.DeleteSubnetInput) (*ec2.DeleteSubnetOutput, error) {
	return m.deleteSubnetFn(input)
}

func (m *mockEc2Client) DeleteSecurityGroup(input *ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error) {
	return m.deleteSecurityGroupFn(input)
}

func (m *mockEc2Client) DeleteRouteTable(input *ec2.DeleteRouteTableInput) (*ec2.DeleteRouteTableOutput, error) {
	return m.deleteRouteTableFn(input)
}

func fakeS3Client(modifyFn func(c *s3ClientMock) error) (*s3ClientMock, error) {
	if modifyFn == nil {
		return nil, errorMustBeDefined("modifyFn")
	}
	client := &s3ClientMock{
		DeleteBucketFunc: func(in1 *s3.DeleteBucketInput) (output *s3.DeleteBucketOutput, e error) {
			return &s3.DeleteBucketOutput{}, nil
		},
	}
	if err := modifyFn(client); err != nil {
		return nil, errorModifyFailed(err)
	}
	return client, nil
}

func fakeTaggingClient(modifyFn func(c *taggingClientMock) error) (*taggingClientMock, error) {
	if modifyFn == nil {
		return nil, errorMustBeDefined("modifyFn")
	}
	client := &taggingClientMock{
		GetResourcesFunc: func(in1 *resourcegroupstaggingapi.GetResourcesInput) (output *resourcegroupstaggingapi.GetResourcesOutput, e error) {
			return &resourcegroupstaggingapi.GetResourcesOutput{
				ResourceTagMappingList: []*resourcegroupstaggingapi.ResourceTagMapping{
					fakeResourceTagMapping(func(mapping *resourcegroupstaggingapi.ResourceTagMapping) {}),
				},
			}, nil
		},
	}
	if err := modifyFn(client); err != nil {
		return nil, fmt.Errorf("error occurred in modify function: %w", err)
	}
	return client, nil
}

func fakeS3BatchClient(modifyFn func(c *s3BatchDeleteClientMock) error) (*s3BatchDeleteClientMock, error) {
	if modifyFn == nil {
		return nil, errorMustBeDefined("modifyFn")
	}
	client := &s3BatchDeleteClientMock{
		DeleteFunc: func(in1 context.Context, in2 s3manager.BatchDeleteIterator) error {
			return nil
		},
	}
	if err := modifyFn(client); err != nil {
		return nil, fmt.Errorf("error occurred in modify function: %w", err)
	}
	return client, nil
}

//ELASTICACHE
func fakeElasticacheSnapshot() *elasticache.Snapshot {
	return &elasticache.Snapshot{
		CacheClusterId: aws.String(fakeClusterID),
		CacheNodeType:  aws.String(fakeElasticacheClientCacheNodeType),
		Engine:         aws.String(fakeElasticacheClientEngine),
		SnapshotName:   aws.String(fakeElasticacheSnapshotName),
		SnapshotStatus: aws.String(fakeElasticacheSnapshotStatus),
	}
}

func fakeElasticacheReplicationGroup() *elasticache.ReplicationGroup {
	return &elasticache.ReplicationGroup{
		CacheNodeType:      aws.String(fakeElasticacheClientCacheNodeType),
		Description:        aws.String(fakeElasticacheClientDescription),
		ReplicationGroupId: aws.String(fakeElasticacheClientReplicationGroupId),
		Status:             aws.String(fakeElasticacheClientStatusAvailable),
	}
}
func fakeElasticacheCacheCluster() *elasticache.CacheCluster {
	return &elasticache.CacheCluster{
		CacheClusterId:       aws.String(fakeClusterID),
		CacheClusterStatus:   aws.String(fakeCacheClusterStatus),
		CacheNodeType:        aws.String(fakeElasticacheClientCacheNodeType),
		Engine:               aws.String(fakeElasticacheClientEngine),
		ReplicationGroupId:   aws.String(fakeElasticacheClientReplicationGroupId),
		CacheSubnetGroupName: aws.String(fakeElasticacheSubnetGroupNameValue),
	}
}

func fakeElasticacheClient(modifyFn func(c *elasticacheClientMock) error) (*elasticacheClientMock, error) {
	if modifyFn == nil {
		return nil, fmt.Errorf("modifyFn must be defined")
	}
	client := &elasticacheClientMock{
		DescribeReplicationGroupsFunc: func(in1 *elasticache.DescribeReplicationGroupsInput) (output *elasticache.DescribeReplicationGroupsOutput, e error) {
			return &elasticache.DescribeReplicationGroupsOutput{
				ReplicationGroups: []*elasticache.ReplicationGroup{
					fakeElasticacheReplicationGroup(),
				}}, nil
		},
		DescribeSnapshotsFunc: func(in1 *elasticache.DescribeSnapshotsInput) (output *elasticache.DescribeSnapshotsOutput, e error) {
			return &elasticache.DescribeSnapshotsOutput{
				Snapshots: []*elasticache.Snapshot{
					fakeElasticacheSnapshot(),
				}}, nil
		},
		DescribeCacheClustersFunc: func(in1 *elasticache.DescribeCacheClustersInput) (output *elasticache.DescribeCacheClustersOutput, e error) {
			return &elasticache.DescribeCacheClustersOutput{
				CacheClusters: []*elasticache.CacheCluster{
					fakeElasticacheCacheCluster(),
				}}, nil
		},
		DeleteReplicationGroupFunc: func(in1 *elasticache.DeleteReplicationGroupInput) (output *elasticache.DeleteReplicationGroupOutput, e error) {
			return &elasticache.DeleteReplicationGroupOutput{
				ReplicationGroup: fakeElasticacheReplicationGroup(),
			}, nil
		},
		DeleteSnapshotFunc: func(in1 *elasticache.DeleteSnapshotInput) (output *elasticache.DeleteSnapshotOutput, e error) {
			return &elasticache.DeleteSnapshotOutput{
				Snapshot: fakeElasticacheSnapshot(),
			}, nil
		},
		DeleteCacheSubnetGroupFunc: func(in1 *elasticache.DeleteCacheSubnetGroupInput) (out *elasticache.DeleteCacheSubnetGroupOutput, err error) {
			return &elasticache.DeleteCacheSubnetGroupOutput{}, nil
		},
	}
	if err := modifyFn(client); err != nil {
		return nil, fmt.Errorf("error occurred in modify function: %w", err)
	}
	return client, nil
}

func fakeLogger(modifyFn func(l *logrus.Entry) error) (*logrus.Entry, error) {
	if modifyFn == nil {
		return nil, errorMustBeDefined("modifyFn")
	}
	logger := logrus.NewEntry(logrus.StandardLogger())
	if err := modifyFn(logger); err != nil {
		return nil, errorModifyFailed(err)
	}
	return logger, nil
}

func fakeClusterManager(modifyFn func(e *ClusterResourceManagerMock) error) (*ClusterResourceManagerMock, error) {
	if modifyFn == nil {
		return nil, errorMustBeDefined("modifyFn")
	}
	clusterManager := &ClusterResourceManagerMock{
		DeleteResourcesForClusterFunc: func(clusterId string, tags map[string]string, dryRun bool) (items []*clusterservice.ReportItem, e error) {
			return []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress
				}),
			}, nil
		},
		GetNameFunc: func() string {
			return fakeResourceManagerName
		},
	}
	if err := modifyFn(clusterManager); err != nil {
		return nil, errorModifyFailed(err)
	}
	return clusterManager, nil
}

func errorMustBeDefined(varName string) error {
	return fmt.Errorf("%s must be defined", varName)
}

func errorModifyFailed(err error) error {
	return fmt.Errorf("error occurred while modifying resource: %w", err)
}
