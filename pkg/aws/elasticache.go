package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ ActionEngine = &ElasticacheEngine{}

type ElasticacheEngine struct {
	elasticacheClient elasticacheiface.ElastiCacheAPI
	logger            *logrus.Entry
}

func NewDefaultElastiCacheEngine(session *session.Session, logger *logrus.Entry) *ElasticacheEngine {
	return &ElasticacheEngine{
		elasticacheClient: elasticache.New(session),
		logger:            logger.WithField("engine", "aws_elasticache"),
	}
}

func (r *ElasticacheEngine) GetName() string {
	return "AWS elasticache Engine"
}

//Delete all RDS resources for a specified cluster
func (r *ElasticacheEngine) DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) ([]*clusterservice.ReportItem, error) {
	logger := r.logger.WithFields(logrus.Fields{"clusterId": clusterId, "dryRun": dryRun})
	logger.Debug("deleting resources for cluster")
	elasticacheReplicationGroupDescribeInput := &elasticache.DescribeReplicationGroupsInput{}
	elasticacheReplicationGroupDescribeOutput, err := r.elasticacheClient.DescribeReplicationGroups(elasticacheReplicationGroupDescribeInput)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe database clusters, clusterId=%s", clusterId)
	}
	var databasesToDelete []*elasticache.ReplicationGroup
	for _, replicationGroup := range elasticacheReplicationGroupDescribeOutput.ReplicationGroups {
		dbLogger := logger.WithField("elasticache", aws.StringValue(replicationGroup.ReplicationGroupId))
		dbLogger.Debug("checking tags database cluster")
		tagListInput := &elasticache.ListTagsForResourceInput{
			ResourceName: replicationGroup.ReplicationGroupId,
		}
		tagListOutput, err := r.elasticacheClient.ListTagsForResource(tagListInput)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list tags for database cluster, clusterId=%s db=%s", clusterId, *replicationGroup.ReplicationGroupId)
		}
		dbLogger.Debugf("checking for cluster tag match (%s=%s) on database", tagKeyClusterId, clusterId)
		if findElasitCacheTag(tagKeyClusterId, clusterId, tagListOutput.TagList) == nil {
			dbLogger.Debugf("database did not contain cluster tag match (%s=%s)", tagKeyClusterId, clusterId)
			continue
		}
		extraTagsMatch := true
		for extraTagKey, extraTagVal := range tags {
			dbLogger.Debugf("checking for additional tag match (%s=%s) on database", extraTagKey, extraTagVal)
			if findElasitCacheTag(extraTagKey, extraTagVal, tagListOutput.TagList) == nil {
				extraTagsMatch = false
				break
			}
		}
		if !extraTagsMatch {
			dbLogger.Debug("additional tags did not match, ignoring database")
			continue
		}
		databasesToDelete = append(databasesToDelete, replicationGroup)
	}
	logger.Debugf("filtering complete, %d databases matched", len(databasesToDelete))
	var reportItems []*clusterservice.ReportItem
	for _, replicationGroup := range databasesToDelete {
		dbLogger := logger.WithField("replicationGroup", aws.StringValue(replicationGroup.ReplicationGroupId))
		dbLogger.Debugf("building report for database")
		reportItem := &clusterservice.ReportItem{
			ID:           aws.StringValue(replicationGroup.ReplicationGroupId),
			Name:         "elasticache ReplicationGroup",
			Action:       clusterservice.ActionDelete,
			ActionStatus: clusterservice.ActionStatusEmpty,
		}
		reportItems = append(reportItems, reportItem)
		if dryRun {
			dbLogger.Debug("dry run enabled, skipping deletion step")
			reportItem.ActionStatus = clusterservice.ActionStatusDryRun
			continue
		}
		dbLogger.Debug("performing deletion of database")
		reportItem.ActionStatus = clusterservice.ActionStatusInProgress
		//deleting will return an error if the database is already in a deleting state
		if aws.StringValue(replicationGroup.Status) == statusDeleting {
			dbLogger.Debugf("deletion of database already in progress")
			continue
		}

		deleteInput := &elasticache.DeleteReplicationGroupInput{
			FinalSnapshotIdentifier: nil,
			ReplicationGroupId:      replicationGroup.ReplicationGroupId,
			RetainPrimaryCluster:    nil,
		}
		_, err := r.elasticacheClient.DeleteReplicationGroup(deleteInput)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to delete elasticache replicationGroup, db=%s", *replicationGroup.ReplicationGroupId)
		}
	}
	return reportItems, nil
}

func findElasitCacheTag(key, value string, tags []*elasticache.Tag) *elasticache.Tag {
	for _, tag := range tags {
		if key == aws.StringValue(tag.Key) && value == aws.StringValue(tag.Value) {
			return tag
		}
	}
	return nil
}
