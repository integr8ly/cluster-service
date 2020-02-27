package aws

import (
	"fmt"

	"github.com/integr8ly/cluster-service/pkg/clusterservice"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sirupsen/logrus"

	awssdk "github.com/aws/aws-sdk-go/aws"
)

const (
	fakeRDSClientTagKey                     = tagKeyClusterId
	fakeRDSClientTagVal                     = "fakeVal"
	fakeRDSClientInstanceIdentifier         = "testIdentifier"
	fakeRDSClientInstanceARN                = "arn:fake:testIdentifier"
	fakeRDSClientInstanceDeletionProtection = true

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
