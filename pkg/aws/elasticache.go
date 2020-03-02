package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

var _ ActionEngine = &ElasticacheEngine{}
var TagFilters []*resourcegroupstaggingapi.TagFilter

type ElasticacheEngine struct {
	elasticacheClient elasticacheiface.ElastiCacheAPI
	taggingClient     resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
	logger            *logrus.Entry
}

func NewDefaultElastiCacheEngine(session *session.Session, logger *logrus.Entry) *ElasticacheEngine {
	return &ElasticacheEngine{
		elasticacheClient: elasticache.New(session),
		taggingClient: resourcegroupstaggingapi.New(session),
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
		databasesToDelete = append(databasesToDelete, replicationGroup)
		//tagListInput := &elasticache.ListTagsForResourceInput{
		//	ResourceName: replicationGroup.ReplicationGroupId,
		//}
		resourceInput := &resourcegroupstaggingapi.GetResourcesInput{
			ResourceTypeFilters: aws.StringSlice([]string{"elasticache:cluster"}),
			TagFilters: []*resourcegroupstaggingapi.TagFilter{
				{
					Key: aws.String(tagKeyClusterId),
					Values: aws.StringSlice([]string{
						clusterId,
					}),
				},
			},
		}
		resourceOutput, err := r.taggingClient.GetResources(resourceInput)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to Get resources for clusterID : %s", clusterId)
		}
		//TODO declare replication group id list
		var replicationGroupsToDelete []string

		for _, resourceTagMapping := range resourceOutput.ResourceTagMappingList {
			arn := aws.StringValue(resourceTagMapping.ResourceARN)
			arnSplit := strings.Split(arn, ":")
			cacheClusterId := arnSplit[len(arnSplit)-1]
			cacheClusterInput := &elasticache.DescribeCacheClustersInput{
				CacheClusterId: aws.String(cacheClusterId),
			}
			cacheClusterOutput, err := r.elasticacheClient.DescribeCacheClusters(cacheClusterInput)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot get cacheclusterOutput for : %s", cacheClusterInput)
			}
			for _, cacheCluster := range cacheClusterOutput.CacheClusters {
				for _, replicationGroupToDeleteID := range replicationGroupsToDelete {
					if !contains(replicationGroupsToDelete, *cacheCluster.ReplicationGroupId) {
						replicationGroupsToDelete = append(replicationGroupsToDelete, *replicationGroup.ReplicationGroupId)
					}
					if contains(replicationGroupsToDelete, *cacheCluster.ReplicationGroupId) {
						dbLogger.Debugf("Replication Group already exists in deletion list (%s=%s)", replicationGroupToDeleteID, clusterId)
						break
					}
				}
			}
			//delete each replication group in the list
			for _, replicationGroupID := range replicationGroupsToDelete {
				deleteReplicationGroupInput := &elasticache.DeleteReplicationGroupInput{
					ReplicationGroupId:   aws.String(replicationGroupID),
					RetainPrimaryCluster: aws.Bool(false),
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

					_, err := r.elasticacheClient.DeleteReplicationGroup(deleteReplicationGroupInput)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to delete elasticache replicationGroup, db=%s", *replicationGroup.ReplicationGroupId)
					}

				}
			}
		}
	}
	return nil, nil
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}










