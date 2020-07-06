package aws

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestRDSSnapshotManager_DeleteResourcesForCluster(t *testing.T) {
	fakeClusterId := "testClusterId"
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	type fields struct {
		rdsClient     func() *rdsClientMock
		taggingClient func() *taggingClientMock
		logger        *logrus.Entry
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
		wantFn  func(mock *rdsClientMock) error
		wantErr bool
	}{
		{
			name: "pass when snapshots deleted and reportItem has status set to deleting",
			fields: fields{
				logger: fakeLogger,
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					fakeClient, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
			},
			args: args{
				clusterId: fakeClusterId,
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress

				}),
			},
			wantErr: false,
		}, {
			name: "error when delete snapshot fails",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DeleteDBSnapshotFunc = func(in1 *rds.DeleteDBSnapshotInput) (output *rds.DeleteDBSnapshotOutput, e error) {
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
			wantErr: true,
		}, {
			name: "pass when no report is returned if no snapshots deleted ",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
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
			wantErr: true,
		},
		{
			name: "pass when no snapshots are deleted if dry run is true",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DescribeDBSnapshotsFunc = func(in1 *rds.DescribeDBSnapshotsInput) (*rds.DescribeDBSnapshotsOutput, error) {
							return &rds.DescribeDBSnapshotsOutput{
								DBSnapshots: []*rds.DBSnapshot{
									fakeRDSSnapshot(),
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
					item.ID = fakeARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusDryRun
				}),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.DeleteDBSnapshotCalls()) != 0 {
					return errors.New("delete snapshot call count should be 0 as dry run is true")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.fields.rdsClient()
			r := &RDSSnapshotManager{
				rdsClient:     fakeClient,
				taggingClient: tt.fields.taggingClient(),
				logger:        tt.fields.logger,
			}
			got, err := r.DeleteResourcesForCluster(tt.args.clusterId, tt.args.tags, tt.args.dryRun)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteResourcesForCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				for _, ri := range got {
					t.Errorf("DeleteResourcesForCluster() got : %v, %v, %v, %v", ri.ID, ri.ActionStatus, ri.Action, ri.Name)
				}
				for _, ri := range tt.want {
					t.Errorf("DeleteResourcesForCluster() want : %v, %v, %v, %v", ri.ID, ri.ActionStatus, ri.Action, ri.Name)
				}
			}
			if tt.wantFn != nil {
				if err := tt.wantFn(fakeClient); err != nil {
					t.Errorf("DeleteResourcesForCluster() err = %v", err)
				}
			}
		})
	}
}
