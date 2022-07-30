package config

import (
	"testing"
)

func Test_validateTag(t *testing.T) {
	tc := []struct {
		name      string
		expectErr bool
		cfg       *CloudBuild
	}{
		{
			name:      "config with $_TAG substitutions in args and images",
			expectErr: false,
			cfg: &CloudBuild{
				Steps: []Step{
					{
						Args: []string{
							"build",
							"--tag",
							"gcr.io/test-repo/image/$_TAG",
							".",
						},
					},
				},
				Images: []string{
					"gcr.io/test-repo/image/$_TAG",
					"gcr.io/test-repo/image/latest",
				},
			},
		},
		{
			name:      "config with ${_TAG} substitutions in args and images",
			expectErr: false,
			cfg: &CloudBuild{
				Steps: []Step{
					{
						Args: []string{
							"build",
							"--tag",
							"gcr.io/test-repo/image/${_TAG}",
							".",
						},
					},
				},
				Images: []string{
					"gcr.io/test-repo/image/${_TAG}",
					"gcr.io/test-repo/image/latest",
				},
			},
		},
		{
			name:      "config without substitutions in args",
			expectErr: true,
			cfg: &CloudBuild{
				Steps: []Step{
					{
						Args: []string{
							"build",
							"--tag",
							"gcr.io/test-repo/image/asd1234",
							".",
						},
					},
				},
				Images: []string{
					"gcr.io/test-repo/image/${_TAG}",
					"gcr.io/test-repo/image/latest",
				},
			},
		},
		{
			name:      "config without substitutions in images",
			expectErr: true,
			cfg: &CloudBuild{
				Steps: []Step{
					{
						Args: []string{
							"build",
							"--tag",
							"gcr.io/test-repo/image/$_TAG",
							".",
						},
					},
				},
				Images: []string{
					"gcr.io/test-repo/image/latest",
				},
			},
		},
		{
			name:      "config without substitutions in args and images",
			expectErr: true,
			cfg: &CloudBuild{
				Steps: []Step{
					{
						Args: []string{
							"build",
							"--tag",
							"gcr.io/test-repo/image/asd1234",
							".",
						},
					},
				},
				Images: []string{
					"gcr.io/test-repo/image/latest",
				},
			},
		},
		{
			name:      "config with $_TAG substitutions in multiple steps",
			expectErr: false,
			cfg: &CloudBuild{
				Steps: []Step{
					{
						Args: []string{
							"-c",
							"echo 123456",
						},
					},
					{
						Args: []string{
							"build",
							"--tag",
							"gcr.io/test-repo/image/$_TAG",
							".",
						},
					},
				},
				Images: []string{
					"gcr.io/test-repo/image/$_TAG",
				},
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err := validateTag(c.cfg)
			if err != nil && !c.expectErr {
				t.Errorf("validateTag caught error, but didn't want to: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("validateTag didn't catch error, but wanted to.")
			}
		})
	}
}

func Test_validateRepository(t *testing.T) {
	tc := []struct {
		name      string
		expectErr bool
		cfg       *CloudBuild
	}{
		{
			name:      "_REPOSITORY substitution is present in config",
			expectErr: false,
			cfg:       &CloudBuild{Substitutions: map[string]string{"_REPOSITORY": "repo.com"}},
		},
		{
			name:      "substitutions map is nil",
			expectErr: true,
			cfg:       &CloudBuild{Substitutions: nil},
		},
		{
			name:      "substitutions map is initialized, but empty",
			expectErr: true,
			cfg:       &CloudBuild{Substitutions: map[string]string{}},
		},
		{
			name:      "substitutions map doesn't have _REPOSITORY substitution",
			expectErr: true,
			cfg:       &CloudBuild{Substitutions: map[string]string{"_ASD": "123"}},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err := validateRepository(c.cfg)
			if err != nil && !c.expectErr {
				t.Errorf("validateRepository caught error, but didn't want to: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("validateRepository didn't catch error, but wanted to.")
			}
		})
	}
}

func Test_checkVariants(t *testing.T) {
	tc := []struct {
		name      string
		c         *CloudBuild
		vs        Variants
		expectErr bool
	}{
		{
			name:      "variants.yaml provided, variants substitution present",
			expectErr: false,
			c:         &CloudBuild{Images: []string{"$_REPOSITORY/name:$_TAG-$_VARIANT"}},
			vs:        Variants{"main": map[string]string{"KUBECTL_VERSION": "1.24.4"}, "1.23": map[string]string{"KUBECTL_VERSION": "1.23.9"}},
		},
		{
			name:      "variants.yaml provided, variants substitution present",
			expectErr: false,
			c:         &CloudBuild{Images: []string{"$_REPOSITORY/name:$_TAG-${_VARIANT}"}},
			vs:        Variants{"main": map[string]string{"KUBECTL_VERSION": "1.24.4"}, "1.23": map[string]string{"KUBECTL_VERSION": "1.23.9"}},
		},
		{
			name:      "variants.yaml provided, variants substitution missing",
			expectErr: true,
			c:         &CloudBuild{Images: []string{"$_REPOSITORY/name:$_TAG"}},
			vs:        Variants{"main": map[string]string{"KUBECTL_VERSION": "1.24.4"}, "1.23": map[string]string{"KUBECTL_VERSION": "1.23.9"}},
		},
		{
			name:      "variants.yaml missing, variants substitution missing",
			expectErr: false,
			c:         &CloudBuild{Images: []string{"$_REPOSITORY/name:$_TAG"}},
			vs:        nil,
		},
		{
			name:      "variants.yaml missing, variants substitution provided",
			expectErr: true,
			c:         &CloudBuild{Images: []string{"$_REPOSITORY/name:$_TAG-$_VARIANT"}},
			vs:        nil,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err := validateVariants(c.c, c.vs)
			if err != nil && !c.expectErr {
				t.Errorf("validateVariants caught error, but didn't want to: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("validateVariants didn't catch error, but wanted to.")
			}
		})
	}
}
