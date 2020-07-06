package main

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/olekukonko/tablewriter"

	"github.com/integr8ly/cluster-service/pkg/clusterservice"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsclusterservice "github.com/integr8ly/cluster-service/pkg/aws"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup [cluster id] [flags]",
	Short: "delete aws resources for an rhmi cluster",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		//pre-req checks
		clusterId := args[0]
		region, err := cmd.Flags().GetString("region")
		if err != nil {
			exitError(fmt.Sprintf("failed to get regions list from flag: %+v", err), exitCodeErrUnknown)
		}
		outputFormat, err := cmd.Flags().GetString("output")
		if err != nil {
			exitError(fmt.Sprintf("failed to get output format from flag: %+v", err), exitCodeErrUnknown)
		}
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			exitError(fmt.Sprintf("failed to get dry run from flag: %+v", err), exitCodeErrUnknown)
		}
		watch, err := cmd.Flags().GetBool("watch")
		if err != nil {
			exitError(fmt.Sprintf("failed to get watch from flag: %+v", err), exitCodeErrUnknown)
		}
		types, err := cmd.Flags().GetStringSlice("types")
		if err != nil {
			exitError(fmt.Sprintf("failed to get types from flag: %+v", err), exitCodeErrUnknown)
		}
		//ensure the output format is supported
		if outputFormat != "table" {
			exitError(fmt.Sprintf("output format %s not supported, use table", outputFormat), exitCodeErrKnown)
		}
		//setup aws session
		awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
		if awsKeyID == "" {
			exitError("AWS_ACCESS_KEY_ID env var must be defined", exitCodeErrKnown)
		}
		awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if awsSecretKey == "" {
			exitError("AWS_SECRET_ACCESS_KEY env var must be defined", exitCodeErrKnown)
		}
		awsSession := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(awsKeyID, awsSecretKey, ""),
		}))
		clusterService := buildAWSClientFromTypes(awsSession, types, logger)
		if watch {
			err := wait.PollImmediate(30*time.Second, 20*time.Minute, func() (bool, error) {
				var currentReport *clusterservice.Report
				newReport := runCleanupCommand(clusterService, clusterId, dryRun)
				if currentReport == nil {
					currentReport = newReport
				}
				currentReport.MergeForward(newReport)
				printReportTable(currentReport)
				logger.Info("watch is enabled, will attempt to delete resources every 30 seconds")
				return currentReport.AllItemsComplete(), nil
			})
			if err != nil {
				logger.Error(errors.Wrap(err, "failed to clean up all resources"))
			}
			logger.Info("finished cleaning up AWS resources")
		} else {
			report := runCleanupCommand(clusterService, clusterId, dryRun)
			printReportTable(report)
		}
	},
}

func runCleanupCommand(clusterService *awsclusterservice.Client, clusterId string, dryRun bool) *clusterservice.Report {
	report, err := clusterService.DeleteResourcesForCluster(clusterId, map[string]string{}, dryRun)
	if err != nil {
		exitError(fmt.Sprintf("failed to cleanup resources for cluster, clusterId=%s: %+v", clusterId, err), exitCodeErrUnknown)
	}
	return report
}

func buildAWSClientFromTypes(awsSession *session.Session, types []string, logger *logrus.Entry) *awsclusterservice.Client {
	if types == nil || len(types) == 0 {
		return awsclusterservice.NewDefaultClient(awsSession, logger)
	}
	client := &awsclusterservice.Client{
		Logger:           logger,
		ResourceManagers: make([]awsclusterservice.ClusterResourceManager, 0),
	}
	for _, t := range types {
		switch t {
		case "rds:instance":
			client.ResourceManagers = append(client.ResourceManagers, awsclusterservice.NewDefaultRDSInstanceManager(awsSession, logger))
		case "rds:snapshot":
			client.ResourceManagers = append(client.ResourceManagers, awsclusterservice.NewDefaultRDSSnapshotManager(awsSession, logger))
		case "s3":
			client.ResourceManagers = append(client.ResourceManagers, awsclusterservice.NewDefaultS3Engine(awsSession, logger))
		case "elasticache:replicationgroup":
			client.ResourceManagers = append(client.ResourceManagers, awsclusterservice.NewDefaultElasticacheManager(awsSession, logger))
		case "elasticache:snapshot":
			client.ResourceManagers = append(client.ResourceManagers, awsclusterservice.NewDefaultElasticacheSnapshotManager(awsSession, logger))
		case "ec2:subnet":
			client.ResourceManagers = append(client.ResourceManagers, awsclusterservice.NewDefaultSubnetManager(awsSession, logger))
		default:
			logger.Debugf("could not find resource manager for specified type %s", t)
		}
	}
	return client
}

func printReportTable(report *clusterservice.Report) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Action", "Status"})
	for _, reportItem := range report.Items {
		table.Append([]string{reportItem.ID, reportItem.Name, string(reportItem.Action), string(reportItem.ActionStatus)})
	}
	table.Render()
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().StringP("output", "o", "table", "set output format")
	cleanupCmd.Flags().StringP("region", "r", "eu-west-1", "region to delete resources in")
	cleanupCmd.Flags().BoolP("dry-run", "d", true, "skip performing actions")
	cleanupCmd.Flags().BoolP("watch", "w", false, "poll actions being performed indefinitely")
	cleanupCmd.Flags().StringSliceP("types", "t", []string{}, "resource types to cleanup")
}
