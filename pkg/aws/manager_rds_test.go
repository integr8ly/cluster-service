package aws

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"
)

func TestRDSEngine_DeleteResourcesForCluster(t *testing.T) {
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
			name: "error when describing db instances fails",
			fields: fields{
				logger: fakeLogger,
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DescribeDBInstancesFunc = func(in1 *rds.DescribeDBInstancesInput) (output *rds.DescribeDBInstancesOutput, e error) {
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
			wantErr: "failed to describe database clusters: ",
		},
		{
			name: "error when listing tags for db instance fails",
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
			wantErr: "failed to list tags for database cluster: ",
		},
		{
			name: "report empty when no db instances match cluster id tag",
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
				// fakeRDSClientTagKey is the cluster id tag, so ensure the cluster is not fakeRDSClientTagVal to
				// ensure it will not match.
				clusterId: fmt.Sprintf("%s-modified", fakeRDSClientTagVal),
				// we only care about the cluster id tag for this test, leave additional tags empty
				tags:   map[string]string{},
				dryRun: true,
			},
			want: make([]*clusterservice.ReportItem, 0),
		},
		{
			name: "report empty when cluster id tag matches but additional tags do not",
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
				tags: map[string]string{
					"addTagKey": "addTagVal",
				},
				dryRun: true,
			},
			want: make([]*clusterservice.ReportItem, 0),
		},
		{
			name: "modify instance is called when deletion protection is enabled",
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
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeRDSClientInstanceARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress

				}),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.ModifyDBInstanceCalls()) != 1 {
					return errors.New("modify db instance call count should be 1")
				}
				if len(mock.DeleteDBInstanceCalls()) != 1 {
					return errors.New("delete db instance call count should be 1")
				}
				return nil
			},
		},
		{
			name: "modify instance is not called when deletion protection is disabled",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DescribeDBInstancesFunc = func(in1 *rds.DescribeDBInstancesInput) (output *rds.DescribeDBInstancesOutput, e error) {
							fakeDBInstance := fakeRDSClientDBInstance()
							fakeDBInstance.DeletionProtection = aws.Bool(false)
							return &rds.DescribeDBInstancesOutput{
								DBInstances: []*rds.DBInstance{
									fakeDBInstance,
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
				clusterId: fakeRDSClientTagVal,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeRDSClientInstanceARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress

				}),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.ModifyDBInstanceCalls()) != 0 {
					return errors.New("modify db instance call count should be 0")
				}
				if len(mock.DeleteDBInstanceCalls()) != 1 {
					return errors.New("delete db instance call count should be 1")
				}
				return nil
			},
		},
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
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeRDSClientInstanceARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusDryRun
				}),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.ModifyDBInstanceCalls()) != 0 {
					return errors.New("modify db instance call count should be 0")
				}
				if len(mock.DeleteDBInstanceCalls()) != 0 {
					return errors.New("delete db instance call count should be 0")
				}
				return nil
			},
		},
		{
			name: "delete is not performed if db instance is in state deleting",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						fakeDBInstance := fakeRDSClientDBInstance()
						fakeDBInstance.DBInstanceStatus = aws.String(statusDeleting)
						c.DescribeDBInstancesFunc = func(in1 *rds.DescribeDBInstancesInput) (output *rds.DescribeDBInstancesOutput, e error) {
							return &rds.DescribeDBInstancesOutput{
								DBInstances: []*rds.DBInstance{
									fakeDBInstance,
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
				dryRun:    false,
				clusterId: fakeRDSClientTagVal,
				tags:      map[string]string{},
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeRDSClientInstanceARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress

				}),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.DeleteDBInstanceCalls()) != 0 {
					return errors.New("delete db instance call count should be 0")
				}
				return nil
			},
		},
		{
			name: "automated snapshots are deleted and final snapshot is skipped",
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
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				mockReportItem(func(item *clusterservice.ReportItem) {
					item.ID = fakeRDSClientInstanceARN
					item.Name = fakeResourceIdentifier
					item.Action = clusterservice.ActionDelete
					item.ActionStatus = clusterservice.ActionStatusInProgress
				}),
			},
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.DeleteDBInstanceCalls()) != 1 {
					return errors.New("delete db instance call count should be 1")
				}

				callInput := mock.DeleteDBInstanceCalls()[0].In1
				if !aws.BoolValue(callInput.DeleteAutomatedBackups) {
					return errors.New("delete automated backups option must be true when deleting db instance")
				}
				if !aws.BoolValue(callInput.SkipFinalSnapshot) {
					return errors.New("skip final snapshot option must be true")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.fields.rdsClient()
			r := &RDSInstanceManager{
				rdsClient: fakeClient,
				logger:    tt.fields.logger,
			}
			got, err := r.DeleteResourcesForCluster(tt.args.clusterId, tt.args.tags, tt.args.dryRun)
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

func Test_findTag(t *testing.T) {
	testTagKey := "testKey"
	testTagVal := "testVal"

	testTag := &rds.Tag{
		Key:   aws.String(testTagKey),
		Value: aws.String(testTagVal),
	}

	type args struct {
		key   string
		value string
		tags  []*rds.Tag
	}
	tests := []struct {
		name string
		args args
		want *rds.Tag
	}{
		{
			name: "return tag if found",
			args: args{
				key:   testTagKey,
				value: testTagVal,
				tags: []*rds.Tag{
					testTag,
				},
			},
			want: testTag,
		},
		{
			name: "return nil if tag not found",
			args: args{
				key:   testTagKey,
				value: testTagVal,
				tags:  []*rds.Tag{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findTag(tt.args.key, tt.args.value, tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equalReportItems(a, b []*clusterservice.ReportItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if !reflect.DeepEqual(*a[i], *b[i]) {
			return false
		}
	}
	return true
}
