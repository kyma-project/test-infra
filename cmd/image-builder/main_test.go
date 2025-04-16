package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
	"github.com/kyma-project/test-infra/pkg/sets"
	"github.com/kyma-project/test-infra/pkg/sign"
	"github.com/kyma-project/test-infra/pkg/tags"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	defaultPRTag         = tags.Tag{Name: "default_tag", Value: `PR-{{ .PRNumber }}`, Validation: "^(PR-[0-9]+)$"}
	defaultCommitTag     = tags.Tag{Name: "default_tag", Value: `v{{ .Date }}-{{ .ShortSHA }}`, Validation: "^(v[0-9]{8}-[0-9a-f]{8})$"}
	expectedDefaultPRTag = func(prNumber int) tags.Tag {
		return tags.Tag{Name: "default_tag", Value: "PR-" + strconv.Itoa(prNumber), Validation: "^(PR-[0-9]+)$"}
	}
	expectedDefaultCommitTag = func(baseSHA string) tags.Tag {
		return tags.Tag{Name: "default_tag", Value: "v" + time.Now().Format("20060102") + "-" + fmt.Sprintf("%.8s", baseSHA), Validation: "^(v[0-9]{8}-[0-9a-f]{8})$"}
	}
	buildConfig = Config{
		DefaultPRTag:     defaultPRTag,
		DefaultCommitTag: defaultCommitTag,
	}
	prGitState = GitStateConfig{
		BaseCommitSHA:     "abcdef123456",
		PullRequestNumber: 5,
		isPullRequest:     true,
	}
	commitGitState = GitStateConfig{
		BaseCommitSHA: "abcdef123456",
		isPullRequest: false,
	}
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

