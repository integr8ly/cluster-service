package aws

import "github.com/integr8ly/cluster-service/pkg/clusterservice"

//ActionEngine Perform actions for a specific resource
type ActionEngine interface {
	GetName() string
	DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) ([]*clusterservice.ReportItem, error)
}
