package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/kyma-project/test-infra/pkg/azuredevops/pipelines"
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
			expectErr: false,
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

func Test_parseTags(t *testing.T) {
	tc := []struct {
		name      string
		options   options
		tags      []tags.Tag
		expectErr bool
	}{
		{
			name: "PR tag parse",
			options: options{
				gitState: GitStateConfig{
					BaseCommitSHA:     "some-sha",
					PullRequestNumber: 5,
					isPullRequest:     true,
				},
				tags: sets.Tags{
					{Name: "AnotherTest", Value: `{{ .CommitSHA }}`},
				},
			},
			tags: []tags.Tag{{Name: "default_tag", Value: "PR-5"}},
		},
		{
			name: "Tags from commit sha",
			options: options{
				gitState: GitStateConfig{
					BaseCommitSHA: "some-sha",
				},
				Config: Config{
					TagTemplate: tags.Tag{Name: "AnotherTest", Value: `{{ .CommitSHA }}`},
				},
			},
			tags: []tags.Tag{{Name: "AnotherTest", Value: "some-sha"}},
		},
		{
			name: "empty commit sha",
			options: options{
				gitState: GitStateConfig{},
			},
			expectErr: true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			tags, err := parseTags(c.options)
			if err != nil && !c.expectErr {
				t.Errorf("Got unexpected error: %s", err)
			}
			if err == nil && c.expectErr {
				t.Error("Expected error, but no one occured")
			}

			if !reflect.DeepEqual(tags, c.tags) {
				t.Errorf("Got %v, but expected %v", tags, c.tags)
			}
		})
	}
}

