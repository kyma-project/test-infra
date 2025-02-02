package securityconfig

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadSecurityConfig(t *testing.T) {
	tc := []struct {
		Name           string
		WantErr        bool
		ExpectedConfig *SecurityConfig
		FileContent    string
	}{
		{
			Name:    "valid full config, pass",
			WantErr: false,
			ExpectedConfig: &SecurityConfig{
				ModuleName: "test-infra",
				Images:     []string{"europe-docker.pkg.dev/kyma-project/prod/buildpack-go:v20230717-e09b0fee"},
				Mend: Mend{
					Language:    "golang-mod",
					SubProjects: true,
					Exclude:     []string{"**/examples/**"},
				},
			},
			FileContent: `module-name: test-infra
protecode:
  - europe-docker.pkg.dev/kyma-project/prod/buildpack-go:v20230717-e09b0fee
whitesource:
  language: golang-mod
  subprojects: true
  exclude:
    - "**/examples/**"`,
		},
		{
			Name:           "empty config file, fail",
			WantErr:        true,
			ExpectedConfig: nil,
			FileContent:    ``,
		},
		{
			Name:    "Valid config with mend, pass",
			WantErr: false,
			ExpectedConfig: &SecurityConfig{
				ModuleName: "test-infra",
				Images:     []string{"europe-docker.pkg.dev/kyma-project/prod/buildpack-go:v20230717-e09b0fee"},
				Mend: Mend{
					Language:    "golang-mod",
					SubProjects: true,
					Exclude:     []string{"**/examples/**"},
				},
			},
			FileContent: `module-name: test-infra
protecode:
  - europe-docker.pkg.dev/kyma-project/prod/buildpack-go:v20230717-e09b0fee
mend:
  language: golang-mod
  subprojects: true
  exclude:
    - "**/examples/**"`,
		},
	}

	for _, c := range tc {
		t.Run(c.Name, func(t *testing.T) {
			config, err := ParseSecurityConfig(strings.NewReader(c.FileContent))
			if err != nil && !c.WantErr {
				t.Errorf("unexpected error occurred: %s", err)
			}

			if !reflect.DeepEqual(config, c.ExpectedConfig) {
				t.Errorf("%v != %v", config, c.ExpectedConfig)
			}
		})
	}
}
