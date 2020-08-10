package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/onsi/gomega"
)

func Test_contains(t *testing.T) {
	type args struct {
		slice   []string
		element string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if slice includes particular string",
			args: args{
				slice:   []string{"1", "2", "3"},
				element: "1",
			},
			want: true,
		},
		{
			name: "returns false if slice doesnt include particular string",
			args: args{
				slice:   []string{"1", "2", "3"},
				element: "4",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.slice, tt.args.element); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDispatching_GetConfigByName(t *testing.T) {
	type fields struct {
		Config []LogsScrapingConfig
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    LogsScrapingConfig
		wantErr bool
	}{
		{
			name: "return proper config from slice",
			fields: fields{Config: []LogsScrapingConfig{
				{TestCases: []string{"test-name1"}, ChannelName: "test-name1"},
				{TestCases: []string{"test-name2"}, ChannelName: "test-name2"},
				{TestCases: []string{"test-name3"}, ChannelName: "something"},
			}},
			args:    args{name: "test-name1"},
			want:    LogsScrapingConfig{ChannelName: "test-name1", TestCases: []string{"test-name1"}},
			wantErr: false,
		},
		{
			name: "return first config that fits criteria",
			fields: fields{Config: []LogsScrapingConfig{
				{
					TestCases: []string{
						"test-name1",
						"test-name2",
						"test-name3",
					},
					ChannelName: "test-channel-name",
				},
				{TestCases: []string{"test-name4"}, ChannelName: "test-channel-name2"},
			}},
			args: args{name: "test-name3"},
			want: LogsScrapingConfig{
				TestCases: []string{
					"test-name1",
					"test-name2",
					"test-name3",
				},
				ChannelName: "test-channel-name"},
			wantErr: false,
		},
		{
			name: "return error if there's no config that fits criteria",
			fields: fields{Config: []LogsScrapingConfig{
				{TestCases: []string{"test-name1"}, ChannelName: "test-name1"},
				{TestCases: []string{"test-name2"}, ChannelName: "test-name2"},
				{TestCases: []string{"test-name3"}, ChannelName: "something"},
			}},
			args:    args{name: "test-name4"},
			want:    LogsScrapingConfig{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dispatching{
				Config: tt.fields.Config,
			}
			got, err := d.GetConfigByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigByName() error = %+v, wantErr %+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfigByName() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLoadDispatchingConfig(t *testing.T) {
	t.Run("properly reads configuration", func(t *testing.T) {
		g := gomega.NewWithT(t)

		// setup

		content := []byte(`- channelName: "#work"
  channelID: "chanID1"
  onlyReportFailure: false
  testCases:
    - "rafter"
- channelName: "#serverless-test"
  channelID: "chanID2"
  onlyReportFailure: true
  testCases:
    - serverless-long
    - serverless`)
		tmpfile, err := ioutil.TempFile("", "example")
		g.Expect(err).To(gomega.Succeed())

		defer func() {
			g.Expect(os.Remove(tmpfile.Name())).To(gomega.Succeed()) // clean up
		}()

		_, err = tmpfile.Write(content)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		defer func() {
			g.Expect(tmpfile.Close()).Should(gomega.Succeed())
		}()

		// proper test phase

		conf, err := LoadDispatchingConfig(tmpfile.Name())
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(conf).To(gomega.Equal(Dispatching{Config: []LogsScrapingConfig{
			{
				ChannelID:         "chanID1",
				ChannelName:       "#work",
				OnlyReportFailure: false,
				TestCases:         []string{"rafter"},
			},
			{
				ChannelName:       "#serverless-test",
				ChannelID:         "chanID2",
				OnlyReportFailure: true,
				TestCases:         []string{"serverless-long", "serverless"},
			},
		}}))
	})
	t.Run("errors on wrong file path", func(t *testing.T) {
		g := gomega.NewWithT(t)

		conf, err := LoadDispatchingConfig("some-file-that-does-not-exists.txt")
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(conf).To(gomega.Equal(Dispatching{Config: nil}))
	})
	t.Run("errors on misshaped config file", func(t *testing.T) {
		g := gomega.NewWithT(t)

		// setup

		content := []byte(`- channelNamez: "#work"
- channelName2: "#serverless-test"
  channelID2: "chanID2"
  testCases:
    mapInsteadOfSlice:
		some: "data"
  `)
		tmpfile, err := ioutil.TempFile("", "example")
		g.Expect(err).To(gomega.Succeed())

		defer func() {
			g.Expect(os.Remove(tmpfile.Name())).To(gomega.Succeed()) // clean up
		}()

		_, err = tmpfile.Write(content)
		g.Expect(err).ShouldNot(gomega.HaveOccurred())

		defer func() {
			g.Expect(tmpfile.Close()).Should(gomega.Succeed())
		}()

		// proper test phase

		conf, err := LoadDispatchingConfig(tmpfile.Name())
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(conf).To(gomega.Equal(Dispatching{Config: nil}))
	})
}

func TestDispatching_Validate(t *testing.T) {
	type fields struct {
		Config []LogsScrapingConfig
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "proper struct should pass validation",
			fields: fields{Config: []LogsScrapingConfig{
				{ChannelName: "#channel1"},
				{ChannelName: "#channel2"},
				{ChannelName: "#channel3"},
			}},
			wantErr: false,
		},
		{
			name: "struct with channelName that do not start with '#' should not pass validation",
			fields: fields{Config: []LogsScrapingConfig{
				{ChannelName: "#channel1"},
				{ChannelName: "channel2"},
				{ChannelName: "#channel3"},
			}},
			wantErr: true,
		},
		{
			name:    "no error on empty config slice",
			fields:  fields{Config: []LogsScrapingConfig{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dispatching{
				Config: tt.fields.Config,
			}
			if err := d.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDispatching_GetConfigByNameWithFallback(t *testing.T) {
	type fields struct {
		Config []LogsScrapingConfig
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    LogsScrapingConfig
		wantErr bool
	}{
		{
			name: "return default config if there's not config that fits criteria",
			fields: fields{Config: []LogsScrapingConfig{
				{TestCases: []string{"test-name1"}, ChannelName: "test-name1"},
				{TestCases: []string{"test-name2"}, ChannelName: "test-name2"},
				{TestCases: []string{"test-name3"}, ChannelName: "something"},
				{TestCases: []string{"default"}, ChannelName: "something-default"},
			}},
			args:    args{name: "test-name4"},
			want:    LogsScrapingConfig{TestCases: []string{"default"}, ChannelName: "something-default"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dispatching{
				Config: tt.fields.Config,
			}
			got, err := d.GetConfigByNameWithFallback(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigByNameWithFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfigByNameWithFallback() got = %v, want %v", got, tt.want)
			}
		})
	}
}
