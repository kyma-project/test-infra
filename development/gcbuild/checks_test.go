package main

import (
	"testing"
)

func Test_validateTag(t *testing.T) {
	tc := []struct {
		name      string
		expectErr bool
		cfg       *Cloudbuild
	}{
		{
			name:      "config with $_TAG substitutions in args and images",
			expectErr: false,
			cfg: &Cloudbuild{
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
			cfg: &Cloudbuild{
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
			cfg: &Cloudbuild{
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
			cfg: &Cloudbuild{
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
			cfg: &Cloudbuild{
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
			cfg: &Cloudbuild{
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
		cfg       *Cloudbuild
	}{
		{
			name:      "_REPOSITORY substitution is present in config",
			expectErr: false,
			cfg:       &Cloudbuild{Substitutions: map[string]string{"_REPOSITORY": "repo.com"}},
		},
		{
			name:      "substitutions map is nil",
			expectErr: true,
			cfg:       &Cloudbuild{Substitutions: nil},
		},
		{
			name:      "substitutions map is initialized, but empty",
			expectErr: true,
			cfg:       &Cloudbuild{Substitutions: map[string]string{}},
		},
		{
			name:      "substitutions map doesn't have _REPOSITORY substitution",
			expectErr: true,
			cfg:       &Cloudbuild{Substitutions: map[string]string{"_ASD": "123"}},
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
		o         options
		c         *Cloudbuild
		expectErr bool
	}{
		{
			name:      "variants.yaml provided, variants substitution present",
			expectErr: false,
			o:         options{variantsFile: "testdata/variants.yaml"},
			c:         &Cloudbuild{Images: []string{"$_REPOSITORY/name:$_TAG-$_VARIANT"}},
		},
		{
			name:      "variants.yaml provided, variants substitution present",
			expectErr: false,
			o:         options{variantsFile: "testdata/variants.yaml"},
			c:         &Cloudbuild{Images: []string{"$_REPOSITORY/name:$_TAG-${_VARIANT}"}},
		},
		{
			name:      "variants.yaml provided, variants substitution missing",
			expectErr: true,
			o:         options{variantsFile: "testdata/variants.yaml"},
			c:         &Cloudbuild{Images: []string{"$_REPOSITORY/name:$_TAG"}},
		},
		{
			name:      "variants.yaml missing, variants substitution missing",
			expectErr: false,
			o:         options{},
			c:         &Cloudbuild{Images: []string{"$_REPOSITORY/name:$_TAG"}},
		},
		{
			name:      "variants.yaml missing, variants substitution provided",
			expectErr: true,
			o:         options{},
			c:         &Cloudbuild{Images: []string{"$_REPOSITORY/name:$_TAG-$_VARIANT"}},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err := validateVariants(c.o, c.c)
			if err != nil && !c.expectErr {
				t.Errorf("validateVariants caught error, but didn't want to: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("validateVariants didn't catch error, but wanted to.")
			}
		})
	}
}
