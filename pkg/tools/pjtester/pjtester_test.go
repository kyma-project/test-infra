package pjtester

import (
	"context"
	"fmt"
	"github.com/kyma-project/test-infra/pkg/github/git"
	prtagbuildermock "github.com/kyma-project/test-infra/pkg/tools/prtagbuilder/mocks"
	"net/http"
	"os"
	"strconv"

	gogithub "github.com/google/go-github/v48/github"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	prowapi "sigs.k8s.io/prow/prow/apis/prowjobs/v1"
	"sigs.k8s.io/prow/prow/config"
	prowflagutil "sigs.k8s.io/prow/prow/flagutil"
	"sigs.k8s.io/prow/prow/git/localgit"
	k8sgit "sigs.k8s.io/prow/prow/git/v2"
	"sigs.k8s.io/prow/prow/github/fakegithub"
	"sigs.k8s.io/prow/prow/pjutil"
)

type FakeGithubClient struct {
	fakegithub.FakeClient
}

func (f *FakeGithubClient) Email() (string, error) {
	panic("implement me")
}

type GitClient struct {
	k8sgit.ClientFactory
	git.RepoClient
}

func fakeProwYAMLGetterFactory(presubmits []config.Presubmit, postsubmits []config.Postsubmit) config.ProwYAMLGetter {
	return func(_ *config.Config, _ k8sgit.ClientFactory, _, _ string, _ ...string) (*config.ProwYAML, error) {
		return &config.ProwYAML{
			Presubmits:  presubmits,
			Postsubmits: postsubmits,
		}, nil
	}
}

func refsPullsElement(element interface{}) string {
	return strconv.Itoa(element.(prowapi.Pull).Number)
}

func extraRefsElement(element interface{}) string {
	return (element.(prowapi.Refs).Org) + "/" + (element.(prowapi.Refs).Repo)
}

