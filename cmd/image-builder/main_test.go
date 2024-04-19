package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/kyma-project/test-infra/pkg/sets"
	"github.com/kyma-project/test-infra/pkg/sign"
	"github.com/kyma-project/test-infra/pkg/tags"
)

func Test_gatherDestinations(t *testing.T) {
	tc := []struct {
		name     string
		repos    []string
		tags     []tags.Tag
		expected []string
	}{
		{
			name:  "subdirectory/test-image",
			repos: []string{"dev.kyma.io", "dev2.kyma.io"},
			tags: []tags.Tag{
				{Name: "TestName", Value: "20222002-abcd1234"},
				{Name: "", Value: "latest"},
				{Name: "cookie", Value: "cookie"},
			},
			expected: []string{
				"dev.kyma.io/subdirectory/test-image:20222002-abcd1234",
				"dev2.kyma.io/subdirectory/test-image:20222002-abcd1234",
				"dev.kyma.io/subdirectory/test-image:latest",
				"dev2.kyma.io/subdirectory/test-image:latest",
				"dev.kyma.io/subdirectory/test-image:cookie",
				"dev2.kyma.io/subdirectory/test-image:cookie",
			},
		},
		{
			name:  "test-no-directory",
			repos: []string{"kyma.dev"},
			tags: []tags.Tag{
				{Name: "", Value: "latest"},
			},
			expected: []string{"kyma.dev/test-no-directory:latest"},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			got := gatherDestinations(c.repos, c.name, c.tags)
			if len(c.expected) != len(got) {
				t.Errorf("result length mismatch. wanted %v, got %v", len(c.expected), len(got))
			}
			if !reflect.DeepEqual(c.expected, got) {
				t.Errorf("%s != %s", got, c.expected)
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
			name:     "key -> key=val",
			expected: "key=val",
			test:     "key",
		},
		{
			name:     "_KEY -> _KEY=val",
			expected: "_KEY=val",
			test:     "_KEY",
		},
		{
			name:     "_key -> _key=val",
			expected: "_key=val",
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
			name:      "parsed config",
			expectErr: false,
			opts: options{
				context:    "directory/",
				name:       "test-image",
				dockerfile: "dockerfile",
				configPath: "config.yaml",
			},
		},
		{
			name:      "context missing",
			expectErr: true,
			opts: options{
				name:       "test-image",
				dockerfile: "dockerfile",
			},
		},
		{
			name:      "name missing",
			expectErr: true,
			opts: options{
				context:    "directory/",
				dockerfile: "dockerfile",
			},
		},
		{
			name:      "dockerfile missing",
			expectErr: true,
			opts: options{
				context: "directory/",
				name:    "test-image",
			},
		},
		{
			name:      "Empty configPath",
			expectErr: true,
			opts: options{
				context:    "directory/",
				name:       "test-image",
				dockerfile: "dockerfile",
			},
		},
		{
			name:      "signOnly without imagesToSign",
			expectErr: true,
			opts: options{
				context:      "directory/",
				name:         "test-image",
				dockerfile:   "dockerfile",
				configPath:   "config.yaml",
				signOnly:     true,
				imagesToSign: []string{},
			},
		},
		{
			name:      "imagesToSign without signOnly",
			expectErr: true,
			opts: options{
				context:      "directory/",
				name:         "test-image",
				dockerfile:   "dockerfile",
				configPath:   "config.yaml",
				signOnly:     false,
				imagesToSign: []string{"image1"},
			},
		},
		{
			name:      "envFile with buildInADO",
			expectErr: true,
			opts: options{
				context:    "directory/",
				name:       "test-image",
				dockerfile: "dockerfile",
				configPath: "config.yaml",
				envFile:    "envfile",
				buildInADO: true,
			},
		},
		{
			name:      "variant with buildInADO",
			expectErr: true,
			opts: options{
				context:    "directory/",
				name:       "test-image",
				dockerfile: "dockerfile",
				configPath: "config.yaml",
				variant:    "variant",
				buildInADO: true,
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
				context:    ".",
				configPath: "/config/image-builder-config.yaml",
				dockerfile: "dockerfile",
				logDir:     "/logs/artifacts",
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
				name: "test-image",
				tags: []tags.Tag{
					{Name: "latest", Value: "latest"},
					{Name: "cookie", Value: "cookie"},
				},
				context:    "prow/build",
				configPath: "config.yaml",
				dockerfile: "dockerfile",
				logDir:     "prow/logs",
				orgRepo:    "kyma-project/test-infra",
				silent:     true,
			},
			args: []string{
				"--config=config.yaml",
				"--dockerfile=dockerfile",
				"--repo=kyma-project/test-infra",
				"--name=test-image",
				"--tag=latest",
				"--tag=cookie=cookie",
				"--context=prow/build",
				"--log-dir=prow/logs",
				"--silent",
			},
		},
		{
			name: "export tag, pass",
			expectedOpts: options{
				context:    ".",
				configPath: "/config/image-builder-config.yaml",
				dockerfile: "dockerfile",
				logDir:     "/logs/artifacts",
				exportTags: true,
			},
			args: []string{
				"--export-tags",
			},
		},
		{
			name: "build args, pass",
			expectedOpts: options{
				context:    ".",
				configPath: "/config/image-builder-config.yaml",
				dockerfile: "dockerfile",
				logDir:     "/logs/artifacts",
				buildArgs: sets.Tags{
					tags.Tag{Name: "BIN", Value: "test"},
					tags.Tag{Name: "BIN2", Value: "test2"},
				},
			},
			args: []string{
				"--build-arg=BIN=test",
				"--build-arg=BIN2=test2",
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

func Test_gatTags(t *testing.T) {
	tc := []struct {
		name           string
		pr             string
		sha            string
		tagTemplate    tags.Tag
		env            map[string]string
		additionalTags []tags.Tag
		expectErr      bool
		expectResult   []tags.Tag
	}{
		{
			name:         "pr variable is present",
			pr:           "1234",
			expectResult: []tags.Tag{{Name: "default_tag", Value: "PR-1234"}},
		},
		{
			name:      "sha is empty",
			expectErr: true,
		},
		{
			name:        "bad tagTemplate",
			expectErr:   true,
			sha:         "abcd1234",
			tagTemplate: tags.Tag{Name: "TagTemplate", Value: `v{{ .ASD }}`},
		},
		{
			name:        "custom template, additional fields, env variable",
			expectErr:   false,
			sha:         "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			tagTemplate: tags.Tag{Name: "TagTemplate", Value: `{{ .ShortSHA }}`},
			env:         map[string]string{"CUSTOM_ENV": "customEnvValue"},
			additionalTags: []tags.Tag{
				{Name: "latest", Value: "latest"},
				{Name: "Test", Value: "cookie"},
				{Name: "AnotherTest", Value: `{{ .CommitSHA }}`},
				{Name: "TestEnv", Value: `{{ .Env "CUSTOM_ENV" }}`},
			},
			expectResult: []tags.Tag{
				{Name: "latest", Value: "latest"},
				{Name: "Test", Value: "cookie"},
				{Name: "AnotherTest", Value: `da39a3ee5e6b4b0d3255bfef95601890afd80709`},
				{Name: "TestEnv", Value: "customEnvValue"},
				{Name: "TagTemplate", Value: "da39a3ee"},
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			got, err := getTags(c.pr, c.sha, append(c.additionalTags, c.tagTemplate))
			if err != nil && !c.expectErr {
				t.Errorf("got error but didn't want to: %s", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't get error but wanted to")
			}
			if !reflect.DeepEqual(c.expectResult, got) {
				t.Errorf("%v != %v", got, c.expectResult)
			}
		})
	}
}

func Test_loadEnv(t *testing.T) {
	// static value that should not be overridden
	t.Setenv("key3", "static-value")
	vfs := fstest.MapFS{
		".env": &fstest.MapFile{Data: []byte("KEY=VAL\nkey2=val2\nkey3=val3\nkey4=val4=asf"), Mode: 0666},
	}
	expected := map[string]string{
		"KEY":  "VAL",
		"key2": "val2",
		"key3": "static-value",
		"key4": "val4=asf",
	}
	_, err := loadEnv(vfs, ".env")
	if err != nil {
		t.Errorf("%v", err)
	}

	for k, v := range expected {
		got := os.Getenv(k)
		if got != v {
			t.Errorf("%v != %v", got, v)
		}
		os.Unsetenv(k)
	}
}

func Test_getSignersForOrgRepo(t *testing.T) {
	tc := []struct {
		name          string
		expectErr     bool
		expectSigners int
		orgRepo       string
		jobType       string
		ci            bool
	}{
		{
			name:          "1 notary signer org/repo, pass",
			expectErr:     false,
			expectSigners: 1,
			orgRepo:       "org/repo",
		},
		{
			name:          "2 notary signer org/repo2, pass",
			expectErr:     false,
			expectSigners: 2,
			orgRepo:       "org/repo2",
		},
		{
			name:          "only global signer, one notary signer, pass",
			expectErr:     false,
			expectSigners: 1,
			orgRepo:       "org/repo-empty",
		},
		{
			name:          "1 global signer for presubmit job",
			expectErr:     false,
			expectSigners: 1,
			orgRepo:       "ci-org/ci-repo",
			jobType:       "presubmit",
			ci:            true,
		},
		{
			name:          "2 signers for postsubmit job",
			expectErr:     false,
			expectSigners: 2,
			orgRepo:       "ci-org/ci-repo",
			jobType:       "postsubmit",
			ci:            true,
		},
		{
			name:          "1 signer in non-CI environment",
			expectErr:     false,
			expectSigners: 1,
			orgRepo:       "ci-org/ci-repo",
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("JOB_TYPE", c.jobType)
			o := &options{isCI: c.ci, Config: Config{SignConfig: SignConfig{
				EnabledSigners: map[string][]string{
					"*":              {"test-notary"},
					"org/repo":       {"test-notary"},
					"org/repo2":      {"test-notary2"},
					"ci-org/ci-repo": {"ci-notary"},
				},
				Signers: []sign.SignerConfig{
					{
						Name:   "test-notary",
						Type:   sign.TypeNotaryBackend,
						Config: sign.NotaryConfig{},
					},
					{
						Name:   "test-notary2",
						Type:   sign.TypeNotaryBackend,
						Config: sign.NotaryConfig{},
					},
					{
						Name:    "ci-notary",
						Type:    sign.TypeNotaryBackend,
						Config:  sign.NotaryConfig{},
						JobType: []string{"postsubmit"},
					},
				},
			}}}
			got, err := getSignersForOrgRepo(o, c.orgRepo)
			if err != nil && !c.expectErr {
				t.Errorf("got error but didn't want to %v", err)
			}
			if len(got) != c.expectSigners {
				t.Errorf("wrong number of requested signers %v != %v", len(got), c.expectSigners)
			}
		})
	}
}

func Test_addTagsToEnv(t *testing.T) {
	tc := []struct {
		name         string
		envs         map[string]string
		tags         []tags.Tag
		expectedEnvs map[string]string
	}{
		{
			name: "multiple envs and tags",
			envs: map[string]string{"KEY_1": "VAL1", "KEY_2": "VAL2"},
			tags: []tags.Tag{
				{Name: "latest", Value: "latest"},
				{Name: "ShortSHA", Value: "dca515151"},
				{Name: "test", Value: "test-tag"},
			},
			expectedEnvs: map[string]string{"KEY_1": "VAL1", "KEY_2": "VAL2", "TAG_latest": "latest", "TAG_ShortSHA": "dca515151", "TAG_test": "test-tag"},
		},
		{
			name:         "no tags",
			envs:         map[string]string{"KEY": "VAL"},
			tags:         []tags.Tag{},
			expectedEnvs: map[string]string{"KEY": "VAL"},
		},
		{
			name: "no envs",
			envs: map[string]string{},
			tags: []tags.Tag{
				{Name: "Test", Value: "latest"},
			},
			expectedEnvs: map[string]string{"TAG_Test": "latest"},
		},
	}

	for _, c := range tc {
		actualEnv := addTagsToEnv(c.tags, c.envs)

		for k, v := range c.expectedEnvs {
			if actualEnv[k] != v {
				t.Errorf("%v != %v", actualEnv[k], v)
			}
		}
	}
}

func Test_appendMissing(t *testing.T) {
	tc := []struct {
		name         string
		existing     map[string]string
		newTags      []tags.Tag
		expectedEnvs map[string]string
	}{
		{
			name: "multiple source and targets",
			existing: map[string]string{
				"KEY_1": "VAL1",
				"KEY_2": "VAL2",
				"KEY_3": "VAL3",
			},
			newTags: []tags.Tag{
				{Name: "KEY_3", Value: "VAL5"},
				{Name: "KEY_4", Value: "VAL4"},
			},
			expectedEnvs: map[string]string{
				"KEY_1": "VAL1",
				"KEY_2": "VAL2",
				"KEY_3": "VAL3",
				"KEY_4": "VAL4",
			},
		},
	}

	for _, c := range tc {
		appendMissing(&c.existing, c.newTags)

		for k, v := range c.expectedEnvs {
			if c.existing[k] != v {
				t.Errorf("%v != %v", c.existing[k], v)
			}
		}
	}
}

func Test_parseTagsFromEnv(t *testing.T) {
	tc := []struct {
		name      string
		options   options
		env       map[string]string
		tags      []tags.Tag
		expectErr bool
	}{
		{
			name: "Prow based tags parse",
			options: options{
				isCI:     true,
				ciSystem: Prow,
			},
			env: map[string]string{
				"JOB_TYPE":      "presubmit",
				"PULL_NUMBER":   "5",
				"PULL_BASE_SHA": "testShaOfCOmmit",
			},
			tags: []tags.Tag{{Name: "default_tag", Value: "PR-5"}},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			// Prepare env vars
			for key, value := range c.env {
				t.Setenv(key, value)
			}
		})
	}
}

type mockSigner struct {
	signFunc func([]string) error
}

func (m *mockSigner) Sign(images []string) error {
	return m.signFunc(images)
}

// TODO: add tests for functions related to execution in ado.
// 		Test copied from pkg/azuredevops/pipelines/pipelines_test.go, rewrite to run it here.
// Describe("Run", func() {
// 	var (
// 		templateParams  map[string]string
// 		runPipelineArgs adoPipelines.RunPipelineArgs
// 	)
//
// 	BeforeEach(func() {
// 		templateParams = map[string]string{"param1": "value1", "param2": "value2"}
// 		runPipelineArgs = adoPipelines.RunPipelineArgs{
// 			Project:    &adoConfig.ADOProjectName,
// 			PipelineId: &adoConfig.ADOPipelineID,
// 			RunParameters: &adoPipelines.RunPipelineParameters{
// 				PreviewRun:         ptr.To(false),
// 				TemplateParameters: &templateParams,
// 			},
// 			PipelineVersion: &adoConfig.ADOPipelineVersion,
// 		}
// 	})
//
// 	It("should run the pipeline", func() {
// 		mockRun := &adoPipelines.Run{Id: ptr.To(123)}
// 		mockADOClient.On("RunPipeline", ctx, runPipelineArgs).Return(mockRun, nil)
//
// 		run, err := pipelines.Run(ctx, mockADOClient, templateParams, adoConfig)
// 		Expect(err).ToNot(HaveOccurred())
// 		Expect(run.Id).To(Equal(ptr.To(123)))
// 		mockADOClient.AssertCalled(t, "RunPipeline", ctx, runPipelineArgs)
// 		mockADOClient.AssertNumberOfCalls(t, "RunPipeline", 1)
// 		mockADOClient.AssertExpectations(GinkgoT())
// 	})
//
// 	It("should handle ADO client error", func() {
// 		mockADOClient.On("RunPipeline", ctx, runPipelineArgs).Return(nil, fmt.Errorf("ADO client error"))
//
// 		_, err := pipelines.Run(ctx, mockADOClient, templateParams, adoConfig)
// 		Expect(err).To(HaveOccurred())
// 		mockADOClient.AssertCalled(t, "RunPipeline", ctx, runPipelineArgs)
// 		mockADOClient.AssertNumberOfCalls(t, "RunPipeline", 1)
// 		mockADOClient.AssertExpectations(GinkgoT())
// 	})
//
// 	It("should run the pipeline in preview mode", func() {
// 		finalYaml := "pipeline:\n  stages:\n  - stage: Build\n    jobs:\n    - job: Build\n      steps:\n      - script: echo Hello, world!\n        displayName: 'Run a one-line script'"
// 		runPipelineArgs.RunParameters.PreviewRun = ptr.To(true)
// 		mockRun := &adoPipelines.Run{Id: ptr.To(123), FinalYaml: &finalYaml}
// 		mockADOClient.On("RunPipeline", ctx, runPipelineArgs).Return(mockRun, nil)
//
// 		run, err := pipelines.Run(ctx, mockADOClient, templateParams, adoConfig, pipelines.PipelinePreviewRun)
// 		Expect(err).ToNot(HaveOccurred())
// 		Expect(run.Id).To(Equal(ptr.To(123)))
// 		Expect(run.FinalYaml).To(Equal(&finalYaml))
// 		mockADOClient.AssertCalled(t, "RunPipeline", ctx, runPipelineArgs)
// 		mockADOClient.AssertNumberOfCalls(t, "RunPipeline", 1)
// 		mockADOClient.AssertExpectations(GinkgoT())
// 	})
// })
