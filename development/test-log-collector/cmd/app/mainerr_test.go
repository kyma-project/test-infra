package app

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	octopusTypes "github.com/kyma-project/test-infra/development/test-log-collector/pkg/resources/clustertestsuite/types"
)

func Test_getNewestClusterTestSuite(t *testing.T) {

	date2009 := metav1.NewTime(time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC))
	date2020 := metav1.NewTime(time.Date(2020, 11, 17, 20, 34, 58, 651387237, time.UTC))

	type args struct {
		ctsList octopusTypes.ClusterTestSuiteList
	}
	tests := []struct {
		name    string
		args    args
		want    octopusTypes.ClusterTestSuite
		wantErr bool
	}{
		{
			name: "should throw error if there's no clustertestsuites",
			args: args{ctsList: octopusTypes.ClusterTestSuiteList{
				Items: []octopusTypes.ClusterTestSuite{},
			}},
			want:    octopusTypes.ClusterTestSuite{},
			wantErr: true,
		},
		{
			name: "should throw error on only 1 clustertestsuite which is not finished",
			args: args{ctsList: octopusTypes.ClusterTestSuiteList{
				Items: []octopusTypes.ClusterTestSuite{
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: nil,
						},
					},
				},
			}},
			want:    octopusTypes.ClusterTestSuite{},
			wantErr: true,
		},
		{
			name: "should pass on 1 clustertestsuite",
			args: args{ctsList: octopusTypes.ClusterTestSuiteList{
				Items: []octopusTypes.ClusterTestSuite{
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: &date2009,
						},
					},
				},
			}},
			want: octopusTypes.ClusterTestSuite{Status: octopusTypes.TestSuiteStatus{
				CompletionTime: &date2009,
			}},
			wantErr: false,
		},
		{
			name: "should return newest clustertestsuite",
			args: args{ctsList: octopusTypes.ClusterTestSuiteList{
				Items: []octopusTypes.ClusterTestSuite{
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: &date2009,
						},
					},
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: &date2020,
						},
					},
				},
			}},
			want: octopusTypes.ClusterTestSuite{Status: octopusTypes.TestSuiteStatus{
				CompletionTime: &date2020,
			}},
			wantErr: false,
		},
		{
			name: "should return newest clustertestsuite",
			args: args{ctsList: octopusTypes.ClusterTestSuiteList{
				Items: []octopusTypes.ClusterTestSuite{
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: &date2009,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-name-no-status-completion-time",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-name-no-status-completion-time2",
						},
					},
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: &date2020,
						},
					},
				},
			}},
			want: octopusTypes.ClusterTestSuite{Status: octopusTypes.TestSuiteStatus{
				CompletionTime: &date2020,
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNewestClusterTestSuite(tt.args.ctsList)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNewestClusterTestSuite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNewestClusterTestSuite() got = %v, want %v", got, tt.want)
			}
		})
	}
}
