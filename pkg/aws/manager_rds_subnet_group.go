package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/integr8ly/cluster-service/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	loggingKeySubnetGroup = "subnet-group-name"
)

var _ ClusterResourceManager = &RDSSubnetGroupManager{}

type RDSSubnetGroupManager struct {
	rdsClient rdsClient
	logger    *logrus.Entry
}

func NewDefaultRDSSubnetGroupManager(session *session.Session, logger *logrus.Entry) *RDSSubnetGroupManager {
	fmt.Println("creating new RDS Subnet Group Manager")
	return &RDSSubnetGroupManager{
		rdsClient: rds.New(session),
		logger:    logger.WithField("engine", managerRDS),
	}
}

func (r *RDSSubnetGroupManager) GetName() string {
	return "AWS RDS Subnet Group Manager"
}

// Delete all RDS Subnet Groups for a specified cluster
func (r *RDSSubnetGroupManager) DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) ([]*clusterservice.ReportItem, error) {
	r.logger.Debug("deleting resources for cluster")
	subnetGroupsDescribeInput := &rds.DescribeDBSubnetGroupsInput{}
	subnetGroupsDescribeOutput, err := r.rdsClient.DescribeDBSubnetGroups(subnetGroupsDescribeInput)

	if err != nil {
		return nil, errors.WrapLog(err, "failed to describe database subnet groups", r.logger)
	}

	var subnetGroupsToDelete []*rds.DBSubnetGroup

	for _, subnetGroup := range subnetGroupsDescribeOutput.DBSubnetGroups {
		subnetGroupLogger := r.logger.WithField(loggingKeySubnetGroup, aws.StringValue(subnetGroup.DBSubnetGroupName))
		subnetGroupLogger.Debug("checking tags for database subnet group")
		tagListInput := &rds.ListTagsForResourceInput{
			ResourceName: subnetGroup.DBSubnetGroupArn,
		}
		tagListOutput, err := r.rdsClient.ListTagsForResource((tagListInput))

		if err != nil {
			return nil, errors.WrapLog(err, "failed to list tags for database subnet group", r.logger)
		}

		subnetGroupLogger.Debugf("checking for cluster tag match (%s=%s) on subnet group", tagKeyClusterId, clusterId)

		if findTag(tagKeyClusterId, clusterId, tagListOutput.TagList) == nil {
			subnetGroupLogger.Debugf("subnet group did not contain cluster tag match (%s=%s)", tagKeyClusterId, clusterId)
			continue
		}
		subnetGroupsToDelete = append(subnetGroupsToDelete, subnetGroup)
		r.logger.Debugf("filtering complete, %d subnet groups matched", len(subnetGroupsToDelete))
	}
	reportItems := make([]*clusterservice.ReportItem, 0)
	for _, dbSubnetGroup := range subnetGroupsToDelete {
		subnetGroupLogger := r.logger.WithField(loggingKeySubnetGroup, aws.StringValue(dbSubnetGroup.DBSubnetGroupName))
		subnetGroupLogger.Debug("creating report for rds subnet group")

		reportItem := &clusterservice.ReportItem{
			ID:           aws.StringValue(dbSubnetGroup.DBSubnetGroupArn),
			Name:         aws.StringValue(dbSubnetGroup.DBSubnetGroupName),
			Action:       clusterservice.ActionDelete,
			ActionStatus: clusterservice.ActionStatusEmpty,
		}
		reportItems = append(reportItems, reportItem)

		if dryRun {
			subnetGroupLogger.Debug("dry run enabled, skipping deletion step")
			reportItem.ActionStatus = clusterservice.ActionStatusDryRun
			continue
		}
		deleteInput := &rds.DeleteDBSubnetGroupInput{
			DBSubnetGroupName: dbSubnetGroup.DBSubnetGroupName,
		}

		_, err := r.rdsClient.DeleteDBSubnetGroup(deleteInput)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "InvalidDBSubnetGroupStateFault" {
				subnetGroupLogger.Debug("the DB subnet group cannot be deleted because it's in use, skipping")
				reportItem.ActionStatus = clusterservice.ActionStatusSkipped
				continue
			}
			return nil, errors.WrapLog(err, "failed to delete rds db subnet group", subnetGroupLogger)
		}
		reportItem.ActionStatus = clusterservice.ActionStatusComplete
	}
	return reportItems, nil
}
