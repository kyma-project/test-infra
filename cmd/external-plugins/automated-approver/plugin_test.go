package main_test

import (
	"os"
	"sync"

	"github.com/kyma-project/test-infra/cmd/external-plugins/automated-approver"
	consolelog "github.com/kyma-project/test-infra/pkg/logging"
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"sigs.k8s.io/prow/prow/github"
	"sigs.k8s.io/prow/prow/github/fakegithub"
)

const (
	repoOwner                                      = "kyma-project"
	repoName                                       = "test-infra"
	prNumber                                       = 9046
	prHeadSha                                      = "0ebd2807221fbbc428dba4f09524bec0a8c4cec9"
	oldPrHeadSha                                   = "85z1hk4veijfd6c4itym06mqct1m6inr"
	pullRequestReviewRequestedEventPayloadFilePath = "test_files/pullRequestReviewRequestedEventPayload"
	pullRequestSynchronizeEventPayloadFilePath     = "test_files/pullRequestSynchronizeEventPayload"
	oldPullRequestSynchronizeEventPayloadFilePath  = "test_files/oldPullRequestSynchronizeEventPayload"
	pullRequestReviewDismissedEventPayloadFilePath = "test_files/pullRequestReviewDismissedEventPayload"
)

func setupEventHelper(eventPayloadFilePath string) (externalplugin.Event, error) {
	GinkgoHelper()
	var pullRequestEvent externalplugin.Event
	file, err := os.ReadFile(eventPayloadFilePath)
	pullRequestEvent.Payload = file
	pullRequestEvent.EventType = "pull_request"
	pullRequestEvent.EventGUID = "1234"
	return pullRequestEvent, err
}

