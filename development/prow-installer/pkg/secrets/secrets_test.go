package secrets

import (
	"context"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	passedCtx := context.Background()

	type args struct {
		ctx  context.Context
		opts Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name:    "No options",
			args:    args{ctx: passedCtx, opts: Option{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Only prefix options",
			args:    args{ctx: passedCtx, opts: Option{Prefix: "hello"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Only project options",
			args:    args{ctx: passedCtx, opts: Option{ProjectID: "hello"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Only location options",
			args:    args{ctx: passedCtx, opts: Option{LocationID: "hello"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Only prefix, project options",
			args:    args{ctx: passedCtx, opts: Option{Prefix: "hello", ProjectID: "hello"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Only prefix, location options",
			args:    args{ctx: passedCtx, opts: Option{Prefix: "hello", LocationID: "hello"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Only project, location options",
			args:    args{ctx: passedCtx, opts: Option{ProjectID: "hello", LocationID: "hello"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "All required options",
			args:    args{ctx: passedCtx, opts: Option{Prefix: "hello", ProjectID: "hello", LocationID: "hello"}},
			want:    &Client{ctx: passedCtx, Option: Option{Prefix: "hello", ProjectID: "hello", LocationID: "hello"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
