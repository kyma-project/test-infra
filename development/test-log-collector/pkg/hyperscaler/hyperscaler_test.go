package hyperscaler

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_extractHyperScalerFromCm(t *testing.T) {
	type args struct {
		configmap v1.ConfigMap
	}
	tests := []struct {
		name    string
		args    args
		want    Platform
		wantErr bool
	}{
		{
			name: "properly extracts gcp hyperscaler platform",
			args: args{configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{},
				Data: map[string]string{
					"provider": "gcp",
				},
			}},
			want:    GardenerGcp,
			wantErr: false,
		},
		{
			name: "properly extracts azure hyperscaler platform",
			args: args{configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{},
				Data: map[string]string{
					"provider": "azure",
				},
			}},
			want:    GardenerAzure,
			wantErr: false,
		},
		{
			name: "properly extracts unknown hyperscaler platform",
			args: args{configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{},
				Data: map[string]string{
					"provider": "digital-ocean",
				},
			}},
			want:    UnknownGardener,
			wantErr: false,
		},
		{
			name: "returns error when there's no 'provider' key",
			args: args{configmap: v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{},
				Data: map[string]string{
					"hyperscaler": "digital-ocean",
				},
			}},
			want:    UnknownGardener,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractHyperScalerFromCm(tt.args.configmap)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractHyperScalerFromCm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractHyperScalerFromCm() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractHyperScalerFromNode(t *testing.T) {
	type args struct {
		node v1.Node
	}
	tests := []struct {
		name string
		args args
		want Platform
	}{
		{
			name: "gke",
			args: args{node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "gke-some-suffix"},
			}},
			want: Gke,
		},
		{
			name: "aks",
			args: args{node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "aks-some-suffix"},
			}},
			want: Aks,
		},
		{
			name: "unknown platform",
			args: args{node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "some other name"},
			}},
			want: Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractHyperScalerFromNode(tt.args.node); got != tt.want {
				t.Errorf("extractHyperScalerFromNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
