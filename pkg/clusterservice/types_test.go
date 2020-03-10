package clusterservice

import (
	"reflect"
	"testing"
)

func TestReport_MergeForward(t *testing.T) {
	type fields struct {
		Items []*ReportItem
	}
	type args struct {
		mergeTarget *Report
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Report
	}{
		{
			name: "status moves from in progress to complete when item no longer exists",
			fields: fields{
				Items: []*ReportItem{
					{
						ID:           "test",
						Name:         "test",
						Action:       ActionDelete,
						ActionStatus: ActionStatusInProgress,
					},
				},
			},
			args: args{
				mergeTarget: &Report{
					Items: []*ReportItem{},
				},
			},
			want: &Report{
				Items: []*ReportItem{
					{
						ID:           "test",
						Name:         "test",
						Action:       ActionDelete,
						ActionStatus: ActionStatusComplete,
					},
				},
			},
		},
		{
			name: "overrides and appends work as expected",
			fields: fields{
				Items: []*ReportItem{
					{
						ID:           "willChange",
						Name:         "willChange",
						Action:       ActionDelete,
						ActionStatus: ActionStatusInProgress,
					},
				},
			},
			args: args{
				mergeTarget: &Report{
					Items: []*ReportItem{
						{
							ID:           "willChange",
							Name:         "willChange2",
							Action:       ActionDelete,
							ActionStatus: ActionStatusInProgress,
						},
						{
							ID:           "willAppend",
							Name:         "willAppend",
							Action:       ActionDelete,
							ActionStatus: ActionStatusInProgress,
						},
					},
				},
			},
			want: &Report{
				Items: []*ReportItem{
					{
						ID:           "willChange",
						Name:         "willChange2",
						Action:       ActionDelete,
						ActionStatus: ActionStatusInProgress,
					},
					{
						ID:           "willAppend",
						Name:         "willAppend",
						Action:       ActionDelete,
						ActionStatus: ActionStatusInProgress,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Report{
				Items: tt.fields.Items,
			}
			r.MergeForward(tt.args.mergeTarget)
			if !reflect.DeepEqual(tt.want, r) {
				t.Errorf("DeleteResourcesForCluster() got = %v, want %v", r, tt.want)
			}
		})
	}
}
