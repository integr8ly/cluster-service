package aws

import (
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
)

const (
	tagKeyClusterId = "integreatly.org/clusterID"
	statusDeleting  = "deleting"
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