var _ = Describe("Image Builder", func() {
	DescribeTable("Test validate options",
		func(options options, expectedError bool) {
			err := validateOptions(options)
			if !expectedError {
				Expect(err).NotTo(HaveOccured(),fmt.Sprintf("caught error, but didn't want to: %v", err))
			}
			if expectedError {
				Expect(err).To(HaveOccured(), "didn't catch error, but wanted to")
			}
		},
		Entry(
			"parsed config",
			options{
				context:     "directory/",
				name:        "test-image",
				dockerfile:  "dockerfile",
				configPath:  "config.yaml",
				buildEngine: "kaniko",
			},
			false,
		),
		Entry(
			"context missing",
			options{
				name:        "test-image",
				dockerfile:  "dockerfile",
				buildEngine: "kaniko",
			},
			true,
		),
		Entry(
			"name missing",
			options{
				context:     "directory/",
				dockerfile:  "dockerfile",
				buildEngine: "kaniko",
			},
			true,
		),
		Entry(
			"dockerfile missing",
			options{
				context:     "directory/",
				name:        "test-image",
				buildEngine: "kaniko",
			},
			true,
		),
		Entry(
			"Empty configPath",
			options{
				context:     "directory/",
				name:        "test-image",
				dockerfile:  "dockerfile",
				buildEngine: "kaniko",
			},
			true,
		),
		Entry(
			"signOnly without imagesToSign",
			options{
				context:      "directory/",
				name:         "test-image",
				dockerfile:   "dockerfile",
				configPath:   "config.yaml",
				signOnly:     true,
				imagesToSign: []string{},
				buildEngine:  "kaniko",
			},
			true,
		),
		Entry(
			"imagesToSign without signOnly",
			options{
				context:      "directory/",
				name:         "test-image",
				dockerfile:   "dockerfile",
				configPath:   "config.yaml",
				signOnly:     false,
				imagesToSign: []string{"image1"},
				buildEngine:  "kaniko",
			},
			true,
		),
		Entry(
			"envFile with buildInADO",
			options{
				context:     "directory/",
				name:        "test-image",
				dockerfile:  "dockerfile",
				configPath:  "config.yaml",
				envFile:     "envfile",
				buildInADO:  true,
				buildEngine: "kaniko",
			},
			false,
		),
		Entry(
			"variant with buildInADO",
			options{
				context:     "directory/",
				name:        "test-image",
				dockerfile:  "dockerfile",
				configPath:  "config.yaml",
				variant:     "variant",
				buildInADO:  true,
				buildEngine: "kaniko",
			},
			true,
		),
		Entry(
			"incorrect build engine",
			options{
				context:     "directory/",
				name:        "test-image",
				dockerfile:  "dockerfile",
				configPath:  "config.yaml",
				variant:     "variant",
				buildInADO:  true,
				buildEngine: "incorrect-build-engine",
			},
			true,
		),
		Entry(
			"correct build engine",
			options{
				context:     "directory/",
				name:        "test-image",
				dockerfile:  "dockerfile",
				configPath:  "config.yaml",
				buildEngine: "buildx",
			},
			false,
		),
	)

	DescribeTable("Test Flags",
		func(args []string, expectedOptions options, expectedError bool) {
			fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
			o := options{}
			o.gatherOptions(fs)
			err := fs.Parse(args)
			if err != nil && !expectedError {
				Fail(fmt.Sprintf("caught error, but didn't want to: %v", err))
			}
			if err == nil && expectedError {
				Fail("didn't catch error, but wanted to")
			}

			if !reflect.DeepEqual(o, expectedOptions) {
				Fail(fmt.Sprintf("%v != %v", o, expectedOptions))
			}
		},
		Entry("unknown flag, fail",
			[]string{
				"--unknown-flag=asdasd",
			},
			options{
				context:        ".",
				configPath:     "/config/image-builder-config.yaml",
				dockerfile:     "dockerfile",
				logDir:         "/logs/artifacts",
				tagsOutputFile: "/generated-tags.json",
				buildEngine:    "kaniko",
			},
			true,
		),
		Entry("parsed config, pass",
			[]string{
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
			options{
				name: "test-image",
				tags: []tags.Tag{
					{Name: "latest", Value: "latest"},
					{Name: "cookie", Value: "cookie"},
				},
				context:        "prow/build",
				configPath:     "config.yaml",
				dockerfile:     "dockerfile",
				logDir:         "prow/logs",
				orgRepo:        "kyma-project/test-infra",
				silent:         true,
				tagsOutputFile: "/generated-tags.json",
				buildEngine:    "kaniko",
			},
			false,
		),
		Entry("export tag, pass",
			[]string{
				"--export-tags",
			},
			options{
				context:        ".",
				configPath:     "/config/image-builder-config.yaml",
				dockerfile:     "dockerfile",
				logDir:         "/logs/artifacts",
				exportTags:     true,
				tagsOutputFile: "/generated-tags.json",
				buildEngine:    "kaniko",
			},
			false,
		),
		Entry("build args, pass",
			[]string{
				"--build-arg=BIN=test",
				"--build-arg=BIN2=test2",
			},
			options{
				context:    ".",
				configPath: "/config/image-builder-config.yaml",
				dockerfile: "dockerfile",
				logDir:     "/logs/artifacts",
				buildArgs: sets.Tags{
					tags.Tag{Name: "BIN", Value: "test"},
					tags.Tag{Name: "BIN2", Value: "test2"},
				},
				tagsOutputFile: "/generated-tags.json",
				buildEngine:    "kaniko",
			},
			false,
		),
		Entry("build engine, pass",
			[]string{
				"--build-engine=buildx",
			},
			options{
				context:        ".",
				configPath:     "/config/image-builder-config.yaml",
				dockerfile:     "dockerfile",
				logDir:         "/logs/artifacts",
				tagsOutputFile: "/generated-tags.json",
				buildEngine:    "buildx",
			},
			false,
		),
	)

	DescribeTable("Test prepareADOTemplateParameters",
		func(expectedtOptions options, want pipelines.OCIImageBuilderTemplateParams, wantErr bool) {
			got, err := prepareADOTemplateParameters(expectedtOptions)
			if (err != nil) != wantErr {
				Fail(fmt.Sprintf("caught error, but didn't want to: %v", err))
			}
			if err == nil && wantErr {
				Fail("didn't catch error, but wanted to")
			}

			if !reflect.DeepEqual(got, want) {
				Fail(fmt.Sprintf("%v != %v", got, want))
			}
		},
		Entry("Tag with parentheses",
			options{
				gitState: GitStateConfig{
					JobType: "postsubmit",
				},
				tags: sets.Tags{
					{Name: "{{ .Env \"GOLANG_VERSION\" }}-ShortSHA", Value: "{{ .Env \"GOLANG_VERSION\" }}-{{ .ShortSHA }}"},
				},
				buildEngine: "kaniko",
			},
			pipelines.OCIImageBuilderTemplateParams{
				"Context":     "",
				"Dockerfile":  "",
				"ExportTags":  "false",
				"JobType":     "postsubmit",
				"Name":        "",
				"PullBaseSHA": "",
				"RepoName":    "",
				"RepoOwner":   "",
				"Tags":        "e3sgLkVudiAiR09MQU5HX1ZFUlNJT04iIH19LVNob3J0U0hBPXt7IC5FbnYgIkdPTEFOR19WRVJTSU9OIiB9fS17eyAuU2hvcnRTSEEgfX0=",
				"BuildEngine": "kaniko",
			},
			false,
		),
		Entry("On demand job type with base commit SHA and base commit ref",
			options{
				gitState: GitStateConfig{
					JobType:       "workflow_dispatch",
					BaseCommitSHA: "abc123",
					BaseCommitRef: "main",
				},
				tags: sets.Tags{
					{Name: "{{ .Env \"GOLANG_VERSION\" }}-ShortSHA", Value: "{{ .Env \"GOLANG_VERSION\" }}-{{ .ShortSHA }}"},
				},
				buildEngine: "kaniko",
			},
			pipelines.OCIImageBuilderTemplateParams{
				"Context":     "",
				"Dockerfile":  "",
				"ExportTags":  "false",
				"JobType":     "workflow_dispatch",
				"Name":        "",
				"PullBaseSHA": "abc123",
				"BaseRef":     "main",
				"RepoName":    "",
				"RepoOwner":   "",
				"Tags":        "e3sgLkVudiAiR09MQU5HX1ZFUlNJT04iIH19LVNob3J0U0hBPXt7IC5FbnYgIkdPTEFOR19WRVJTSU9OIiB9fS17eyAuU2hvcnRTSEEgfX0=",
				"BuildEngine": "kaniko",
			},
			false,
		),
		Entry("Buildx engine",
			options{
				gitState: GitStateConfig{
					JobType: "postsubmit",
				},
				buildEngine: "buildx",
			},
			pipelines.OCIImageBuilderTemplateParams{
				"Context":     "",
				"Dockerfile":  "",
				"ExportTags":  "false",
				"JobType":     "postsubmit",
				"Name":        "",
				"PullBaseSHA": "",
				"RepoName":    "",
				"RepoOwner":   "",
				"BuildEngine": "buildx",
			},
			false,
		),
	)
})

func Test_getTags(t *testing.T) {
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
			name:        "generate default pr tag, when no pr number and commit sha provided",
			tagTemplate: defaultPRTag,
			expectErr:   true,
		},
		{
			name:        "generate default commit tag, when no pr number and commit sha provided",
			tagTemplate: defaultCommitTag,
			expectErr:   true,
		},
		{
			name:         "generate default pr tag, when pr number provided",
			pr:           "1234",
			tagTemplate:  defaultPRTag,
			expectResult: []tags.Tag{expectedDefaultPRTag(1234)},
		},
		{
			name:         "generate default commit tag, when commit sha provided",
			sha:          "1a2b3c4d5e6f78",
			tagTemplate:  defaultCommitTag,
			expectResult: []tags.Tag{expectedDefaultCommitTag("1a2b3c4d5e6f78")},
		},
		{
			name:           "generate default pr tag and additional tags",
			pr:             "1234",
			tagTemplate:    defaultPRTag,
			additionalTags: []tags.Tag{{Name: "additional_tag", Value: "additional"}},
			expectResult:   []tags.Tag{{Name: "additional_tag", Value: "additional"}, expectedDefaultPRTag(1234)},
		},
		{
			name:           "generate default commit tag and additional tags",
			sha:            "1a2b3c4d5e6f78",
			tagTemplate:    defaultCommitTag,
			additionalTags: []tags.Tag{{Name: "additional_tag", Value: "additional"}},
			expectResult:   []tags.Tag{{Name: "additional_tag", Value: "additional"}, expectedDefaultCommitTag("1a2b3c4d5e6f78")},
		},
		{
			name:      "no pr, sha and default tag provided",
			expectErr: true,
		},
		{
			name:        "bad tagTemplate",
			expectErr:   true,
			sha:         "1a2b3c4d5e6f78",
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
			zapLogger, err := zap.NewProduction()
			if err != nil {
				t.Errorf("got error but didn't want to: %s", err)
			}
			logger := zapLogger.Sugar()
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			got, err := getTags(logger, c.pr, c.sha, append(c.additionalTags, c.tagTemplate))
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
	zapLogger, err := zap.NewProduction()
	if err != nil {
		t.Errorf("got error but didn't want to: %s", err)
	}
	logger := zapLogger.Sugar()
	_, err = loadEnv(logger, vfs, ".env")
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
			mockFactory := &mockSignerFactory{}

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
						Config: mockFactory,
					},
					{
						Name:   "test-notary2",
						Type:   sign.TypeNotaryBackend,
						Config: mockFactory,
					},
					{
						Name:    "ci-notary",
						Type:    sign.TypeNotaryBackend,
						Config:  mockFactory,
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

func Test_parseTags(t *testing.T) {
	tagsFlag := sets.Tags{{Name: "base64testtag", Value: "testtag"}, {Name: "base64testtemplate", Value: "test-{{ .PRNumber }}"}}
	base64Tags := base64.StdEncoding.EncodeToString([]byte(tagsFlag.String()))
	zapLogger, err := zap.NewProduction()
	if err != nil {
		t.Errorf("got error but didn't want to: %s", err)
	}
	logger := zapLogger.Sugar()
	tc := []struct {
		name         string
		options      options
		expectedTags []tags.Tag
		expectErr    bool
	}{
		{
			name: "pares only PR default tag",
			options: options{
				gitState: prGitState,
				Config:   buildConfig,
				logger:   logger,
			},
			expectedTags: []tags.Tag{expectedDefaultPRTag(prGitState.PullRequestNumber)},
		},
		{
			name: "parse only commit default tag",
			options: options{
				gitState: commitGitState,
				Config:   buildConfig,
				logger:   logger,
			},
			expectedTags: []tags.Tag{expectedDefaultCommitTag(commitGitState.BaseCommitSHA)},
		},
		{
			name: "parse PR default and additional tags",
			options: options{
				gitState: prGitState,
				Config:   buildConfig,
				tags: sets.Tags{
					{Name: "AnotherTest", Value: `Another-{{ .PRNumber }}`},
					{Name: "Test", Value: "tag-value"},
				},
				logger: logger,
			},
			expectedTags: []tags.Tag{{Name: "AnotherTest", Value: "Another-" + strconv.Itoa(prGitState.PullRequestNumber)}, {Name: "Test", Value: "tag-value"}, expectedDefaultPRTag(prGitState.PullRequestNumber)},
		},
		{
			name: "parse commit default and additional tags",
			options: options{
				gitState: commitGitState,
				Config:   buildConfig,
				tags: sets.Tags{
					{Name: "AnotherTest", Value: `Another-{{ .CommitSHA }}`},
					{Name: "Test", Value: "tag-value"},
				},
				logger: logger,
			},
			expectedTags: []tags.Tag{{Name: "AnotherTest", Value: "Another-" + commitGitState.BaseCommitSHA}, {Name: "Test", Value: "tag-value"}, expectedDefaultCommitTag(commitGitState.BaseCommitSHA)},
		},
		{
			name: "parse bad tag template",
			options: options{
				gitState: prGitState,
				Config:   buildConfig,
				tags: sets.Tags{
					{Name: "BadTagTemplate", Value: `{{ .ASD }}`},
				},
				logger: logger,
			},
			expectErr: true,
		},
		{
			name: "parse tags from base64 encoded flag",
			options: options{
				gitState:   prGitState,
				Config:     buildConfig,
				tagsBase64: base64Tags,
				logger:     logger,
			},
			expectedTags: []tags.Tag{
				{Name: "base64testtag", Value: "testtag"},
				{Name: "base64testtemplate", Value: "test-5"},
				expectedDefaultPRTag(prGitState.PullRequestNumber)},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			logger := c.options.logger
			tags, err := parseTags(logger, c.options)
			if err != nil && !c.expectErr {
				t.Errorf("Got unexpected error: %s", err)
			}
			if err == nil && c.expectErr {
				t.Error("Expected error, but no one occured")
			}

			if !reflect.DeepEqual(tags, c.expectedTags) {
				t.Errorf("Got %v, but expected %v", tags, c.expectedTags)
			}
		})
	}
}

func Test_getDefaultTag(t *testing.T) {
	g := NewGomegaWithT(t)

	zapLogger, err := zap.NewProduction()
	if err != nil {
		t.Errorf("got error but didn't want to: %s", err)
	}
	logger := zapLogger.Sugar()
	tests := []struct {
		name    string
		options options
		want    tags.Tag
		wantErr bool
	}{
		{
			name: "Success - Pull Request",
			options: options{
				gitState: prGitState,
				Config:   buildConfig,
				logger:   logger,
			},
			want:    defaultPRTag,
			wantErr: false,
		},
		{
			name: "Success - Commit SHA",
			options: options{
				gitState: commitGitState,
				Config:   buildConfig,
				logger:   logger,
			},
			want:    defaultCommitTag,
			wantErr: false,
		},
		{
			name: "Failure - No PR number or commit SHA",
			options: options{
				gitState: GitStateConfig{},
				logger:   logger,
			},
			want:    tags.Tag{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tt.options.logger
			got, err := getDefaultTag(logger, tt.options)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(got).To(Equal(tt.want))
			}
		})
	}
}

