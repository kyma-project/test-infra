package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineMatching(t *testing.T) {
	tests := []struct {
		name,
		line string
		skipComments bool
		expected     bool
	}{
		{
			name:         "Don't match random line",
			line:         "unimportant",
			skipComments: false,
			expected:     false,
		},
		{
			name:         "Don't match \"image:\" line",
			line:         "image:",
			skipComments: false,
			expected:     false,
		},
		{
			name: "Don't match modern include line",
			// "{{" breaks syntax colouring in Visual Studio Code, The comment at the end prevents that
			line:         "image: {{include \"imageurl\" (dict blahblah)", // }}"
			skipComments: false,
			expected:     false,
		},
		{
			name: "Don't match quoted modern include line",
			// "{{" breaks syntax colouring in Visual Studio Code, The comment at the end prevents that
			line:         "image: \"{{ include \"imageurl\" (dict blahblah)", // }}"
			skipComments: false,
			expected:     false,
		},
		{
			name: "Don't  match \" commented modern include line with SkipComments set to false",
			// "{{" breaks syntax colouring in Visual Studio Code, The comment at the end prevents that
			line:         "# image: {{ include \"imageurl\" (dict blahblah)", // }}"
			skipComments: false,
			expected:     false,
		},
		{
			name: "Don't match \" commented modern include line with SkipComments set to true",
			// "{{" breaks syntax colouring in Visual Studio Code, The comment at the end prevents that
			line:         "# image: {{ include \"imageurl\" (dict blahblah)", // }}"
			skipComments: true,
			expected:     false,
		},
		{
			name: "Don't match modern include without space before include",
			// "{{" breaks syntax colouring in Visual Studio Code, The comment at the end prevents that
			line:         "# image: {{include \"imageurl\" (dict blahblah)", // }}"
			skipComments: true,
			expected:     false,
		},
		{
			name:         "Match old include line",
			line:         "image: busybox",
			skipComments: false,
			expected:     true,
		},
		{
			name:         "Match quoted old image line",
			line:         "image: \"eu.gcr.io/kyma-project/external/busybox",
			skipComments: false,
			expected:     true,
		},
		{
			name:         "Match commented old image line with SkipComments set to false",
			line:         "# image: eu.gcr.io/kyma-project/external/busybox",
			skipComments: false,
			expected:     true,
		},
		{
			name:         "Don't match commented old image line with SkipComments set to true",
			line:         "# image: eu.gcr.io/kyma-project/external/busybox",
			skipComments: true,
			expected:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := oldImageFormat(test.line, test.skipComments)
			assert.New(t).Equal(test.expected, actual)
		})
	}
}
