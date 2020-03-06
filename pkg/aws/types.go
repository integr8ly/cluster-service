package aws

import (
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
)

type EngineType string

const (
	tagKeyClusterId = "integreatly.org/clusterID"
	statusDeleting  = "deleting"

	engineRDS EngineType = "aws_rds"
	engineS3  EngineType = "aws_s3"

	loggingKeyClusterID = "cluster-id"
	loggingKeyDryRun    = "dry-run"
	loggingKeyEngine    = "engine"
)

//go:generate moq -out moq_actionengine_test.go . ActionEngine
//ActionEngine Perform actions for a specific resource
type ActionEngine interface {
	GetName() string
	DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) ([]*clusterservice.ReportItem, error)
}

//go:generate moq -out moq_rdsclient_test.go . rdsClient
//rdsClient alias for use with moq
type rdsClient interface {
	rdsiface.RDSAPI
}

//go:generate moq -out moq_elasticacheclient_test.go . elasticacheClient
//elasticacheClient alias for use with moq
type elasticacheClient interface {
	elasticacheiface.ElastiCacheAPI
}

//go:generate moq -out moq_s3client_test.go . s3Client
//s3Client alias for use with moq
type s3Client interface {
	s3iface.S3API
}

//go:generate moq -out moq_s3batchdeleteclient.go . s3BatchDeleteClient
//s3BatchDeleteClient alias for use with moq
type s3BatchDeleteClient interface {
	s3manageriface.BatchDelete
}

//go:generate moq -out moq_taggingclient_test.go . taggingClient
//taggingClient alias for use with moq
type taggingClient interface {
	resourcegroupstaggingapiiface.ResourceGroupsTaggingAPIAPI
}
