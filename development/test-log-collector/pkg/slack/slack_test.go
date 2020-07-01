package slack

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestCLient_groupMessagesByChannelID(t *testing.T) {
	tests := []struct {
		name string
		args []Message
		want map[string][]Message
	}{
		{
			name: "simple",
			args: []Message{
				{ChannelID: "test-name-1", Data: "data-1"},
				{ChannelID: "test-name-1", Data: "data-2"},
				{ChannelID: "test-name-1", Data: "data-3"},
				{ChannelID: "test-name-2", Data: "data-4"},
			},
			want: map[string][]Message{
				"test-name-1": {
					{ChannelID: "test-name-1", Data: "data-1"},
					{ChannelID: "test-name-1", Data: "data-2"},
					{ChannelID: "test-name-1", Data: "data-3"},
				},
				"test-name-2": {
					{ChannelID: "test-name-2", Data: "data-4"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := CLient{}
			got := s.groupMessagesByChannelID(tt.args)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
