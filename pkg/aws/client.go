package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/integr8ly/cluster-service/pkg/errors"
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
	elasticacheEngine := NewDefaultElastiCacheEngine(awsSession, logger)
	s3Engine := NewDefaultS3Engine(awsSession, logger)
	return &Client{
		actionEngines: []ActionEngine{rdsEngine, s3Engine, elasticacheEngine},
		logger:        log,
	}
}

//DeleteResourcesForCluster Delete AWS resources based on tags using provided action engines
func (c *Client) DeleteResourcesForCluster(clusterId string, tags map[string]string, dryRun bool) (*clusterservice.Report, error) {
	logger := c.logger.WithFields(logrus.Fields{loggingKeyClusterID: clusterId, loggingKeyDryRun: dryRun})
	logger.Debugf("deleting resources for cluster")
	report := &clusterservice.Report{}
	for _, engine := range c.actionEngines {
		engineLogger := logger.WithField(loggingKeyEngine, engine.GetName())
		engineLogger.Debugf("found logger")
		reportItems, err := engine.DeleteResourcesForCluster(clusterId, tags, dryRun)
		if err != nil {
			return nil, errors.WrapLog(err, fmt.Sprintf("failed to run engine %s", engine.GetName()), engineLogger)
		}
		report.Items = append(report.Items, reportItems...)
	}
	return report, nil
}
