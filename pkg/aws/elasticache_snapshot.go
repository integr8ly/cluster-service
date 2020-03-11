package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/integr8ly/cluster-service/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

var _ ActionEngine = &ElasticacheSnapshotEngine{}

type ElasticacheSnapshotEngine struct {
	elasticacheClient elasticacheiface.ElastiCacheAPI
	taggingClient     resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
	logger            *logrus.Entry
}

func newDefaultElasticacheSnapshotEngine(session *session.Session, logger *logrus.Entry) *ElasticacheSnapshotEngine {
	return &ElasticacheSnapshotEngine{
		elasticacheClient: elasticache.New(session),
		taggingClient:     resourcegroupstaggingapi.New(session),
		logger:            logger.WithField("engine", "aws_elasticache_snapshot"),
	}
}

func (r *ElasticacheSnapshotEngine) GetName() string {
	return "AWS elasticache Snapshot Engine"
}

func (r *ElasticacheSnapshotEngine) DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) ([]*clusterservice.ReportItem, error) {
	logger := r.logger.WithFields(logrus.Fields{"clusterId": clusterId, "dryRun": dryRun})
	logger.Debug("deleting resources for cluster")

	var reportItems []*clusterservice.ReportItem
	//collection of clusterID's for respective snapshots
	var snapshotsToDeleteCacheClusterId []string

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
		return nil, errors.WrapLog(err, "failed to describe cache clusters", logger)
	}
	for _, resourceTagMapping := range resourceOutput.ResourceTagMappingList {
		arn := aws.StringValue(resourceTagMapping.ResourceARN)
		arnSplit := strings.Split(arn, ":")
		cacheClusterId := arnSplit[len(arnSplit)-1]
		print(cacheClusterId)
		cacheClusterInput := &elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(cacheClusterId),
		}
		cacheClusterOutput, err := r.elasticacheClient.DescribeCacheClusters(cacheClusterInput)
		if err != nil {
			return nil, errors.WrapLog(err, "cannot get cacheCluster output", logger)
		}
		for _, cacheCluster := range cacheClusterOutput.CacheClusters {
			rgLogger := logger.WithField("snapshotCacheClusterID", cacheCluster.CacheClusterId)
			if contains(snapshotsToDeleteCacheClusterId, *cacheCluster.CacheClusterId) {
				rgLogger.Debugf("cacheCluster already exists in deletion list (%s=%s)", *cacheCluster.CacheClusterId, clusterId)
				break
			}
			snapshotsToDeleteCacheClusterId = append(snapshotsToDeleteCacheClusterId, *cacheCluster.CacheClusterId)
		}
	}

	logger.Debugf("filtering complete, %d cacheclusters matched", len(snapshotsToDeleteCacheClusterId))
	for _, cacheClusterId := range snapshotsToDeleteCacheClusterId {
		//delete each cacheCluster in the list
		ssLogger := logger.WithField("cacheClusterID", aws.String(cacheClusterId))
		ssLogger.Debugf("building report for database")
		reportItem := &clusterservice.ReportItem{
			ID:           cacheClusterId,
			Name:         "elasticache snapshot",
			Action:       clusterservice.ActionDelete,
			ActionStatus: clusterservice.ActionStatusInProgress,
		}
		reportItems = append(reportItems, reportItem)
		if dryRun {
			ssLogger.Debug("dry run enabled, skipping deletion step")
			reportItem.ActionStatus = clusterservice.ActionStatusDryRun
			continue
		}
		ssLogger.Debug("performing deletion of snapshotOutput cachecluster")
		snapshotInput := &elasticache.DescribeSnapshotsInput{
			CacheClusterId: aws.String(cacheClusterId),
		}
		snapshotOutput, err := r.elasticacheClient.DescribeSnapshots(snapshotInput)
		if err != nil {
			return nil, errors.WrapLog(err, "cannot Describe snapshotInput", logger)
		}
		if len(snapshotOutput.Snapshots) > 0 && aws.StringValue(snapshotOutput.Snapshots[0].SnapshotStatus) == statusDeleting {
			ssLogger.Debugf("deletion of snapshots already in progress")
			reportItem.ActionStatus = clusterservice.ActionStatusInProgress
			continue
		}
		var snapshotNamesToDelete []string
		for _, snapshot := range snapshotOutput.Snapshots {
			ssLogger := logger.WithField("snapshotName", snapshot.SnapshotName)
			if contains(snapshotNamesToDelete, *snapshot.SnapshotName) {
				ssLogger.Debugf("snapshot already exists in deletion list (%s=%s)", *snapshot.SnapshotName, clusterId)
				break
			}
			snapshotNamesToDelete = append(snapshotNamesToDelete, *snapshot.SnapshotName)
		}
		for _, snapshotName := range snapshotNamesToDelete {
			deleteSnapshotInput := &elasticache.DeleteSnapshotInput{
				SnapshotName: aws.String(snapshotName),
			}
			if _, err := r.elasticacheClient.DeleteSnapshot(deleteSnapshotInput); err != nil {
				return nil, errors.WrapLog(err, "failed to delete snapshot", logger)
			}
		}
	}
	if reportItems != nil {
		return reportItems, nil
	}
	return nil, nil
}