var _ = Describe("Pjtester", func() {
	var (
		testCfgFile       string
		pjtesterPrAuthor  string
		pjtesterPrBaseRef string
		pjtesterPrBaseSHA string
		pjtesterPrNumber  int
		pjtesterPrHeadSHA string
		pjtesterPrOrg     string
		pjtesterPrRepo    string
		// testInfraPrAuthor        string
		// testInfraPrHeadSHA       string
		testInfraPrNumber        int
		testInfraPrOrg           string
		testInfraPrRepo          string
		testInfraBaseRef         string
		testInfraBaseSHA         string
		testInfraMainName        string
		testInfraProtectedBranch bool
		testInfraMerged          bool
		testInfraCommitMessage   string
	)

	Describe("loading and validating pjtester.yaml file", func() {
		Context("with valid full config file", func() {
			BeforeEach(func() {
				testCfgFile = "./test_artifacts/full-pjtester.yaml"
				Expect(testCfgFile).Should(BeAnExistingFile())
			})
			It("create pjtester.testCfg instance without errors", func() {
				testConfig, err := readTestCfg(testCfgFile)

				Expect(err).To(BeNil())
				Expect(testConfig).ToNot(BeZero())
			})
			It("create valid pjtester.testCfg instance", func() {
				testConfig, _ := readTestCfg(testCfgFile)

				Expect(testConfig).To(MatchAllFields(Fields{
					// "ConfigPath": Equal("fake/config/file/path.yaml"),
					"PjConfigs": MatchAllFields(Fields{
						"PrConfig": MatchAllKeys(Keys{
							"kyma-project": MatchAllKeys(Keys{
								"test-infra": MatchFields(IgnoreExtras, Fields{
									"PrNumber": Equal(1414),
								}),
							}),
						}),
						"ProwJobs": MatchAllKeys(Keys{
							"kyma-project": MatchAllKeys(Keys{
								"test-infra": ConsistOf(pjConfig{
									PjName: "pre-test-infra-validate-dockerfiles",
								}, pjConfig{
									PjName: "test-infra-presubmit-test-job",
									// PjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/",
									Report: true,
								}),
							}),
							"kyma-incubator": MatchAllKeys(Keys{
								"busola": ConsistOf(pjConfig{PjName: "busola-fake-pjname"}),
							}),
						}),
					}),
					"PrConfigs": MatchAllKeys(Keys{
						"kyma-project": MatchAllKeys(Keys{
							"kyma": MatchFields(IgnoreExtras, Fields{
								"PrNumber": Equal(1212),
							}),
							"test-infra": MatchFields(IgnoreExtras, Fields{
								"PrNumber": Equal(1313),
							}),
						}),
					}),
				}), "pjtester created pjtester.testCfg instance which doesn't match input")
			})
		})

		Context("with valid minimal config file", func() {
			BeforeEach(func() {
				testCfgFile = "./test_artifacts/minimal-pjtester.yaml"
				Expect(testCfgFile).Should(BeAnExistingFile(), "Pjtester test config file doesn't exist")
			})

			It("create pjtester.testCfg instance without errors", func() {
				testConfig, err := readTestCfg(testCfgFile)

				Expect(err).To(BeNil())
				Expect(testConfig).ToNot(BeZero())
			})

			It("create valid pjtester.testCfg instance", func() {
				testConfig, _ := readTestCfg(testCfgFile)

				Expect(testConfig).To(MatchAllFields(Fields{
					"PrConfigs": BeZero(),
					// "ConfigPath": BeZero(),
					"PjConfigs": MatchAllFields(Fields{
						"PrConfig": BeZero(),
						"ProwJobs": MatchAllKeys(Keys{
							"kyma-project": MatchAllKeys(Keys{
								"test-infra": SatisfyAll(
									HaveLen(1),
									ContainElement(MatchAllFields(Fields{
										"PjName": Equal("pre-test-infra-validate-dockerfiles"),
										// "PjPath": BeZero(),
										"Report": BeFalse(),
									}))),
							}),
						}),
					}),
				}), "pjtester created pjtester.testCfg instance which doesn't match input")
			})
		})

		Context("with invalid config file", func() {
			BeforeEach(func() {
				testCfgFile = "./test_artifacts/invalid-pjtester.yaml"
				Expect(testCfgFile).Should(BeAnExistingFile(), "Pjtester test config file doesn't exist")
			})

			It("raise validation error", func() {
				testConfig, err := readTestCfg(testCfgFile)

				Expect(err).ToNot(BeNil())
				Expect(testConfig).To(BeZero())
			})
		})
	})
	Describe("scheduling test prowjob", func() {
		var (
			ghOptions                  prowflagutil.GitHubOptions
			opts                       options
			ghRepoMock                 *prtagbuildermock.GithubRepoService
			ghPrMock                   *prtagbuildermock.GithubPullRequestsService
			err                        error
			testConfig                 testCfg
			pj                         prowapi.ProwJob
			ProwYAMLGetterWithDefaults config.ProwYAMLGetter
		)
		JustBeforeEach(func() {
			pjtesterPrAuthor = "pjtesterPrAuthor"
			pjtesterPrBaseRef = "main"
			pjtesterPrBaseSHA = "pjtesterPrBaseSHA"
			pjtesterPrNumber = 12345
			pjtesterPrHeadSHA = "pjtesterPrHeadSHA"
			pjtesterPrOrg = "kyma-project"
			pjtesterPrRepo = "test-infra"

			// set env variables for pjtester pull request
			_ = os.Setenv("PULL_BASE_REF", pjtesterPrBaseRef)
			_ = os.Setenv("PULL_BASE_SHA", pjtesterPrBaseSHA)
			_ = os.Setenv("PULL_NUMBER", strconv.Itoa(pjtesterPrNumber))
			_ = os.Setenv("PULL_PULL_SHA", pjtesterPrHeadSHA)
			_ = os.Setenv("REPO_OWNER", pjtesterPrOrg)
			_ = os.Setenv("REPO_NAME", pjtesterPrRepo)
			_ = os.Setenv("JOB_SPEC", fmt.Sprintf("{\"type\":\"presubmit\",\"job\":\"job-name\",\"buildid\":\"0\",\"prowjobid\":\"uuid\",\"refs\":{\"pjtesterPrOrg\":\"%s\",\"repo\":\"%s\",\"base_ref\":\"%s\",\"base_sha\":\"%s\",\"pulls\":[{\"number\":%d,\"author\":\"%s\",\"sha\":\"%s\"}]}}", pjtesterPrOrg, pjtesterPrRepo, pjtesterPrBaseRef, pjtesterPrBaseSHA, pjtesterPrNumber, pjtesterPrAuthor, pjtesterPrHeadSHA))

			ghOptions = prowflagutil.GitHubOptions{}
			// TODO: test this function in separate context, opts must be defined at the beginning.
			opts, err = newCommonOptions(&ghOptions)
			Expect(err).To(Succeed())

			testConfig, err = readTestCfg(testCfgFile)
			// Make sure config file was load without errors.
			Expect(err).To(BeNil())
			Expect(testConfig).ToNot(BeZero())

			// Override default job config path
			opts.jobConfigPath = "./test_artifacts/test-job.yaml"
			// Override default job config path
			opts.configPath = "./test_artifacts/test-prow-config.yaml"

			// Create fake GitHub client
			fake := &FakeGithubClient{*fakegithub.NewFakeClient()}
			// TODO: merge map from BeforeEach with existing map under fakeGitHub client.
			/*
				fake.PullRequests = map[int]*github.PullRequest{
					testInfraPrNumber: {
						User: github.User{
							Login: testInfraPrAuthor,
						},
						Head: github.PullRequestBranch{
							SHA: testInfraPrHeadSHA,
						},
						Number: testInfraPrNumber,
						Base: github.PullRequestBranch{
							SHA: testInfraBaseSHA,
							Ref: testInfraBaseRef,
						},
					},
				}

			*/
			opts.githubClient = fake

			// Create fake git client
			opts.gitOptions = git.ClientConfig{}
			lg, gc, err := localgit.NewV2()
			// localgit.DefaultBranch(lg.Dir)
			err = lg.MakeFakeRepo(pjtesterPrOrg, pjtesterPrRepo)
			Expect(err).To(Succeed())
			// opts.gitClient, err = opts.gitOptions.NewClient(git.WithGithubClient(opts.githubClient))
			opts.gitClient = GitClient{ClientFactory: gc}
			// Make sure gitClient was created without errors.
			Expect(err).To(BeNil())
			Expect(opts.gitClient).ToNot(BeZero())

			// Mock prFinder
			opts.prFinder = prtagbuildermock.NewFakeGitHubClient(&http.Client{})
			ghRepoMock = new(prtagbuildermock.GithubRepoService)
			ghPrMock = new(prtagbuildermock.GithubPullRequestsService)
			ghRepoMock.On("GetBranch", context.Background(), testInfraPrOrg, testInfraPrRepo, testInfraBaseRef).Return(&gogithub.Branch{
				Name: &testInfraMainName,
				Commit: &gogithub.RepositoryCommit{
					Commit: &gogithub.Commit{
						SHA:     &testInfraBaseSHA,
						Message: &testInfraCommitMessage,
					},
					SHA: &testInfraBaseSHA,
				},
				Protected: &testInfraProtectedBranch,
			}, nil, nil)
			ghPrMock.On("Get", context.Background(), testInfraPrOrg, testInfraPrRepo, testInfraPrNumber).Return(&gogithub.PullRequest{
				Merged:         &testInfraMerged,
				MergeCommitSHA: &testInfraBaseSHA,
			}, nil, nil)
			opts.prFinder.Repositories = ghRepoMock
			opts.prFinder.PullRequests = ghPrMock

		})
		When("test prowjob definition is in test-infra", func() {
			When("pjtester pull request is in test-infra ", func() {
				/*
					Context("with provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
				*/
				When("not provided pull request with test prowjob definition", func() {
					/*
						Context("with defined test prowjob pull requests", func() {
						})
					*/
					When("not provided test prowjob pull requests", func() {
						BeforeEach(func() {
							// testInfraPrAuthor = "testInfraPrAuthor"
							// testInfraPrHeadSHA = "testInfraHeadSHA"
							testInfraPrNumber = 1515
							testInfraPrOrg = "kyma-project"
							testInfraPrRepo = "test-infra"
							testInfraBaseRef = "main"
							testInfraBaseSHA = "fakeRepoSHA"
							testInfraMainName = "main"
							testInfraProtectedBranch = false
							testInfraMerged = true
							testInfraCommitMessage = fmt.Sprintf("Fake Repo commit message (#%s)", strconv.Itoa(testInfraPrNumber))
							testCfgFile = "./test_artifacts/no_prconfigs-no_prowjob_prconfig-pjtester.yaml"
							Expect(testCfgFile).Should(BeAnExistingFile(), "Pjtester test config file doesn't exist")

							// Return nil, nil as there is no inrepo config tested in this scenario
							ProwYAMLGetterWithDefaults = fakeProwYAMLGetterFactory(
								nil,
								nil,
							)

						})
						It("create valid prowjob specification", func() {
							if &testConfig.PrConfigs != nil {
								pullRequests, err := opts.getPullRequests(testConfig.PrConfigs)
								Expect(err).To(BeNil())
								opts.testPullRequests = pullRequests
							}
							if &testConfig.PjConfigs.PrConfig != nil {
								pullRequests, err := opts.getPullRequests(testConfig.PjConfigs.PrConfig)
								Expect(err).To(BeNil())
								for _, prorg := range pullRequests {
									for _, prconfig := range prorg {
										opts.pjConfigPullRequest = prconfig
									}
								}
							}
							for org, pjOrg := range testConfig.PjConfigs.ProwJobs {
								for repo, pjconfigs := range pjOrg {
									for _, pjCfg := range pjconfigs {
										// generate prowjob specification to test.
										testPjOpts := testProwJobOptions{
											repoName: repo,
											orgName:  org,
											pjConfig: pjCfg,
										}
										conf, err := config.Load(opts.configPath, opts.jobConfigPath, nil, "")
										Expect(err).To(Succeed())
										Expect(conf).Should(BeAssignableToTypeOf(&config.Config{}))
										conf.ProwYAMLGetterWithDefaults = ProwYAMLGetterWithDefaults
										err = testPjOpts.setRefsGetters(opts)
										Expect(err).To(Succeed())
										_, pjSpecification, err := testPjOpts.genJobSpec(opts, conf, pjCfg)
										Expect(err).To(Succeed())
										pj = pjutil.NewProwJob(pjSpecification, map[string]string{"created-by-pjtester": "true", "prow.k8s.io/is-optional": "true"}, map[string]string{})
										pj.Spec.Cluster = "untrusted-workload"
										if pjCfg.Report {
											pj.Spec.Report = true
										} else {
											pj.Spec.ReporterConfig = &prowapi.ReporterConfig{Slack: &prowapi.SlackReporterConfig{Channel: "kyma-prow-dev-null"}}
										}
									}
								}
							}
							Expect(testConfig.PrConfigs).To(BeNil())
							Expect(testConfig.PjConfigs.PrConfig).To(BeNil())
							Expect(pj).To(MatchFields(3, Fields{
								"Spec": MatchFields(3, Fields{
									"Job":     Equal("pjtesterprauthor_test_of_prowjob_test-infra-presubmit-test-job"),
									"Context": Equal("pjtesterprauthor_test_of_prowjob_test-infra-presubmit-test-job"),
									"Cluster": Equal("untrusted-workload"),
									"Report":  BeTrue(),
									"ReporterConfig": PointTo(MatchAllFields(Fields{
										"Slack": PointTo(MatchFields(3, Fields{
											"Channel": Equal("kyma-prow-dev-null"),
										})),
									})),
									"Refs": PointTo(MatchFields(3, Fields{
										"Org":     Equal("kyma-project"),
										"Repo":    Equal("test-infra"),
										"BaseRef": Equal(pjtesterPrBaseRef),
										"BaseSHA": Equal(pjtesterPrBaseSHA),
										"Pulls": MatchAllElements(refsPullsElement, Elements{
											"12345": MatchFields(3, Fields{
												"Number": Equal(pjtesterPrNumber),
												"Author": Equal(pjtesterPrAuthor),
												"SHA":    Equal(pjtesterPrHeadSHA),
											}),
										}),
									})),
									"ExtraRefs": MatchAllElements(extraRefsElement, Elements{
										"kyma-project/kyma": MatchFields(3, Fields{
											"Org":     Equal("kyma-project"),
											"Repo":    Equal("kyma"),
											"BaseRef": Equal("main"),
											"BaseSHA": BeZero(),
											"Pulls":   BeNil(),
										}),
									}),
								}),
							}))
						})
					})
				})
			})
			/*
				Context("with pjtester pull request not in test-infra", func() {
					Context("with provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
					Context("without provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
				})
			*/
		})
		/*
			Context("with test prowjob definition not in test-infra", func() {
				Context("with pjtester pull request in test-infra ", func() {
					Context("with provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
					Context("without provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
				})
				Context("with pjtester pull request not in test-infra", func() {
					Context("with provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
					Context("without provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
						})
						Context("without defined test prowjob pull requests", func() {
						})
					})
				})
			})
		*/
	})
	/*
		When("Creating new common options", func() {

		})
	*/
})
