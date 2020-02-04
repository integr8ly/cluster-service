package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ ActionEngine = &RDSEngine{}

type RDSEngine struct {
	rdsClient rdsiface.RDSAPI
	logger    *logrus.Entry
}

func NewDefaultRDSEngine(session *session.Session, logger *logrus.Entry) *RDSEngine {
	return &RDSEngine{
		rdsClient: rds.New(session),
		logger:    logger.WithField("engine", "aws_rds"),
	}
}

func (r *RDSEngine) GetName() string {
	return "AWS RDS Engine"
}

//Delete all RDS resources for a specified cluster
func (r *RDSEngine) DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) ([]*clusterservice.ReportItem, error) {
	r.logger.Debugf("deleting resources for cluster, clusterId=%s dryRun=%t", clusterId, dryRun)
	clusterDescribeInput := &rds.DescribeDBInstancesInput{}
	clusterDescribeOutput, err := r.rdsClient.DescribeDBInstances(clusterDescribeInput)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe database clusters, clusterId=%s", clusterId)
	}
	var databasesToDelete []*rds.DBInstance
	for _, dbCluster := range clusterDescribeOutput.DBInstances {
		r.logger.Debugf("checking tags database cluster, clusterId=%s db=%s", clusterId, dbCluster.DBInstanceIdentifier)
		tagListInput := &rds.ListTagsForResourceInput{
			ResourceName: dbCluster.DBInstanceArn,
		}
		tagListOutput, err := r.rdsClient.ListTagsForResource(tagListInput)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list tags for database cluster, clusterId=%s db=%s", clusterId, dbCluster.DBInstanceIdentifier)
		}
		r.logger.Debugf("checking cluster tag on database, clusterId=%s db=%s tagName=%s", clusterId, dbCluster.DBInstanceIdentifier, tagKeyClusterId)
		if findTag(tagKeyClusterId, clusterId, tagListOutput.TagList) == nil {
			r.logger.Debugf("database did not contain cluster tag, skipping, clusterId=%s db=%s tagName=%s", clusterId, dbCluster.DBInstanceIdentifier, tagKeyClusterId)
			continue
		}
		extraTagsMatch := true
		for extraTagKey, extraTagVal := range tags {
			r.logger.Debugf("checking extra tag on database, clusterId=%s db=%s tagKey=%s tagVal=%s", clusterId, dbCluster.DBInstanceIdentifier, extraTagKey, extraTagVal)
			if findTag(extraTagKey, extraTagVal, tagListOutput.TagList) == nil {
				extraTagsMatch = false
				break
			}
		}
		if !extraTagsMatch {
			r.logger.Debugf("additional tags did not match, ignoring database, clusterId=%s db=%s", clusterId, dbCluster.DBInstanceIdentifier)
			continue
		}
		databasesToDelete = append(databasesToDelete, dbCluster)
	}
	r.logger.Debugf("filtering complete, %d databases were found to match filters, building report", len(databasesToDelete))
	var reportItems []*clusterservice.ReportItem
	for _, dbInstance := range databasesToDelete {
		reportItem := &clusterservice.ReportItem{
			ID:           aws.StringValue(dbInstance.DBInstanceArn),
			Name:         aws.StringValue(dbInstance.DBClusterIdentifier),
			Action:       clusterservice.ActionDelete,
			ActionStatus: clusterservice.ActionStatusEmpty,
		}
		reportItems = append(reportItems, reportItem)
		if dryRun {
			r.logger.Debugf("dry run enabled, skipping deletion step")
			reportItem.ActionStatus = clusterservice.ActionStatusDryRun
			continue
		}
		r.logger.Debugf("performing deletion of database, db=%s", dbInstance.DBInstanceIdentifier)
		reportItem.ActionStatus = clusterservice.ActionStatusInProgress
		//deleting will return an error if the database is already in a deleting state
		if aws.StringValue(dbInstance.DBInstanceStatus) == statusDeleting {
			r.logger.Debugf("deletion of database already in progress, db=%s", dbInstance.DBClusterIdentifier)
			continue
		}
		if aws.BoolValue(dbInstance.DeletionProtection) {
			r.logger.Debugf("removing deletion protection on database, db=%s", dbInstance.DBInstanceIdentifier)
			modifyInput := &rds.ModifyDBInstanceInput{
				DBInstanceIdentifier: dbInstance.DBInstanceIdentifier,
				DeletionProtection:   aws.Bool(false),
			}
			modifyOutput, err := r.rdsClient.ModifyDBInstance(modifyInput)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to remove instance protection on database, db=%s", dbInstance.DBInstanceIdentifier)
			}
			dbInstance = modifyOutput.DBInstance
		}

		deleteInput := &rds.DeleteDBInstanceInput{
			DBInstanceIdentifier:   dbInstance.DBInstanceIdentifier,
			DeleteAutomatedBackups: aws.Bool(true),
			SkipFinalSnapshot:      aws.Bool(true),
		}
		_, err := r.rdsClient.DeleteDBInstance(deleteInput)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to delete rds instance, db=%s", dbInstance.DBInstanceIdentifier)
		}
	}
	return reportItems, nil
}

func findTag(key, value string, tags []*rds.Tag) *rds.Tag {
	for _, tag := range tags {
		if key == aws.StringValue(tag.Key) && value == aws.StringValue(tag.Value) {
			return tag
		}
	}
	return nil
}
