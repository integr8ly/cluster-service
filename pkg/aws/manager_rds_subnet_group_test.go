package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"
)

func TestRDSSubnetGroupManager_DeleteResourcesForCluster(t *testing.T) {
	fakeClusterId := "testClusterId"
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		rdsClient func() *rdsClientMock
		logger    *logrus.Entry
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
		wantErr string
	}{
		{
			name: "error when describing subnet groups fails",
			fields: fields{
				logger: fakeLogger,
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DescribeDBSubnetGroupsFunc = func(in1 *rds.DescribeDBSubnetGroupsInput) (output *rds.DescribeDBSubnetGroupsOutput, e error) {
							return nil, errors.New("")
						}
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
				tags:      map[string]string{},
				dryRun:    true,
			},
			wantErr: "failed to describe database subnet groups: ",
		},
		{
			name: "error when listing tags for subnet group fails",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.ListTagsForResourceFunc = func(in1 *rds.ListTagsForResourceInput) (output *rds.ListTagsForResourceOutput, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
			},
			wantErr: "failed to list tags for database subnet group: ",
		},
		{
			name: "error when deleting subnet group fails",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.ListTagsForResourceFunc = func(in1 *rds.ListTagsForResourceInput) (output *rds.ListTagsForResourceOutput, e error) {
							return &rds.ListTagsForResourceOutput{
								TagList: []*rds.Tag{
									{
										Key:   aws.String(fakeRDSClientTagKey),
										Value: aws.String(fakeClusterId),
									},
								},
							}, nil
						}
						c.DeleteDBSubnetGroupFunc = func(in *rds.DeleteDBSubnetGroupInput) (*rds.DeleteDBSubnetGroupOutput, error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
			},
			wantErr: "failed to delete rds db subnet group: ",
		},
		{
			name: "report empty when no subnet groups match cluster id tag",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DescribeDBSubnetGroupsFunc = func(in1 *rds.DescribeDBSubnetGroupsInput) (*rds.DescribeDBSubnetGroupsOutput, error) {
							return &rds.DescribeDBSubnetGroupsOutput{
								DBSubnetGroups: []*rds.DBSubnetGroup{
									{
										DBSubnetGroupArn: aws.String(fakeARN),
									},
								},
							}, nil
						}
						c.ListTagsForResourceFunc = func(in1 *rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error) {
							return &rds.ListTagsForResourceOutput{
								TagList: []*rds.Tag{
									{
										Key:   aws.String(tagKeyClusterId),
										Value: aws.String("not a real clusterId"),
									},
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
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				// we only care about the cluster id tag for this test, leave additional tags empty
				tags:   map[string]string{},
				dryRun: false,
			},
			want: make([]*clusterservice.ReportItem, 0),
		},
		// {
		// 	name: "report empty when cluster id tag matches but additional tags do not",
		// 	fields: fields{
		// 		rdsClient: func() *rdsClientMock {
		// 			fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
		// 				return nil
		// 			})
		// 			if err != nil {
		// 				t.Fatal(err)
		// 			}
		// 			return fakeClient
		// 		},
		// 		logger: fakeLogger,
		// 	},
		// 	args: args{
		// 		clusterId: fakeRDSClientTagVal,
		// 		tags: map[string]string{
		// 			"addTagKey": "addTagVal",
		// 		},
		// 		dryRun: true,
		// 	},
		// 	want: make([]*clusterservice.ReportItem, 0),
		// },
		{
			name: "no destructive methods are used when dry run is true",
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
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeRDSClientTagVal,
				tags:      map[string]string{},
				dryRun:    true,
			},
			want: []*clusterservice.ReportItem{
				fakeReportItemDryRun(),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.DeleteDBSubnetGroupCalls()) != 0 {
					return errors.New("delete db subnet groups call count should be 0")
				}
				return nil
			},
		},
		{
			name: "deleting subnet group is skipped when InvalidDBSubnetGroupStateFault error is returned",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DeleteDBSubnetGroupFunc = func(in *rds.DeleteDBSubnetGroupInput) (*rds.DeleteDBSubnetGroupOutput, error) {
							errorMsg := ""
							return nil, awserr.New("InvalidDBSubnetGroupStateFault", errorMsg, errors.New(errorMsg))
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeRDSClientTagVal,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				{
					ID:           fakeARN,
					Name:         fakeResourceIdentifier,
					Action:       clusterservice.ActionDelete,
					ActionStatus: clusterservice.ActionStatusSkipped,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.fields.rdsClient()
			r := &RDSSubnetGroupManager{
				rdsClient: fakeClient,
				logger:    tt.fields.logger,
			}
			got, err := r.DeleteResourcesForCluster(tt.args.clusterId, tt.args.tags, tt.args.dryRun)
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
