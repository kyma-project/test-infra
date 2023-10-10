package main_test

import (
	"os"

	"github.com/kyma-project/test-infra/cmd/external-plugins/automated-approver"
	consolelog "github.com/kyma-project/test-infra/pkg/logging"
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/github/fakegithub"
)

const (
	repoOwner                                      = "kyma-project"
	repoName                                       = "test-infra"
	prNumber                                       = 9046
	prHeadSha                                      = "0ebd2807221fbbc428dba4f09524bec0a8c4cec9"
	pullRequestReviewRequestedEventPayloadFilePath = "test_files/pullRequestReviewRequestedEventPayload"
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
		server               externalplugin.Plugin
		handler              main.HandlerBackend
		ghc                  *fakegithub.FakeClient
		eventPayloadFilePath string
		eventHandler         func(*externalplugin.Plugin, externalplugin.Event)
	)
	BeforeEach(func() {
		logger := consolelog.NewLogger()
		ghc = fakegithub.NewFakeClient()
		handler = main.HandlerBackend{
			LogLevel:               zapcore.DebugLevel,
			WaitForStatusesTimeout: 1,
		}
		handler.PrLocks = make(map[string]map[string]map[int]map[string]context.CancelFunc)
		server = externalplugin.Plugin{}
		server.Name = "automated-approver-test"
		server.WithLogger(logger)
	})

	When("handling event,", func() {
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

		When("matching rule is found,", func() {
			BeforeEach(func() {
				handler.RulesPath = "test_files/automated-approver-rules.yaml"
				err := handler.ReadConfig()
				Expect(err).ShouldNot(HaveOccurred())
				ghc.PullRequestChanges = map[int][]github.PullRequestChange{
					prNumber: {
						{Filename: "test1.yaml"},
					},
				}
			})

			When("pull request tests are successful,", func() {
				BeforeEach(func() {
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
				})

				When("processing pull request synchronize action,", func() {
					// TODO: Implement
				})

				When("processing review dismissed action,", func() {
					// TODO: Implement
				})
			})

			When("pull request tests failed,", func() {
				BeforeEach(func() {
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
					// TODO: Implement
				})

				When("processing review dismissed action,", func() {
					// TODO: Implement
				})
			})
		})

		When("when matching rule is not found,", func() {

			When("user has not defined approval conditions,", func() {
				BeforeEach(func() {
					handler.RulesPath = "test_files/automated-approver-rules-no-user.yaml"
					err := handler.ReadConfig()
					Expect(err).ShouldNot(HaveOccurred())
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					// TODO: Implement
				})

				When("processing review dismissed action,", func() {
					// TODO: Implement
				})
			})

			When("changed files do not match any rule,", func() {
				BeforeEach(func() {
					handler.RulesPath = "test_files/automated-approver-rules-no-file.yaml"
					err := handler.ReadConfig()
					Expect(err).ShouldNot(HaveOccurred())
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					// TODO: Implement
				})

				When("processing review dismissed action,", func() {
					// TODO: Implement
				})
			})

			When("required labels are not present,", func() {
				BeforeEach(func() {
					handler.RulesPath = "test_files/automated-approver-rules-no-label.yaml"
					err := handler.ReadConfig()
					Expect(err).ShouldNot(HaveOccurred())
				})

				When("processing review requested action,", func() {
					BeforeEach(func() {
						eventPayloadFilePath = pullRequestReviewRequestedEventPayloadFilePath
						eventHandler = handler.PullRequestEventHandler
					})

					AssertNotApprovePullRequest()
				})

				When("processing pull request synchronize action,", func() {
					// TODO: Implement
				})

				When("processing review dismissed action,", func() {
					// TODO: Implement
				})
			})
		})
	})
})
