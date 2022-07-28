package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

func TestFlags(t *testing.T) {
	testcases := []struct {
		name         string
		expectedOpts options
		expectedErr  bool
		args         []string
	}{
		{
			name: "unknown flag, fail",
			expectedOpts: options{
				cloudbuild:   "cloudbuild.yaml",
				variantsFile: "",
				buildDir:     ".",
				logDir:       "/logs/artifacts",
			},
			expectedErr: true,
			args: []string{
				"--unknown-flag=asdasd",
			},
		},
		{
			name:        "parsed config, pass",
			expectedErr: false,
			expectedOpts: options{
				cloudbuild:   "cloud.yaml",
				configPath:   "config.yaml",
				variantsFile: "var.yaml",
				buildDir:     "prow/build",
				logDir:       "prow/logs",
				silent:       true,
			},
			args: []string{
				"--config=config.yaml",
				"--cloudbuild-file=cloud.yaml",
				"--variants-file=var.yaml",
				"--build-dir=prow/build",
				"--log-dir=prow/logs",
				"--silent",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			o := options{}
			o.gatherOptions(fs)
			if err := fs.Parse(tc.args); err != nil && !tc.expectedErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if !reflect.DeepEqual(o, tc.expectedOpts) {
				t.Errorf("%v != %v", o, tc.expectedOpts)
			}
		})
	}
}

func Test_parseVariable(t *testing.T) {
	tc := []struct {
		name     string
		expected string
		test     string
	}{
		{
			name:     "key -> _KEY=val",
			expected: "_KEY=val",
			test:     "key",
		},
		{
			name:     "_KEY -> _KEY=val",
			expected: "_KEY=val",
			test:     "_KEY",
		},
		{
			name:     "_key -> _KEY=val",
			expected: "_KEY=val",
			test:     "_key",
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			got := parseVariable(c.test, "val")
			if got != c.expected {
				t.Errorf("%s != %s", got, c.expected)
			}
		})
	}
}

func Test_validateOptions(t *testing.T) {
	tc := []struct {
		name      string
		expectErr bool
		opts      options
	}{
		{
			name:      "project is missing",
			expectErr: true,
			opts: options{
				buildDir:   "dir/",
				cloudbuild: "cloud.yaml",
			},
		},
		{
			name:      "variantsFie is missing",
			expectErr: true,
			opts: options{
				buildDir:   "dir/",
				cloudbuild: "cloud.yaml",
				variant:    "main",
				Config:     Config{Project: "sample-project"},
			},
		},
		{
			name:      "buildDir is missing",
			expectErr: true,
			opts: options{
				cloudbuild: "cloud.yaml",
				Config:     Config{Project: "sample-project"},
			},
		},
		{
			name:      "cloudbuild is missing",
			expectErr: true,
			opts: options{
				buildDir: ".",
				Config:   Config{Project: "sample-project"},
			},
		},
		{
			name:      "options are valid",
			expectErr: false,
			opts: options{
				buildDir:     ".",
				cloudbuild:   "cloud.yaml",
				variant:      "main",
				variantsFile: "variants.yaml",
				Config:       Config{Project: "sample-project"},
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err := validateOptions(c.opts)
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't catch error, but wanted to")
			}
		})
	}
}

func Test_getImageNames(t *testing.T) {
	tc := []struct {
		name         string
		images       []string
		expectString string
	}{
		{
			name: "replaced strings without {}",
			images: []string{
				"$_REPOSITORY/image:$_TAG-$_VARIANT",
				"$_REPOSITORY/image:latest",
			},
			expectString: "kyma.dev/image:12345678-abcdef-main kyma.dev/image:latest",
		},
		{
			name: "replaced strings with {}",
			images: []string{
				"${_REPOSITORY}/image:${_TAG}-${_VARIANT}",
				"${_REPOSITORY}/image:latest",
			},
			expectString: "kyma.dev/image:12345678-abcdef-main kyma.dev/image:latest",
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			repo := "kyma.dev"
			tag := "12345678-abcdef"
			variant := "main"
			s := getImageNames(repo, tag, variant, c.images)
			if s != c.expectString {
				t.Errorf("%s != %s", s, c.expectString)
			}
		})
	}
}

func Test_getVariants(t *testing.T) {
	tc := []struct {
		name           string
		variant        string
		expectVariants variants
		expectErr      bool
		variantsFile   string
	}{
		{
			name:         "valid file, variants passed",
			expectErr:    false,
			variantsFile: "testdata/variants.yaml",
			expectVariants: variants{
				"main": map[string]string{"KUBECTL_VERSION": "1.24.4"},
				"1.23": map[string]string{"KUBECTL_VERSION": "1.23.9"},
			},
		},
		{
			name:           "variant file does not exist, pass",
			expectErr:      false,
			variantsFile:   "testdata/not_found.yaml",
			expectVariants: nil,
		},
		{
			name:         "get only single variant, pass",
			expectErr:    false,
			variantsFile: "testdata/variants.yaml",
			variant:      "main",
			expectVariants: variants{
				"main": map[string]string{"KUBECTL_VERSION": "1.24.4"},
			},
		},
		{
			name:           "variant is not present in variants file, fail",
			expectErr:      true,
			variantsFile:   "testdata/variants.yaml",
			variant:        "missing-variant",
			expectVariants: nil,
		},
		{
			name:           "malformed variants file, fail",
			expectErr:      true,
			variantsFile:   "testdata/malformed-variants.yaml",
			expectVariants: nil,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			o := options{
				variantsFile: c.variantsFile,
				variant:      c.variant,
			}
			v, err := getVariants(o)
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't catch error, but wanted to")
			}
			if !reflect.DeepEqual(v, c.expectVariants) {
				t.Errorf("%v != %v", v, c.expectVariants)
			}
		})
	}
}

func Test_getCloudbuild(t *testing.T) {
	tc := []struct {
		name             string
		expectErr        bool
		expectCloudbuild *Cloudbuild
		cloudbuildFile   string
	}{
		{
			name:           "missing cloudbuild.yaml, fail",
			cloudbuildFile: "testdata/missing-cloudbuild.yaml",
			expectErr:      true,
		},
		{
			name:           "malformed cloudbuild, fail",
			cloudbuildFile: "testdata/malformed-cloudbuild.yaml",
			expectErr:      true,
		},
		{
			name:           "valid cloudbuild, pass",
			cloudbuildFile: "testdata/test-cloudbuild.yaml",
			expectErr:      false,
			expectCloudbuild: &Cloudbuild{
				Steps: []Step{
					{
						Name: "gcr.io/cloud-builders/docker",
						Args: []string{
							"build",
							"--tag=$_REPOSITORY/test:$_TAG",
							"--tag=$_REPOSITORY/test:latest",
							".",
						},
					},
				},
				Substitutions: map[string]string{"_REPOSITORY": "kyma.dev"},
				Images:        []string{"$_REPOSITORY/test:$_TAG", "$_REPOSITORY/test:latest"},
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			cb, err := getCloudbuild(c.cloudbuildFile)
			if err != nil && !c.expectErr {
				t.Errorf("caught error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't catch error, but wanted to")
			}
			if !reflect.DeepEqual(cb, c.expectCloudbuild) {
				t.Errorf("%v != %v", cb, c.expectCloudbuild)
			}
		})
	}
}
