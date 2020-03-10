package aws

import (
	"errors"
	"reflect"
	"testing"

	"github.com/integr8ly/cluster-service/pkg/clusterservice"
	"github.com/sirupsen/logrus"
)

func TestClient_DeleteResourcesForCluster(t *testing.T) {
	fakeLogger, err := fakeLogger(func(l *logrus.Entry) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		actionEngines func() []ClusterResourceManager
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
		want    *clusterservice.Report
		wantErr string
	}{
		{
			name: "error when an engine fails",
			fields: fields{
				actionEngines: func() []ClusterResourceManager {
					fakeEngine, err := fakeActionEngine(func(e *ActionEngineMock) error {
						e.DeleteResourcesForClusterFunc = func(clusterId string, tags map[string]string, dryRun bool) (items []*clusterservice.ReportItem, e error) {
							return nil, errors.New("")
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return []ClusterResourceManager{fakeEngine}
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeRDSClientTagVal,
				tags:      map[string]string{},
				dryRun:    false,
			},
			wantErr: "failed to run engine Fake Action Engine: ",
		},
		{
			name: "multiple engine report items are appended",
			fields: fields{
				actionEngines: func() []ClusterResourceManager {
					fakeEngine, err := fakeActionEngine(func(e *ActionEngineMock) error {
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					fakeDryRunEngine, err := fakeActionEngine(func(e *ActionEngineMock) error {
						e.DeleteResourcesForClusterFunc = func(clusterId string, tags map[string]string, dryRun bool) (items []*clusterservice.ReportItem, e error) {
							return []*clusterservice.ReportItem{
								fakeReportItemDryRun(),
							}, nil
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					return []ClusterResourceManager{fakeEngine, fakeDryRunEngine}
				},
				logger: fakeLogger,
			},
			args: args{
				clusterId: fakeRDSClientTagVal,
				tags:      map[string]string{},
				dryRun:    false,
			},
			want: &clusterservice.Report{
				Items: []*clusterservice.ReportItem{
					fakeReportItemDeleting(),
					fakeReportItemDryRun(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				ResourceManagers: tt.fields.actionEngines(),
				Logger:           tt.fields.logger,
			}
			got, err := c.DeleteResourcesForCluster(tt.args.clusterId, tt.args.tags, tt.args.dryRun)
			if err != nil && err.Error() != tt.wantErr {
				t.Errorf("DeleteResourcesForCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteResourcesForCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}