type mockSignerFactory struct{}

func (m *mockSignerFactory) NewSigner() (sign.Signer, error) {
	return &mockSigner{}, nil
}

type mockSigner struct{}

func (m *mockSigner) Sign([]string) error {
	return nil
}

func Test_getDockerfileDirPath(t *testing.T) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		t.Errorf("got error but didn't want to: %s", err)
	}
	logger := zapLogger.Sugar()
	type args struct {
		o options
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Dockerfile in root directory",
			args: args{
				o: options{
					context:    ".",
					dockerfile: "Dockerfile",
					logger:     logger,
				},
			},
			want:    "/test-infra/cmd/image-builder",
			wantErr: false,
		},
		{
			name: "Dockerfile in root directory",
			args: args{
				o: options{
					context:    "cmd/image-builder",
					dockerfile: "Dockerfile",
					logger:     logger,
				},
			},
			want:    "/test-infra/cmd/image-builder",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tt.args.o.logger
			got, err := getDockerfileDirPath(logger, tt.args.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDockerfileDirPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if strings.HasSuffix(got, "tt.want") {
				t.Errorf("getDockerfileDirPath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_getEnvs(t *testing.T) {
// 	type args struct {
// 		o              options
// 		dockerfilePath string
// 	}
//
// 	zapLogger, err := zap.NewProduction()
// 	if err != nil {
// 		t.Errorf("got error but didn't want to: %s", err)
// 	}
// 	logger := zapLogger.Sugar()
//
// 	tests := []struct {
// 		name string
// 		args args
// 		want map[string]string
// 	}{
// 		{
// 			name: "Empty env file path",
// 			args: args{
// 				o: options{
// 					context:    ".",
// 					dockerfile: "Dockerfile",
// 					envFile:    "",
// 					logger:     logger,
// 				},
// 			},
// 			want: map[string]string{},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := getEnvs(tt.args.o, tt.args.dockerfilePath)
// 			if err != nil {
// 				t.Errorf("getEnvs() error = %v", err)
// 			}
// 			if got != nil {
// 				t.Errorf("getEnvs() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func Test_appendToTags(t *testing.T) {
	type args struct {
		target *[]tags.Tag
		source map[string]string
	}
	tests := []struct {
		name string
		args args
		want *[]tags.Tag
	}{
		{
			name: "Append tags",
			args: args{
				target: &[]tags.Tag{{Name: "key1", Value: "val1"}},
				source: map[string]string{"key2": "val2"},
			},
			want: &[]tags.Tag{{Name: "key1", Value: "val1"}, {Name: "key2", Value: "val2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapLogger, err := zap.NewProduction()
			if err != nil {
				t.Errorf("got error but didn't want to: %s", err)
			}
			logger := zapLogger.Sugar()
			appendToTags(logger, tt.args.target, tt.args.source)

			if !reflect.DeepEqual(tt.args.target, tt.want) {
				t.Errorf("appendToTags() got = %v, want %v", tt.args.target, tt.want)
			}
		})
	}
}

func Test_getParsedTagsAsJSON(t *testing.T) {
	type args struct {
		parsedTags []tags.Tag
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Empty tags",
			args: args{
				parsedTags: []tags.Tag{},
			},
			want:    "[]",
			wantErr: false,
		},
		{
			name: "Multiple tags",
			args: args{
				parsedTags: []tags.Tag{{Name: "key1", Value: "val1"}, {Name: "key2", Value: "val2"}},
			},
			want:    `[{"name":"key1","value":"val1"},{"name":"key2","value":"val2"}]`,
			wantErr: false,
		},
	}
	{
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := tagsAsJSON(tt.args.parsedTags)
				if err != nil && !tt.wantErr {
					t.Errorf("got error but didn't want to: %s", err)
				}
				if string(got) != tt.want {
					t.Errorf("tagsAsJSON() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}