func Test_prepareADOTemplateParameters(t *testing.T) {
	tests := []struct {
		name    string
		options options
		want    pipelines.OCIImageBuilderTemplateParams
		wantErr bool
	}{
		{
			name: "Tag with parentheses",
			options: options{
				gitState: GitStateConfig{
					JobType: "postsubmit",
				},
				tags: sets.Tags{
					{Name: "{{ .Env \"VERSION\" }}-ShortSHA", Value: "{{ .Env \"VERSION\" }}-{{ .ShortSHA }}"},
				},
			},
			want: pipelines.OCIImageBuilderTemplateParams{
				"Context":               "",
				"Dockerfile":            "",
				"ExportTags":            "false",
				"JobType":               "postsubmit",
				"Name":                  "",
				"PullBaseSHA":           "",
				"RepoName":              "",
				"RepoOwner":             "",
				"Tags":                  "e3sgLkVudiAiVkVSU0lPTiIgfX0te3sgLlNob3J0U0hBIH19",
				"UseKanikoConfigFromPR": "false",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareADOTemplateParameters(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepareADOTemplateParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareADOTemplateParameters() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractImagesFromADOLogs(t *testing.T) {
	tc := []struct {
		name           string
		expectedImages []string
		logs           string
	}{
		{
			name:           "sign image task log",
			expectedImages: []string{"europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10854", "europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10852"},
			logs: `2024-05-28T09:49:07.8176591Z ==============================================================================
			2024-05-28T09:49:07.8176701Z Task         : Docker
			2024-05-28T09:49:07.8176776Z Description  : Build or push Docker images, login or logout, start or stop containers, or run a Docker command
			2024-05-28T09:49:07.8176902Z Version      : 2.240.2
			2024-05-28T09:49:07.8176962Z Author       : Microsoft Corporation
			2024-05-28T09:49:07.8177044Z Help         : https://aka.ms/azpipes-docker-tsg
			2024-05-28T09:49:07.8177121Z ==============================================================================
			2024-05-28T09:49:08.2220004Z [command]/usr/bin/docker run --env REPO_NAME=test-infra --env REPO_OWNER=kyma-project --env CI=true --env JOB_TYPE=presubmit --mount type=bind,src=/agent/_work/1/s/kaniko-build-config.yaml,dst=/kaniko-build-config.yaml --mount type=bind,src=/agent/_work/1/s/signify-prod-secret.yaml,dst=/secret-prod/secret.yaml europe-docker.pkg.dev/kyma-project/prod/image-builder:v20240515-f756e622 --sign-only --name=image-builder --context=. --dockerfile=cmd/image-builder/images/kaniko/Dockerfile --images-to-sign=europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10854 --images-to-sign=europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10852 --config=/kaniko-build-config.yaml
			2024-05-28T09:49:08.4547604Z sign images using services signify-prod
			2024-05-28T09:49:08.4548507Z signer signify-prod ignored, because is not enabled for a CI job of type: presubmit
			2024-05-28T09:49:08.4549247Z Start signing images europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10854
			2024-05-28T09:49:08.5907215Z ##[section]Finishing: sign_images`,
		},

		{
			name:           "prepare args and sign tasks log",
			expectedImages: []string{"europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10696"},
			logs: `2024-05-28T07:36:31.8953681Z ##[section]Starting: prepare_build_and_sign_args
			2024-05-28T07:36:31.8958057Z ==============================================================================
			2024-05-28T07:36:31.8958168Z Task         : Python script
			2024-05-28T07:36:31.8958230Z Description  : Run a Python file or inline script
			2024-05-28T07:36:31.8958324Z Version      : 0.237.1
			2024-05-28T07:36:31.8958385Z Author       : Microsoft Corporation
			2024-05-28T07:36:31.8958459Z Help         : https://docs.microsoft.com/azure/devops/pipelines/tasks/utility/python-script
			2024-05-28T07:36:31.8958587Z ==============================================================================
			2024-05-28T07:36:33.6944350Z [command]/usr/bin/python /agent/_work/1/s/scripts/prepare_kaniko_and_sign_arguments.py --PreparedTagsFile /agent/_work/_temp/task_outputs/run_1716881791884.txt --ExportTags False --JobType presubmit --Context . --Dockerfile cmd/image-builder/images/kaniko/Dockerfile --ImageName image-builder --BuildArgs  --Platforms  --BuildConfigPath /agent/_work/1/s/kaniko-build-config.yaml
			2024-05-28T07:36:33.7426177Z ##[command]Read build config file:
			2024-05-28T07:36:33.7426567Z ##[group]Build config file content:
			2024-05-28T07:36:33.7430240Z ##[debug] {'tag-template': 'v{{ .Date }}-{{ .ShortSHA }}', 'registry': ['europe-docker.pkg.dev/kyma-project/prod'], 'dev-registry': ['europe-docker.pkg.dev/kyma-project/dev'], 'reproducible': False, 'log-format': 'json', 'ado-config': {'ado-organization-url': 'https://dev.azure.com/hyperspace-pipelines', 'ado-project-name': 'kyma', 'ado-pipeline-id': 14902}, 'cache': {'enabled': True, 'cache-repo': 'europe-docker.pkg.dev/kyma-project/cache/cache', 'cache-run-layers': True}, 'sign-config': {'enabled-signers': {'*': ['signify-prod']}, 'signers': [{'name': 'signify-prod', 'type': 'notary', 'job-type': ['postsubmit'], 'config': {'endpoint': 'https://signing.repositories.cloud.sap/signingsvc/sign', 'timeout': '5m', 'retry-timeout': '10s', 'secret': {'path': '/secret-prod/secret.yaml', 'type': 'signify'}}}]}}
			2024-05-28T07:36:33.7431327Z ##[endgroup]
			2024-05-28T07:36:33.7431542Z Running in presubmit mode
			2024-05-28T07:36:33.7432035Z ##[debug]Using dev registries: ['europe-docker.pkg.dev/kyma-project/dev']
			2024-05-28T07:36:33.7432334Z ##[debug]Using build context: .
			2024-05-28T07:36:33.7432779Z ##[debug]Using Dockerfile: ./cmd/image-builder/images/kaniko/Dockerfile
			2024-05-28T07:36:33.7433181Z ##[debug]Using image name: image-builder
			2024-05-28T07:36:33.7433438Z ##[command]Using prepared OCI image tags:
			2024-05-28T07:36:33.7433924Z ##[debug]Prepared tags file content: [{"name":"default_tag","value":"PR-10696"}]
			2024-05-28T07:36:33.7434608Z 
			2024-05-28T07:36:33.7435959Z ##[command]Setting job scope pipeline variable kanikoArgs with value: --cache=True --cache-run-layers=True --cache-repo=europe-docker.pkg.dev/kyma-project/cache/cache --context=dir:///workspace/. --dockerfile=/workspace/./cmd/image-builder/images/kaniko/Dockerfile --build-arg=default_tag=PR-10696 --destination=europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10696
			2024-05-28T07:36:33.7438292Z ##[command]Setting job scope pipeline variable imagesToSign with value: --images-to-sign=europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10696
			2024-05-28T07:36:33.7496968Z 
			2024-05-28T07:36:33.7549637Z ##[section]Finishing: prepare_build_and_sign_args
			2024-05-28T07:38:12.4360275Z ##[section]Starting: sign_images
2024-05-28T07:38:12.4364459Z ==============================================================================
2024-05-28T07:38:12.4364568Z Task         : Docker
2024-05-28T07:38:12.4364645Z Description  : Build or push Docker images, login or logout, start or stop containers, or run a Docker command
2024-05-28T07:38:12.4364762Z Version      : 2.240.2
2024-05-28T07:38:12.4364823Z Author       : Microsoft Corporation
2024-05-28T07:38:12.4364906Z Help         : https://aka.ms/azpipes-docker-tsg
2024-05-28T07:38:12.4364993Z ==============================================================================
2024-05-28T07:38:12.8400661Z [command]/usr/bin/docker run --env REPO_NAME=test-infra --env REPO_OWNER=kyma-project --env CI=true --env JOB_TYPE=presubmit --mount type=bind,src=/agent/_work/1/s/kaniko-build-config.yaml,dst=/kaniko-build-config.yaml --mount type=bind,src=/agent/_work/1/s/signify-prod-secret.yaml,dst=/secret-prod/secret.yaml europe-docker.pkg.dev/kyma-project/prod/image-builder:v20240515-f756e622 --sign-only --name=image-builder --context=. --dockerfile=cmd/image-builder/images/kaniko/Dockerfile --images-to-sign=europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10696 --config=/kaniko-build-config.yaml
2024-05-28T07:38:13.0389131Z sign images using services signify-prod
2024-05-28T07:38:13.0389670Z signer signify-prod ignored, because is not enabled for a CI job of type: presubmit
2024-05-28T07:38:13.0390290Z Start signing images europe-docker.pkg.dev/kyma-project/dev/image-builder:PR-10696
2024-05-28T07:38:13.1669325Z ##[section]Finishing: sign_images`,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			actualImages := extractImagesFromADOLogs(c.logs)

			if !reflect.DeepEqual(actualImages, c.expectedImages) {
				t.Errorf("Expected %v, but got %v", c.expectedImages, actualImages)
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
