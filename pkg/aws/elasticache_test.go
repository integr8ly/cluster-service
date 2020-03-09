package aws

import (
	"github.com/aws/aws-sdk-go/aws"
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
		taggingClient     func() *resourcetaggingClientMock
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				fakeReportItemReplicationGroupDeleting(),
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				fakeReportItemReplicationGroupDeleting(),
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
				taggingClient: func() *resourcetaggingClientMock {
					fakeTaggingClient, err := fakeResourcetaggingClient(func(c *resourcetaggingClientMock) error {
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
				fakeReportItemReplicationGroupDryRun(),
			},
			wantFn: func(mock *elasticacheClientMock) error {
				if len(mock.DeleteReplicationGroupCalls()) != 0 {
					return errors.New("delete replication group call count should be 0 as dry run is true")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.fields.elasticacheClient()
			r := &ElasticacheEngine{
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
				t.Errorf("DeleteResourcesForCluster() got = %v, want %v", got, tt.want)
			}
			if tt.wantFn != nil {
				if err := tt.wantFn(fakeClient); err != nil {
					t.Errorf("DeleteResourcesForCluster() err = %v", err)
				}
			}
		})
	}
}
