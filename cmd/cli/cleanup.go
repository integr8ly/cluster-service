package main

import (
	"fmt"
	"os"
	"time"

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
		clusterService := awsclusterservice.NewDefaultClient(awsSession, logger)
		//this could probably leverage channels
		var currentReport *clusterservice.Report
		for {
			newReport, err := clusterService.DeleteResourcesForCluster(clusterId, map[string]string{}, dryRun)
			if err != nil {
				exitError(fmt.Sprintf("failed to cleanup resources for cluster, clusterId=%s: %+v", clusterId, err), exitCodeErrUnknown)
			}
			if currentReport == nil {
				currentReport = newReport
			}
			currentReport.MergeForward(newReport)
			printReportTable(currentReport)
			if !watch {
				break
			}
			logger.Debug("watch is enabled, waiting 10 seconds before re-invoking")
			time.Sleep(time.Second * 10)
		}
	},
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
}
