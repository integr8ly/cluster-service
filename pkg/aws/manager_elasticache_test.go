package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	//"reflect"
	"testing"
)

func TestElasticacheEngine_DeleteResourcesForCluster(t *testing.T) {
	fakeClusterId := fakeClusterID
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		elasticacheClient func() *elasticacheClientMock
		taggingClient     func() *taggingClientMock
		logger            *logrus.Entry
	}
	type args struct {
		clusterId string
		tags      map[string]string
		dryRun    bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*clusterservice.ReportItem
		wantFn  func(mock *elasticacheClientMock) error
		wantErr string
	}{
		{
			name: "error when describing clusters fail",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (output *resourcegroupstaggingapi.GetResourcesOutput, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}

					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    true,
			},
			wantErr: "failed to describe cache clusters: ",
		}, {
			name: "error when getting cacheCluster output",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DescribeCacheClustersFunc = func(in1 *elasticache.DescribeCacheClustersInput) (*elasticache.DescribeCacheClustersOutput, error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    true,
			},
			wantErr: "cannot get cacheCluster output: ",
		}, {
			name: "error when describe replicationGroups fail",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DescribeReplicationGroupsFunc = func(in1 *elasticache.DescribeReplicationGroupsInput) (output *elasticache.DescribeReplicationGroupsOutput, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			wantErr: "cannot describe replicationGroups: ",
		}, {
			name: "error when delete replicationGroups fail",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DeleteReplicationGroupFunc = func(in1 *elasticache.DeleteReplicationGroupInput) (output *elasticache.DeleteReplicationGroupOutput, err error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			wantErr: "failed to delete elasticache replication group: ",
		}, {
			name: "error when delete subnet groups fail",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DeleteCacheSubnetGroupFunc = func(in1 *elasticache.DeleteCacheSubnetGroupInput) (output *elasticache.DeleteCacheSubnetGroupOutput, err error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			wantErr: "failed to delete cache subnet group: ",
		}, {
			name: "pass when no report is returned  if no replicationgroups deleted ",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (output *resourcegroupstaggingapi.GetResourcesOutput, err error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    true,
			},
			want:    []*clusterservice.ReportItem{},
			wantErr: "",
		}, {
			name: "pass when replicationGroup deleted and reportItem has status set to deleting ",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheClientReplicationGroupId
					item.Name = fakeElasticacheClientName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress
				}),
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheSubnetGroupID
					item.Name = fakeElasticacheSubnetGroupName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusComplete
				}),
			},
			wantErr: "",
		}, {
			name: "pass when deleteReplicationGroup method isn't called if a replicationGroup is already deleting ",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						fakeReplicationGroup := fakeElasticacheReplicationGroup()
						fakeReplicationGroup.Status = aws.String(statusDeleting)
						c.DescribeReplicationGroupsFunc = func(in1 *elasticache.DescribeReplicationGroupsInput) (output *elasticache.DescribeReplicationGroupsOutput, err error) {
							return &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []*elasticache.ReplicationGroup{
									fakeReplicationGroup,
								},
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheClientReplicationGroupId
					item.Name = fakeElasticacheClientName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress
				}),
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheSubnetGroupID
					item.Name = fakeElasticacheSubnetGroupName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusComplete
				}),
			},
			wantFn: func(mock *elasticacheClientMock) error {
				if len(mock.DeleteReplicationGroupCalls()) != 0 {
					return errors.New("delete replication group call count should be 0")
				}
				return nil
			},
		}, {
			name: "pass when no replicationGroups are deleted if dry run is true",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DescribeReplicationGroupsFunc = func(in1 *elasticache.DescribeReplicationGroupsInput) (output *elasticache.DescribeReplicationGroupsOutput, err error) {
							return &elasticache.DescribeReplicationGroupsOutput{
								ReplicationGroups: []*elasticache.ReplicationGroup{
									fakeElasticacheReplicationGroup(),
								},
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    true,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheClientReplicationGroupId
					item.Name = fakeElasticacheClientName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusDryRun
				}),
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheSubnetGroupID
					item.Name = fakeElasticacheSubnetGroupName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusDryRun
				}),
			},
			wantFn: func(mock *elasticacheClientMock) error {
				if len(mock.DeleteReplicationGroupCalls()) != 0 {
					return errors.New("delete replication group call count should be 0 as dry run is true")
				}
				return nil
			},
		}, {
			name: "pass when no subnet groups are skipped elasticache client returns CacheSubnetGroupInUse",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DeleteCacheSubnetGroupFunc = func(in1 *elasticache.DeleteCacheSubnetGroupInput) (out *elasticache.DeleteCacheSubnetGroupOutput, err error) {
							errorMsg := "cache subnet group is still in use"
							return nil, awserr.New("CacheSubnetGroupInUse", errorMsg, errors.New(errorMsg))
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeTaggingClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheClientReplicationGroupId
					item.Name = fakeElasticacheClientName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress
				}),
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheSubnetGroupID
					item.Name = fakeElasticacheSubnetGroupName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusSkipped
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.fields.elasticacheClient()
			r := &ElasticacheManager{
				elasticacheClient: fakeClient,
				taggingClient:     tt.fields.taggingClient(),
				logger:            tt.fields.logger,
			}
			got, err := r.DeleteResourcesForCluster(tt.args.clusterId, nil, tt.args.dryRun)
			if tt.wantErr != "" && err.Error() != tt.wantErr {
				t.Errorf("DeleteResourcesForCluster() error = %v, wantErr %v", err, tt.wantErr)
				return

			}

			if !equalReportItems(got, tt.want) {
				t.Errorf("DeleteResourcesForCluster()\n\ngot:\n\n%v\n\nwant:\n\n%v\n\n", buildReportItemsString(got), buildReportItemsString(tt.want))
			}

			if tt.wantFn != nil {
				if err := tt.wantFn(fakeClient); err != nil {
					t.Errorf("DeleteResourcesForCluster() err = %v", err)
				}
			}
		})
	}
}

