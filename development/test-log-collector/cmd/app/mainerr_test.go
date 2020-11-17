package app

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
			name: "should pass on only 1 clustertestsuite which is not finished",
			args: args{ctsList: octopusTypes.ClusterTestSuiteList{
				Items: []octopusTypes.ClusterTestSuite{
					{
						Status: octopusTypes.TestSuiteStatus{
							CompletionTime: nil,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "testst",
						},
					},
				},
			}},
			want: octopusTypes.ClusterTestSuite{
				Status: octopusTypes.TestSuiteStatus{
					CompletionTime: nil,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "testst",
				},
			},
			wantErr: false,
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
			name: "should return first clustertestsuite",
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
			want: octopusTypes.ClusterTestSuite{
				Status: octopusTypes.TestSuiteStatus{
					CompletionTime: &date2009,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			got, err := getClusterTestSuite(tt.args.ctsList)
			// then
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func Test_getTestContainerName(t *testing.T) {
	type args struct {
		pod corev1.Pod
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1 non-istio container",
			args: args{pod: corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "that's not istio container",
				}},
			}}},
			want:    "that's not istio container",
			wantErr: false,
		},
		{
			name: "1 non-istio container + istio container",
			args: args{pod: corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "that's not istio container"},
					{Name: "istio-proxy"},
				},
			}}},
			want:    "that's not istio container",
			wantErr: false,
		},
		{
			name: "istio container only",
			args: args{pod: corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "istio-proxy"},
				},
			}}},
			want:    "",
			wantErr: true,
		},
		{
			name: "error on more than 1 non-istio containers",
			args: args{pod: corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "container1"},
					{Name: "container2"},
				},
			}}},
			want:    "",
			wantErr: true,
		},
		{
			name: "ignores init-containers",
			args: args{pod: corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Name: "that's not istio container",
				}},
				InitContainers: []corev1.Container{{
					Name: "istio-init",
				}},
			}}},
			want:    "that's not istio container",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			got, err := getTestContainerName(tt.args.pod)

			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func Test_extractTestStatus(t *testing.T) {
	type args struct {
		defName string
		cts     octopusTypes.ClusterTestSuite
	}
	tests := []struct {
		name    string
		args    args
		want    octopusTypes.TestStatus
		wantErr bool
	}{
		{
			name: "simplest case",
			args: args{
				defName: "specific-test-name",
				cts: octopusTypes.ClusterTestSuite{
					Status: octopusTypes.TestSuiteStatus{
						Results: []octopusTypes.TestResult{
							{
								Name:   "specific-test-name",
								Status: octopusTypes.TestSucceeded,
							},
							{
								Name:   "other-test-name",
								Status: octopusTypes.TestFailed,
							},
						},
					},
				},
			},
			wantErr: false,
			want:    octopusTypes.TestSucceeded,
		},
		{
			name: "no such test in results",
			args: args{
				defName: "not-in-results",
				cts: octopusTypes.ClusterTestSuite{
					ObjectMeta: metav1.ObjectMeta{Name: "some-cts-name"},
					Status: octopusTypes.TestSuiteStatus{
						Results: []octopusTypes.TestResult{
							{
								Name:   "specific-test-name",
								Status: octopusTypes.TestSucceeded,
							},
							{
								Name:   "other-test-name",
								Status: octopusTypes.TestFailed,
							},
						},
					},
				},
			},
			wantErr: true,
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			g := gomega.NewWithT(t)

			// when
			got, err := extractTestStatus(tt.args.defName, tt.args.cts)

			// then
			g.Expect(err != nil).To(gomega.Equal(tt.wantErr))
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
