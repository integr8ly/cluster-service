package aws

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
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
		wantErr string
	}{
		{
			name: "error when getting listing resources with tagging api fails",
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
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (output *resourcegroupstaggingapi.GetResourcesOutput, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    true,
			},
			wantErr: "failed to filter rds subnet groups: ",
		},
		{
			name: "report empty when no subnet groups match cluster id tag",
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
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						c.GetResourcesFunc = func(in1 *resourcegroupstaggingapi.GetResourcesInput) (output *resourcegroupstaggingapi.GetResourcesOutput, e error) {
							return &resourcegroupstaggingapi.GetResourcesOutput{}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: make([]*clusterservice.ReportItem, 0),
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
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
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
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
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
			wantFn: func(mock *rdsClientMock) error {
				if len(mock.DeleteDBSubnetGroupCalls()) != 1 {
					return errors.New("delete db subnet groups call count should be 1")
				}
				return nil
			},
		},
		{
			name: "item is reported as deleted when deletion is successful",
			fields: fields{
				rdsClient: func() *rdsClientMock {
					fakeClient, err := fakeRDSClient(func(c *rdsClientMock) error {
						c.DeleteDBSubnetGroupFunc = func(in *rds.DeleteDBSubnetGroupInput) (*rds.DeleteDBSubnetGroupOutput, error) {
							return &rds.DeleteDBSubnetGroupOutput{}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return fakeClient
				},
				taggingClient: func() *taggingClientMock {
					client, err := fakeTaggingClient(func(c *taggingClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
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
					ActionStatus: clusterservice.ActionStatusComplete,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTaggingClient := tt.fields.taggingClient()
			fakeClient := tt.fields.rdsClient()
			r := &RDSSubnetGroupManager{
				rdsClient:     fakeClient,
				taggingClient: fakeTaggingClient,
				logger:        tt.fields.logger,
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