var _ = Describe("automated-approver", func() {
	var (
		server                  externalplugin.Plugin
		handler                 main.HandlerBackend
		ghc                     *fakegithub.FakeClient
		eventPayloadFilePath    string
		oldEventPayloadFilePath string
		eventHandler            func(*externalplugin.Plugin, externalplugin.Event)
		wg                      sync.WaitGroup
	)
	BeforeEach(func() {
		logger, level := consolelog.NewLoggerWithLevel()
		// level.SetLevel(zapcore.DebugLevel)
		level.SetLevel(zapcore.InfoLevel)
		ghc = fakegithub.NewFakeClient()
		// We are testing methods on HandlerBackend struct. We need to initialize it and set needed fields.
		handler = main.HandlerBackend{
			// LogLevel: zapcore.DebugLevel,
			LogLevel:                       zapcore.InfoLevel,
			WaitForStatusesTimeout:         1,
			WaitForContextsCreationTimeout: 1,
		}
		// Tested methods are using prLocks map. We need to initialize it.
		handler.PrLocks = make(map[string]map[string]map[int]map[string]context.CancelFunc)
		// Plugin is passed as parameter to tested methods. We need to initialize it.
		server = externalplugin.Plugin{}
		server.Name = "automated-approver-test"
		server.WithLogger(logger)
	})

	When("handling event,", func() {
		// AssertApprovePullRequest asserts that pull request is approved.
		// This is a helper function for testing to avoid code duplication.
		// It contains ginkgo subject node definition.
		// It's called in multiple testing scenarios to make assertions.
		AssertApprovePullRequest := func() {
			It("should approve pull request", func() {
				pullRequestEvent, err := setupEventHelper(eventPayloadFilePath)
				Expect(err).ShouldNot(HaveOccurred())
				eventHandler(&server, pullRequestEvent)
				prReviews, err := ghc.ListReviews(repoOwner, repoName, prNumber)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(prReviews).Should(HaveLen(1))
				Expect(prReviews[0].User.Login).Should(Equal("k8s-ci-robot"))
				Expect(ghc.IssueLabelsAdded).Should(HaveLen(1))
				Expect(ghc.IssueLabelsAdded).Should(ContainElement("kyma-project/test-infra#9046:auto-approved"))
			})
		}

		// AssertApprovePullRequestOnlyOnceForManyEvents asserts that pull request is approved only once when processed
		// the same pull request in parallel in more than one threads.
		// This is a helper function for testing to avoid code duplication.
		// It contains ginkgo subject node definition.
		// It's called in multiple testing scenarios to make assertions.
		AssertApprovePullRequestOnlyOnceForManyEvents := func() {
			It("should approve pull request only once", func() {
				prEvents := make(map[int]externalplugin.Event)
				oldPullRequestEvent, err := setupEventHelper(oldEventPayloadFilePath)
				Expect(err).ShouldNot(HaveOccurred())
				prEvents[1] = oldPullRequestEvent
				pullRequestEvent, err := setupEventHelper(eventPayloadFilePath)
				Expect(err).ShouldNot(HaveOccurred())
				prEvents[2] = pullRequestEvent
				wg.Add(2)
				for i := 1; i <= 2; i++ {
					i := i
					go func() {
						eventHandler(&server, prEvents[i])
						wg.Done()
					}()
				}
				wg.Wait()
				prReviews, err := ghc.ListReviews(repoOwner, repoName, prNumber)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(prReviews).Should(HaveLen(1))
				Expect(prReviews[0].User.Login).Should(Equal("k8s-ci-robot"))
				Expect(prReviews[0].ID).Should(Equal(1))
			})
		}

		// AssertNotApprovePullRequest asserts that pull request is not approved.
		// This is a helper function for testing to avoid code duplication.
		// It contains ginkgo subject node definition.
		// It's called in multiple testing scenarios to make assertions.
		AssertNotApprovePullRequest := func() {
			It("should not approve pull request", func() {
				pullRequestEvent, err := setupEventHelper(eventPayloadFilePath)
				Expect(err).ShouldNot(HaveOccurred())
				eventHandler(&server, pullRequestEvent)
				prReviews, err := ghc.ListReviews(repoOwner, repoName, prNumber)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(prReviews).Should(BeEmpty())
				Expect(ghc.IssueLabelsAdded).Should(BeEmpty())
			})
		}

		BeforeEach(func() {
			// Adding pull request changed files definition to fake GitHub client.
			ghc.PullRequestChanges = map[int][]github.PullRequestChange{
				prNumber: {
					{Filename: "test1.yaml"},
				},
			}
		})

		When("matching rule is found,", func() {
			BeforeEach(func() {
				// Reading rules from test rules.yaml file. Test events are validated against these rules.
				handler.RulesPath = "test_files/automated-approver-rules.yaml"
				err := handler.ReadConfig()
				Expect(err).ShouldNot(HaveOccurred())
			})

			When("pull request tests are successful,", func() {
				BeforeEach(func() {
					// Adding pull request tests statuses to fake GitHub client.
					ghc.CombinedStatuses = map[string]*github.CombinedStatus{
						prHeadSha: {
							State: github.StatePending,
							Statuses: []github.Status{
								{
									State:   github.StatusPending,
									Context: "tide",
								},
								{
									State:   github.StatusSuccess,
									Context: "test1",
								},
							},
						},
					}
					handler.Ghc = ghc
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertApprovePullRequest()

					When("processing multiple events for the same commit,", func() {
						BeforeEach(func() {
							oldEventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						})

						AssertApprovePullRequestOnlyOnceForManyEvents()
					})
				})

				When("processing pull request synchronize action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertApprovePullRequest()

					When("processing multiple events for the same commit,", func() {
						BeforeEach(func() {
							oldEventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						})

						AssertApprovePullRequestOnlyOnceForManyEvents()
					})

					When("processing multiple events for old and current commits,", func() {
						BeforeEach(func() {
							// Adding pull request tests statuses to fake GitHub client for old commit sha.
							ghc.CombinedStatuses[oldPrHeadSha] = &github.CombinedStatus{
								State: github.StatePending,
								Statuses: []github.Status{
									{
										State:   github.StatusPending,
										Context: "tide",
									},
									{
										State:   github.StatusSuccess,
										Context: "test1",
									},
								},
							}
							handler.Ghc = ghc
							oldEventPayloadFilePath = oldPullRequestSynchronizeEventPayloadFilePath
						})

						AssertApprovePullRequestOnlyOnceForManyEvents()
					})
				})

				When("processing review dismissed action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewDismissedEventPayloadFilePath
						eventHandler = handler.PullRequestReviewEventHandler
					})

					AssertApprovePullRequest()

					When("processing multiple events for the same commit,", func() {
						BeforeEach(func() {
							oldEventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						})

						AssertApprovePullRequestOnlyOnceForManyEvents()
					})
				})
			})

			When("pull request tests failed,", func() {
				BeforeEach(func() {
					// Adding pull request tests statuses to fake GitHub client.
					ghc.CombinedStatuses = map[string]*github.CombinedStatus{
						prHeadSha: {
							State: github.StatePending,
							Statuses: []github.Status{
								{
									State:   github.StatusPending,
									Context: "tide",
								},
								{
									State:   github.StatusFailure,
									Context: "test1",
								},
							},
						},
					}
					handler.Ghc = ghc
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing review dismissed action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewDismissedEventPayloadFilePath
						eventHandler = handler.PullRequestReviewEventHandler
					})

					AssertNotApprovePullRequest()
				})
			})
		})

		When("when matching rule is not found,", func() {

			When("user has not defined approval conditions,", func() {
				BeforeEach(func() {
					// Reading rules from test rules.yaml file. Test events are validated against these rules.
					handler.RulesPath = "test_files/automated-approver-rules-no-user.yaml"
					err := handler.ReadConfig()
					Expect(err).ShouldNot(HaveOccurred())
					handler.Ghc = ghc
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing review dismissed action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewDismissedEventPayloadFilePath
						eventHandler = handler.PullRequestReviewEventHandler
					})

					AssertNotApprovePullRequest()
				})
			})

			When("changed files do not match any rule,", func() {
				BeforeEach(func() {
					// Reading rules from test rules.yaml file. Test events are validated against these rules.
					handler.RulesPath = "test_files/automated-approver-rules-no-file.yaml"
					err := handler.ReadConfig()
					Expect(err).ShouldNot(HaveOccurred())
					handler.Ghc = ghc
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing review dismissed action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewDismissedEventPayloadFilePath
						eventHandler = handler.PullRequestReviewEventHandler
					})

					AssertNotApprovePullRequest()
				})
			})

			When("required labels are not present,", func() {
				BeforeEach(func() {
					// Reading rules from test rules.yaml file. Test events are validated against these rules.
					handler.RulesPath = "test_files/automated-approver-rules-no-label.yaml"
					err := handler.ReadConfig()
					Expect(err).ShouldNot(HaveOccurred())
					handler.Ghc = ghc
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestSynchronizeEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing review dismissed action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewDismissedEventPayloadFilePath
						eventHandler = handler.PullRequestReviewEventHandler
					})

					AssertNotApprovePullRequest()
				})
			})
		})
	})
})
