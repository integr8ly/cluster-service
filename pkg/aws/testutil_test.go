package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

const (
	fakeRDSClientTagKey                     = tagKeyClusterId
	fakeRDSClientTagVal                     = "fakeVal"
	fakeRDSClientInstanceIdentifier         = "testIdentifier"
	fakeRDSClientInstanceARN                = "arn:fake:testIdentifier"
	fakeRDSClientInstanceDeletionProtection = true
	fakeElasticacheClientRegion             = "eu-west-1"
	fakeElasticacheClientReplicationGroupId = "testRepGroupID"
	fakeElasticacheClientDescription        = "TestDescription"
	fakeElasticacheClientEngine             = "redis"
	fakeElasticacheClientTagKey             = "integreatly.org/clusterID"
	fakeElasticacheClientTagValue           = "test"
	fakeElasticacheClientCacheNodeType      = "cache.t2.micro"
	fakeElasticacheClientStatusAvailable    = "available"

	fakeActionEngineName = "Fake Action Engine"
)

func fakeReportItemDeleting() *clusterservice.ReportItem {
	return &clusterservice.ReportItem{
		ID:           fakeRDSClientInstanceARN,
		Name:         fakeRDSClientInstanceIdentifier,
		Action:       clusterservice.ActionDelete,
		ActionStatus: clusterservice.ActionStatusInProgress,
	}
}

func fakeReportItemDryRun() *clusterservice.ReportItem {
	return &clusterservice.ReportItem{
		ID:           fakeRDSClientInstanceARN,
		Name:         fakeRDSClientInstanceIdentifier,
		Action:       clusterservice.ActionDelete,
		ActionStatus: clusterservice.ActionStatusDryRun,
	}
}

func fakeRDSClientTag() *rds.Tag {
	return &rds.Tag{
		Key:   awssdk.String(fakeRDSClientTagKey),
		Value: awssdk.String(fakeRDSClientTagVal),
	}
}

func fakeRDSClientDBInstance() *rds.DBInstance {
	return &rds.DBInstance{
		DBInstanceIdentifier: awssdk.String(fakeRDSClientInstanceIdentifier),
		DBInstanceArn:        awssdk.String(fakeRDSClientInstanceARN),
		DeletionProtection:   awssdk.Bool(fakeRDSClientInstanceDeletionProtection),
	}
}

func fakeRDSClient(modifyFn func(c *rdsClientMock) error) (*rdsClientMock, error) {
	if modifyFn == nil {
		return nil, fmt.Errorf("modifyFn must be defined")
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
	}
	if err := modifyFn(client); err != nil {
		return nil, fmt.Errorf("error occurred in modify function: %w", err)
	}
	return client, nil
}

//ELASTICACHE
func fakeElasticacheReplicationGroup() *elasticache.ReplicationGroup {
	return &elasticache.ReplicationGroup{
		CacheNodeType:      awssdk.String(fakeElasticacheClientCacheNodeType),
		Description:        awssdk.String(fakeElasticacheClientDescription),
		ReplicationGroupId: awssdk.String(fakeElasticacheClientReplicationGroupId),
		Status:             awssdk.String(fakeElasticacheClientStatusAvailable),
	}
}

func fakeElasticacheClient(modifyFn func(c *elasticacheClientMock) error) (*elasticacheClientMock, error) {
	if modifyFn == nil {
		return nil, fmt.Errorf("modifyFn must be defined")
	}
	client := &elasticacheClientMock{
		DescribeReplicationGroupsFunc: func(in1 *elasticache.DescribeReplicationGroupsInput) (output *elasticache.DescribeReplicationGroupsOutput, e error) {
			return &elasticache.DescribeReplicationGroupsOutput{ReplicationGroups: []*elasticache.ReplicationGroup{
				fakeElasticacheReplicationGroup(),
			}}, nil
		},
		//ListTagsForResourceFunc: func(in1 *rds.ListTagsForResourceInput) (output *rds.ListTagsForResourceOutput, e error) {
		//	return &rds.ListTagsForResourceOutput{
		//		TagList: []*rds.Tag{
		//			fakeRDSClientTag(),
		//		},
		//	}, nil
		//},
		//ModifyDBInstanceFunc: func(in1 *rds.ModifyDBInstanceInput) (output *rds.ModifyDBInstanceOutput, e error) {
		//	return &rds.ModifyDBInstanceOutput{
		//		DBInstance: fakeRDSClientDBInstance(),
		//	}, nil
		//},
		DeleteReplicationGroupFunc: func(in1 *elasticache.DeleteReplicationGroupInput) (output *elasticache.DeleteReplicationGroupOutput, e error) {
			return &elasticache.DeleteReplicationGroupOutput{
				ReplicationGroup: fakeElasticacheReplicationGroup(),
			}, nil
		},
	}
	if err := modifyFn(client); err != nil {
		return nil, fmt.Errorf("error occurred in modify function: %w", err)
	}
	return client, nil
}

func fakeLogger(modifyFn func(l *logrus.Entry) error) (*logrus.Entry, error) {
	if modifyFn == nil {
		return nil, fmt.Errorf("modifyFn must be defined")
	}
	logger := logrus.NewEntry(logrus.StandardLogger())
	if err := modifyFn(logger); err != nil {
		return nil, fmt.Errorf("error occurred in modify function: %w", err)
	}
	return logger, nil
}

func fakeActionEngine(modifyFn func(e *ActionEngineMock) error) (*ActionEngineMock, error) {
	if modifyFn == nil {
		return nil, fmt.Errorf("modifyFn must be defined")
	}
	engine := &ActionEngineMock{
		DeleteResourcesForClusterFunc: func(clusterId string, tags map[string]string, dryRun bool) (items []*clusterservice.ReportItem, e error) {
			return []*clusterservice.ReportItem{
				fakeReportItemDeleting(),
			}, nil
		},
		GetNameFunc: func() string {
			return fakeActionEngineName
		},
	}
	if err := modifyFn(engine); err != nil {
		return nil, fmt.Errorf("error occured in modify function: %w", err)
	}
	return engine, nil
}
