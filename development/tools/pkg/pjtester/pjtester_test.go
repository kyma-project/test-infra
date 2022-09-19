package pjtester

import (
	// "context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	// gogithub "github.com/google/go-github/v42/github"
	"github.com/kyma-project/test-infra/development/github/pkg/git"
	prtagbuildermock "github.com/kyma-project/test-infra/development/tools/pkg/prtagbuilder/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/test-infra/prow/apis/prowjobs/v1"
	"k8s.io/test-infra/prow/config"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/git/localgit"
	k8sgit "k8s.io/test-infra/prow/git/v2"
	"k8s.io/test-infra/prow/github/fakegithub"
)

type FakeGithubClient struct {
	fakegithub.FakeClient
}

func (f *FakeGithubClient) Email() (string, error) {
	panic("implement me")
}

type GitClient struct {
	k8sgit.ClientFactory
	git.GitRepoClient
}

func fakeProwYAMLGetterFactory(presubmits []config.Presubmit, postsubmits []config.Postsubmit) config.ProwYAMLGetter {
	return func(_ *config.Config, _ k8sgit.ClientFactory, _, _ string, _ ...string) (*config.ProwYAML, error) {
		return &config.ProwYAML{
			Presubmits:  presubmits,
			Postsubmits: postsubmits,
		}, nil
	}
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
		/*
			kymaPrAuthor       string
			kymaPrHeadSHA      string
			kymaPrNumber       int
			kymaPrOrg          string
			kymaPrRepo         string
			kymaBaseRef        string
			kymaBaseSHA        string
			fakeRepoPrAuthor   string
			fakeRepoPrHeadSHA  string
			fakeRepoPrNumber   int
			fakeRepoPrOrg      string
			fakeRepoPrRepo     string
			fakeRepoBaseRef    string
			fakeRepoBaseSHA    string
			fakeRepoMainName        string
			fakeRepoProtectedBranch bool
			fakeRepoMerged          bool
			fakeRepoCommitMessage   string
			fakeRepoRefs            prowapi.Refs
			kymaRefs                prowapi.Refs
			testInfraRefs           prowapi.Refs
			ghOptions               *prowflagutil.GitHubOptions
		*/
	)

	BeforeEach(func() {
		/*
			pjPath, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			set data for testing
			pjtesterPrAuthor = "testInfraAuthor"
			pjtesterPrBaseRef = "main"
			pjtesterPrBaseSHA = "testInfraMainSHA"
			pjtesterPrNumber = 12345
			pjtesterPrHeadSHA = "pjtesterPrHeadSHA"
			pjtesterPrOrg = "kyma-project"
			pjtesterPrRepo = "test-infra"
			kymaPrAuthor = "kymaAuthor"
			kymaPrHeadSHA = "kymaPrHeadSHA"
			kymaPrNumber = 1212
			kymaPrOrg = "kyma-project"
			kymaPrRepo = "kyma"
			kymaBaseRef = "master"
			kymaBaseSHA = "kymaBaseSHA"
			fakeRepoPrAuthor = "fakeRepoAuthor"
			fakeRepoPrHeadSHA = "fakeRepoSHA"
			fakeRepoPrNumber = 1515
			fakeRepoPrOrg = "kyma-project"
			fakeRepoPrRepo = "fake-repo"
			fakeRepoBaseRef = "main"
			fakeRepoBaseSHA = "fakeRepoSHA"
			fakeRepoMainName = "main"
			fakeRepoProtectedBranch = false
			fakeRepoMerged = true
			fakeRepoCommitMessage = fmt.Sprintf("Fake Repo commit message (#%s)", strconv.Itoa(fakeRepoPrNumber))
			repoDir := filepath.Clean(fmt.Sprintf("%s/../../../../..", pjPath))
			kymaRefs = prowapi.Refs{
				Org:     kymaPrOrg,
				Repo:    kymaPrRepo,
				BaseRef: kymaBaseRef,
				BaseSHA: kymaBaseSHA,
				Pulls: []prowapi.Pull{{
					Number: kymaPrNumber,
					Author: kymaPrAuthor,
					SHA:    kymaPrHeadSHA,
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", kymaPrOrg, kymaPrRepo),
			}
			testInfraRefs = prowapi.Refs{
				Org:     pjtesterPrOrg,
				Repo:    pjtesterPrRepo,
				BaseRef: pjtesterPrBaseRef,
				BaseSHA: pjtesterPrBaseSHA,
				Pulls: []prowapi.Pull{{
					Number: pjtesterPrNumber,
					Author: pjtesterPrAuthor,
					SHA:    pjtesterPrHeadSHA,
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", pjtesterPrOrg, pjtesterPrRepo),
			}
			fakeRepoRefs = prowapi.Refs{
				Org:     fakeRepoPrOrg,
				Repo:    fakeRepoPrRepo,
				BaseRef: fakeRepoBaseRef,
				BaseSHA: fakeRepoBaseSHA,
				Pulls: []prowapi.Pull{{
					Number: fakeRepoPrNumber,
					Author: fakeRepoPrAuthor,
					SHA:    fakeRepoPrHeadSHA,
				}},
				PathAlias: fmt.Sprintf("github.com/%s/%s", fakeRepoPrOrg, fakeRepoPrRepo),
			}
		*/
	})

	Describe("Loading and validating pjtester.yaml file", func() {
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
					"ConfigPath": Equal("fake/config/file/path.yaml"),
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
									PjPath: "test-infra/development/tools/pkg/pjtester/test_artifacts/",
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
					"PrConfigs":  BeZero(),
					"ConfigPath": BeZero(),
					"PjConfigs": MatchAllFields(Fields{
						"PrConfig": BeZero(),
						"ProwJobs": MatchAllKeys(Keys{
							"kyma-project": MatchAllKeys(Keys{
								"test-infra": SatisfyAll(
									HaveLen(1),
									ContainElement(MatchAllFields(Fields{
										"PjName": Equal("pre-test-infra-validate-dockerfiles"),
										"PjPath": BeZero(),
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
	When("Scheduling test prowjob", func() {
		var (
			ghOptions                  prowflagutil.GitHubOptions
			opts                       options
			ghRepoMock                 *prtagbuildermock.GithubRepoService
			ghPrMock                   *prtagbuildermock.GithubPullRequestsService
			err                        error
			testConfig                 testCfg
			pj                         v1.ProwJob
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
			// os.Setenv("KYMA_PROJECT_DIR", repoDir)
			_ = os.Setenv("PULL_BASE_REF", pjtesterPrBaseRef)
			_ = os.Setenv("PULL_BASE_SHA", pjtesterPrBaseSHA)
			_ = os.Setenv("PULL_NUMBER", strconv.Itoa(pjtesterPrNumber))
			_ = os.Setenv("PULL_PULL_SHA", pjtesterPrHeadSHA)
			_ = os.Setenv("REPO_OWNER", pjtesterPrOrg)
			_ = os.Setenv("REPO_NAME", pjtesterPrRepo)
			_ = os.Setenv("JOB_SPEC", fmt.Sprintf("{\"type\":\"presubmit\",\"job\":\"job-name\",\"buildid\":\"0\",\"prowjobid\":\"uuid\",\"refs\":{\"pjtesterPrOrg\":\"%s\",\"repo\":\"%s\",\"base_ref\":\"%s\",\"base_sha\":\"%s\",\"pulls\":[{\"number\":%d,\"author\":\"%s\",\"sha\":\"%s\"}]}}", pjtesterPrOrg, pjtesterPrRepo, pjtesterPrBaseRef, pjtesterPrBaseSHA, pjtesterPrNumber, pjtesterPrAuthor, pjtesterPrHeadSHA))

			ghOptions = prowflagutil.GitHubOptions{}
			// TODO: test this function in separate context, opts must be defined at the beginning.
			opts = newCommonOptions(ghOptions)

			testConfig, err = readTestCfg(testCfgFile)
			// Make sure config file was load without errors.
			Expect(err).To(BeNil())
			Expect(testConfig).ToNot(BeZero())

			opts.setProwConfigPath(testConfig)

			// Override default job config path
			opts.jobConfigPath = "./test_artifacts/test-job.yaml"
			// Override default job config path
			opts.configPath = "./test_artifacts/test-prow-config.yaml"

			// Create fake GitHub client
			fake := &FakeGithubClient{*fakegithub.NewFakeClient()}
			/*
				fake.PullRequests = map[int]*github.PullRequest{
					kymaPrNumber: {
						User: github.User{
							Login: kymaPrAuthor,
						},
						Head: github.PullRequestBranch{
							SHA: kymaPrHeadSHA,
						},
						Number: kymaPrNumber,
						Base: github.PullRequestBranch{
							SHA: kymaBaseSHA,
							Ref: kymaBaseRef,
						},
					},
					fakeRepoPrNumber: {
						User: github.User{
							Login: fakeRepoPrAuthor,
						},
						Head: github.PullRequestBranch{
							SHA: fakeRepoPrHeadSHA,
						},
						Number: fakeRepoPrNumber,
						Base: github.PullRequestBranch{
							SHA: fakeRepoBaseSHA,
							Ref: fakeRepoBaseRef,
						},
					},
				}
			*/
			opts.githubClient = fake

			// Create fake git client
			opts.gitOptions = git.ClientConfig{}
			lg, gc, err := localgit.NewV2()
			localgit.DefaultBranch(lg.Dir)
			err = lg.MakeFakeRepo(pjtesterPrOrg, pjtesterPrRepo)
			Expect(err).To(Succeed())
			// opts.gitClient, err = opts.gitOptions.NewClient(git.WithGithubClient(opts.githubClient))
			opts.gitClient = GitClient{ClientFactory: gc}
			// Make sure gitClient was created without errors.
			Expect(err).To(BeNil())
			Expect(opts.gitClient).ToNot(BeZero())

			// ctx := context.Background()
			// Mock prFinder
			opts.prFinder = prtagbuildermock.NewFakeGitHubClient(&http.Client{})
			ghRepoMock = new(prtagbuildermock.GithubRepoService)
			ghPrMock = new(prtagbuildermock.GithubPullRequestsService)
			/*
				ghRepoMock.On("GetBranch", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoBaseRef).Return(&gogithub.Branch{
					Name: &fakeRepoMainName,
					Commit: &gogithub.RepositoryCommit{
						Commit: &gogithub.Commit{
							SHA:     &fakeRepoBaseSHA,
							Message: &fakeRepoCommitMessage,
						},
						SHA: &fakeRepoBaseSHA,
					},
					Protected: &fakeRepoProtectedBranch,
				}, nil, nil)
				ghPrMock.On("Get", ctx, fakeRepoPrOrg, fakeRepoPrRepo, fakeRepoPrNumber).Return(&gogithub.PullRequest{
					Merged:         &fakeRepoMerged,
					MergeCommitSHA: &fakeRepoPrHeadSHA,
				}, nil, nil)
			*/
			opts.prFinder.Repositories = ghRepoMock
			opts.prFinder.PullRequests = ghPrMock

		})
		Context("with test prowjob definition in test-infra", func() {
			Context("with pjtester pull request in test-infra ", func() {
				/*
					Context("with provided pull request with test prowjob definition", func() {
						Context("with defined test prowjob pull requests", func() {
							It("", func() {
								if &testCfg.PrConfigs != nil {
									pullRequests, err := opts.getPullRequests(testCfg.PrConfigs)
									if err != nil {
										log.WithError(err).Fatal("Failed get pull request deatils from GitHub.")
									}
									opts.testPullRequests = pullRequests
								}
								if &testCfg.PjConfigs.PrConfig != nil {
									pullRequests, err := opts.getPullRequests(testCfg.PjConfigs.PrConfig)
									if err != nil {
										log.WithError(err).Fatal("Failed get pull request deatils from GitHub.")
									}
									// There is only one element in each map level. This is enforced by testCfg struct values validation.
									for _, prorg := range pullRequests {
										for _, prconfig := range prorg {
											opts.pjConfigPullRequest = prconfig
										}
									}
								}
								// Go over prowjob names to test and create prowjob definitions for each.
								for orgName, pjOrg := range testCfg.PjConfigs.ProwJobs {
									for repoName, pjconfigs := range pjOrg {
										for _, pjconfig := range pjconfigs {
											// generate prowjob specification to test.
											// pj, err := newTestPJ(pjconfig, opts, orgName, repoName)
											pj, _ := newTestPJ(pjconfig, opts, orgName, repoName)
											fmt.Printf("%+v", pj)
										}
									}
								}
							})
						})
						Context("without defined test prowjob pull requests", func() {

						})
					})
				*/
				Context("without provided pull request with test prowjob definition", func() {
					/*
						Context("with defined test prowjob pull requests", func() {

						})
					*/
					Context("without defined test prowjob pull requests", func() {
						BeforeEach(func() {
							testCfgFile = "./test_artifacts/no_prconfigs-no_prowjob_prconfig-pjtester.yaml"
							Expect(testCfgFile).Should(BeAnExistingFile(), "Pjtester test config file doesn't exist")

							// Override default job config path
							opts.jobConfigPath = "./test_artifacts/test-job.yaml"
							// Override default job config path
							opts.configPath = "./test_artifacts/test-prow-config.yaml"

							ProwYAMLGetterWithDefaults = fakeProwYAMLGetterFactory(
								[]config.Presubmit{
									{
										JobBase: config.JobBase{Name: "hans"},
									},
								},
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
							for orgName, pjOrg := range testConfig.PjConfigs.ProwJobs {
								for repoName, pjconfigs := range pjOrg {
									for _, pjconfig := range pjconfigs {
										// generate prowjob specification to test.
										// TODO: how to mock setjobconfigpath
										err = opts.setRefsGetters(orgName, repoName)
										Expect(err).To(BeNil())
										opts.setJobConfigPath(pjconfig, orgName, repoName)
										Expect(opts.jobConfigPath).To(Equal("./test_artifacts/test-job.yaml"))
										pj, err = newTestPJ(pjconfig, opts, orgName, repoName)
									}
								}
							}
							Expect(testConfig.PrConfigs).To(BeNil())
							Expect(testConfig.PjConfigs.PrConfig).To(BeNil())
							Expect(testConfig.ConfigPath).To(Equal("./test_artifacts/test-prow-config.yaml"))
							Expect(err).ToNot(BeNil())
							Expect(pj).To(MatchFields(3, Fields{
								"Spec": MatchFields(3, Fields{
									"Job": Equal("test-infra-presubmit-test-job"),
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
})