func buildReportItemsString(reportItems []*clusterservice.ReportItem) string {
	result := "["
	for _, reportItem := range reportItems {
		result += fmt.Sprintf("%+v,", *reportItem)
	}
	result += "]"
	return result
}

// test for the deletion of cache subnet groups which realistically only happens
// when DeleteResourcesForCluster() is called several times (watch mode)
// On the first attempt, a CacheSubnetGroupInUser error will be thrown so the report item status will be ActionStatusSkipped
// On the second attempt, an error won't be thrown, this time the status will be ActionStatusComplete
func TestElasticacheEngine_DeleteSubnetGroupsAcrossMultipleAttempts(t *testing.T) {
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})

	fakeTaggingClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	deleteResourcesForClusterCallCount := 0

	fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
		c.DescribeReplicationGroupsFunc = func(in1 *elasticache.DescribeReplicationGroupsInput) (output *elasticache.DescribeReplicationGroupsOutput, err error) {
			if deleteResourcesForClusterCallCount > 0 {
				// it's been called before this time return an empty result
				return &elasticache.DescribeReplicationGroupsOutput{
					ReplicationGroups: []*elasticache.ReplicationGroup{},
				}, nil
			}
			return &elasticache.DescribeReplicationGroupsOutput{
				ReplicationGroups: []*elasticache.ReplicationGroup{
					fakeElasticacheReplicationGroup(),
				},
			}, nil
		}
		c.DeleteCacheSubnetGroupFunc = func(in1 *elasticache.DeleteCacheSubnetGroupInput) (out *elasticache.DeleteCacheSubnetGroupOutput, err error) {
			if deleteResourcesForClusterCallCount > 0 {
				// DeleteResourcesForCluster has been called more than once,
				// this time set the replication group status as deleting
				return &elasticache.DeleteCacheSubnetGroupOutput{}, nil
			}
			deleteResourcesForClusterCallCount++
			errorMsg := "cache subnet group is still in use"
			return nil, awserr.New("CacheSubnetGroupInUse", errorMsg, errors.New(errorMsg))
		}
		return nil
	})

	if err != nil {
		t.Fatal(err)
	}

	manager := &ElasticacheManager{
		elasticacheClient: fakeClient,
		taggingClient:     fakeTaggingClient,
		logger:            fakeLogger,
	}

	attempts := []struct {
		wantReport                     []*clusterservice.ReportItem
		wantSubnetGroupsToDeleteLength int
	}{
		{
			wantReport: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheClientReplicationGroupId
					item.Name = fakeElasticacheClientName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress

				}),
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheSubnetGroupID
					item.Name = fakeElasticacheSubnetGroupName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusSkipped
				}),
			},
			wantSubnetGroupsToDeleteLength: 1,
		}, {
			wantReport: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeElasticacheClientReplicationGroupId
					item.Name = fakeElasticacheClientName
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress

				}),
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusComplete
					item.ID = fakeElasticacheSubnetGroupID
					item.Name = fakeElasticacheSubnetGroupName
				}),
			},
			wantSubnetGroupsToDeleteLength: 0,
		},
	}

	for i, attempt := range attempts {
		gotReport, err := manager.DeleteResourcesForCluster(fakeClusterID, nil, false)

		if err != nil {
			t.Error(err)
		}

		if !equalReportItems(gotReport, attempt.wantReport) {
			t.Errorf("DeleteResourcesForCluster() Attempt Number %v\n\ngot:\n\n%v\n\nwant:\n\n%v\n\n", i+1, buildReportItemsString(gotReport), buildReportItemsString(attempt.wantReport))
		}

		subnetGroupsToDeleteLength := len(manager.subnetGroupsToDelete)

		if subnetGroupsToDeleteLength != attempt.wantSubnetGroupsToDeleteLength {
			t.Errorf("DeleteResourcesForCluster() Attempt Number %v Wrong length for subnetGroupsToDelete \n\ngot: %v\nwant:%v", i+1, subnetGroupsToDeleteLength, attempt.wantSubnetGroupsToDeleteLength)
		}
	}
}
