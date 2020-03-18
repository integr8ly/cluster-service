package aws

import (
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	//"reflect"
	"testing"
)

func TestElasticacheSnapshotManager_DeleteResourcesForCluster(t *testing.T) {
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
			name: "error when delete snapshot fails",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DeleteSnapshotFunc = func(in1 *elasticache.DeleteSnapshotInput) (output *elasticache.DeleteSnapshotOutput, e error) {
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
			wantErr: "failed to delete snapshot: ",
		}, {
			name: "pass when no report is returned if no snapshots deleted ",
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
			name: "pass when snapshots deleted and reportItem has status set to deleting",
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
				fakeReportItemElasticacheSnapshotDeleting(),
			},
			wantErr: "",
		},
		{
			name: "pass when no snapshots are deleted if dry run is true",
			fields: fields{
				elasticacheClient: func() *elasticacheClientMock {
					fakeClient, err := fakeElasticacheClient(func(c *elasticacheClientMock) error {
						c.DescribeSnapshotsFunc = func(in1 *elasticache.DescribeSnapshotsInput) (output *elasticache.DescribeSnapshotsOutput, err error) {
							return &elasticache.DescribeSnapshotsOutput{
								Snapshots: []*elasticache.Snapshot{
									fakeElasticacheSnapshot(),
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
				fakeReportItemElasticacheSnapshotDryRun(),
			},
			wantFn: func(mock *elasticacheClientMock) error {
				if len(mock.DeleteSnapshotCalls()) != 0 {
					return errors.New("delete snapshot call count should be 0 as dry run is true")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.fields.elasticacheClient()
			r := &ElasticacheSnapshotManager{
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
