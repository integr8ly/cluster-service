package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var _ clusterservice.Client = &Client{}

type Client struct {
	actionEngines []ActionEngine
	logger        *logrus.Entry
}

func NewDefaultClient(awsSession *session.Session, logger *logrus.Entry) *Client {
	log := logger.WithField("cluster_service_provider", "aws")
	rdsEngine := NewDefaultRDSEngine(awsSession, logger)
	return &Client{
		actionEngines: []ActionEngine{rdsEngine},
		logger:        log,
	}
}

//DeleteResourcesForCluster Delete AWS resources based on tags using provided action engines
func (c *Client) DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) (*clusterservice.Report, error) {
	c.logger.Debugf("deleting resources for cluster, clusterId=%s", clusterId)
	report := &clusterservice.Report{}
	for _, engine := range c.actionEngines {
		reportItems, err := engine.DeleteResourcesForCluster(clusterId, tags, dryRun)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to run engine, engine=%s", engine.GetName())
		}
		report.Items = append(report.Items, reportItems...)
	}
	return report, nil
}
