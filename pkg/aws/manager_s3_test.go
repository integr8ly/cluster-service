package aws

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"

	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"
)

func TestS3Engine_DeleteResourcesForCluster(t *testing.T) {
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		s3Client            func() s3Client
		s3BatchDeleteClient func() s3BatchDeleteClient
		taggingClient       func() taggingClient
		logger              *logrus.Entry
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
		wantErr string
	}{
		{
			name: "fail when getting resources via tags returns an error",
			fields: fields{
				s3BatchDeleteClient: func() s3BatchDeleteClient {
					return nil
				},
				s3Client: func() s3Client {
					return nil
				},
				taggingClient: func() taggingClient {
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
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			wantErr: "failed to filter s3 buckets in aws: ",
		},
		{
			name: "fail when bucket batch deletion returns an error",
			fields: fields{
				s3BatchDeleteClient: func() s3BatchDeleteClient {
					client, err := fakeS3BatchClient(func(c *s3BatchDeleteClientMock) error {
						c.DeleteFunc = func(in1 context.Context, in2 s3manager.BatchDeleteIterator) error {
							return errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				s3Client: func() s3Client {
					client, err := fakeS3Client(func(c *s3ClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				taggingClient: func() taggingClient {
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
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			wantErr: "failed to empty bucket contents: ",
		},
		{
			name: "fail when bucket deletion returns an error",
			fields: fields{
				s3Client: func() s3Client {
					client, err := fakeS3Client(func(c *s3ClientMock) error {
						c.DeleteBucketFunc = func(in1 *s3.DeleteBucketInput) (output *s3.DeleteBucketOutput, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				s3BatchDeleteClient: func() s3BatchDeleteClient {
					client, err := fakeS3BatchClient(func(c *s3BatchDeleteClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				taggingClient: func() taggingClient {
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
			wantErr: "failed to delete bucket: ",
		},
		{
			name: "succeeds with status in progress if dry run is false",
			fields: fields{
				s3Client: func() s3Client {
					client, err := fakeS3Client(func(c *s3ClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				s3BatchDeleteClient: func() s3BatchDeleteClient {
					client, err := fakeS3BatchClient(func(c *s3BatchDeleteClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				taggingClient: func() taggingClient {
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
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: []*clusterservice.ReportItem{
				fakeReportItemDeleting(),
			},
		},
		{
			name: "succeeds with status dry run if dry run is true",
			fields: fields{
				s3Client: func() s3Client {
					client, err := fakeS3Client(func(c *s3ClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				s3BatchDeleteClient: func() s3BatchDeleteClient {
					client, err := fakeS3BatchClient(func(c *s3BatchDeleteClientMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return client
				},
				taggingClient: func() taggingClient {
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
				clusterId: fakeClusterId,
				tags:      map[string]string{},
				dryRun:    true,
			},
			want: []*clusterservice.ReportItem{
				fakeReportItemDryRun(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &S3Engine{
				s3Client:            tt.fields.s3Client(),
				s3BatchDeleteClient: tt.fields.s3BatchDeleteClient(),
				taggingClient:       tt.fields.taggingClient(),
				logger:              tt.fields.logger,
			}
			got, err := s.DeleteResourcesForCluster(tt.args.clusterId, tt.args.tags, tt.args.dryRun)
			if tt.wantErr != "" && err.Error() != tt.wantErr {
				t.Errorf("DeleteResourcesForCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteResourcesForCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}
