package main

import (
	"context"
	"fmt"
	"time"

	// "encoding/json"
	// "fmt"
	// "os"
	// "regexp"
	"sync"
	// "time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fsnotify/fsnotify"
	// "golang.org/x/exp/slices"
	// "gopkg.in/yaml.v3"

	// consolelog "github.com/kyma-project/test-infra/pkg/logging"
	// "github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	githubql "github.com/shurcooL/githubv4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/prow/pkg/github"
)

type githubClient interface {
	CreatePullRequestReviewComment(org, repo string, number int, rc github.ReviewComment) error
	CreateReview(org, repo string, number int, r github.DraftReview) error
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
	AddLabel(org, repo string, number int, label string) error
	CreateComment(org, repo string, number int, comment string) error
	MutateWithGitHubAppsSupport(ctx context.Context, m interface{}, input githubql.Input, vars map[string]interface{}, org string) error
}

// HandlerBackend is a backend for the plugin.
// It contains all the configuration and clients needed to handle events.
type HandlerBackend struct {
	Ghc                            githubClient
	LogLevel                       zapcore.Level                                                                       // Log level is read in backend handlers to keep the same log level for all logs.
	WaitForStatusesTimeout         int                                                                                 // in seconds
	WaitForContextsCreationTimeout int                                                                                 // in seconds
	RulesPath                      string                                                                              // Path to yaml config file
	AutoMergeRulesPath             string                                                                              // Path to yaml config file with auto merge rules
	Conditions                     map[ownerString]map[repoString]map[userString][]ApproveCondition                    `yaml:"conditions"`
	MergeConditions                map[ownerString]map[repoString][]AutoMergeCondition                                 `yaml:"autoMergeConditions"`
	PrLocks                        map[ownerString]map[repoString]map[prNumberInt]map[headShaString]context.CancelFunc // Holds head sha and cancel function of PRs that are being processed. org -> repo -> pr number -> head sha -> cancel function
	PrMutex                        sync.Mutex
}

// TODO (dekiel): Refactor data structure to hold PRs, locks and conditions to not use nested maps.
//
//	Remove aliases and use types directly.
type ownerString = string
type repoString = string
type userString = string
type prNumberInt = int
type headShaString = string

// WatchConfig watches for changes in the rules file and reads it again when a file change occurs.
// TODO: Refactor function to reflect it's working with rules file not configuration file.
func (hb *HandlerBackend) WatchConfig(logger *zap.SugaredLogger) {
	defer logger.Sync()
	logger.Info("Starting config watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logger.Infof("%s %s", event.Name, event.Op)
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					logger.Info("Reloading config")
					err := hb.ReadConfig()
					if err != nil {
						logger.Fatalf("Failed reading config: %s", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("error: %s", err)
			}
		}

	}()

	err = watcher.Add(hb.RulesPath)
	if err != nil {
		logger.Fatalf("Add failed: %s", err)
	}
	logger.Info("Waiting for available rules file.")
	<-done
}

// lockPR locks PR for processing by adding head sha to PrLocks.
// If PR is already locked, returns false.
// Because GitHub sends multiple review request events for one PR, we need to lock PR to avoid processing it multiple times.
// GitHub sends multiple events because it sends one event for each reviewer.
func (hb *HandlerBackend) lockPR(cancel context.CancelFunc, logger *zap.SugaredLogger, org, repo, headSha string, prNumber int) bool {
	// Sync access to PrLocks with mutex.
	hb.PrMutex.Lock()
	defer hb.PrMutex.Unlock()
	defer logger.Sync()
	_, ok := hb.PrLocks[org][repo][prNumber][headSha]
	if !ok {
		if hb.PrLocks[org] == nil {
			hb.PrLocks[org] = make(map[string]map[int]map[string]context.CancelFunc)
		}
		if hb.PrLocks[org][repo] == nil {
			hb.PrLocks[org][repo] = make(map[int]map[string]context.CancelFunc)
		}
		if hb.PrLocks[org][repo][prNumber] == nil {
			hb.PrLocks[org][repo][prNumber] = make(map[string]context.CancelFunc)
		}
		hb.PrLocks[org][repo][prNumber][headSha] = cancel
		return true
	}
	return false
}

// unlockPR unlocks PR by removing head sha from PrLocks.
func (hb *HandlerBackend) unlockPR(logger *zap.SugaredLogger, org, repo, headSha string, prNumber int) {
	// Sync access to PrLocks with mutex.
	hb.PrMutex.Lock()
	defer hb.PrMutex.Unlock()
	defer logger.Sync()
	delete(hb.PrLocks[org][repo][prNumber], headSha)
	if len(hb.PrLocks[org][repo][prNumber]) == 0 {
		delete(hb.PrLocks[org][repo], prNumber)
	}
}

// cancelPR cancels processing of PR for defined head commit sha. It calls cancel function assigned to head sha in PrLocks.
func (hb *HandlerBackend) cancelPR(logger *zap.SugaredLogger, org, repo, headSha string, prNumber int) {
	// Sync access to PrLocks with mutex.
	hb.PrMutex.Lock()
	defer hb.PrMutex.Unlock()
	defer logger.Sync()
	if pr, ok := hb.PrLocks[org][repo][prNumber]; ok {
		for sha, cancel := range pr {
			if sha != headSha {
				cancel()
			}
		}
	}
}

// waitForPrSuccessStatuses checks if all statuses are in a success state.
// Timeout limits time waiting for statuses became a success.
// TODO (dekiel): Refactor to return bool and error.
func (hb *HandlerBackend) waitForPrSuccessStatuses(ctx context.Context, logger *zap.SugaredLogger, prOrg, prRepo, prHeadSha string, prNumber int) error {
	defer logger.Sync()
	// Sleep for 30 seconds to make sure all statuses are registered.
	logger.Debugf("Sleeping for %d seconds to make sure all statuses are registered", hb.WaitForContextsCreationTimeout)
	time.Sleep(time.Duration(hb.WaitForContextsCreationTimeout) * time.Second)

	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = time.Duration(hb.WaitForStatusesTimeout) * time.Second
	backOff.MaxInterval = 10 * time.Minute
	backOff.InitialInterval = 5 * time.Minute
	logger.Debugf("Waiting for statuses to become success. Timeout: %d", hb.WaitForStatusesTimeout)

	// Check if context canceled in function to not process PR if it was canceled.
	err := backoff.Retry(func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(ctx.Err())
		default:
			defer logger.Sync()
			prStatuses, err := hb.Ghc.GetCombinedStatus(prOrg, prRepo, prHeadSha)
			if err != nil {
				gherr := fmt.Errorf("failed get pull request contexts combined status, got error %w", err)
				logger.Error(gherr.Error())
				return gherr
			}
			// Don't check if pr checks status is success as that means all context are success.
			// That means a pr was already approved and is ready for merge.
			// when pr is ready for merge.
			logger.Debugf("Pull request %d status: %s", prNumber, prStatuses.State)
			switch prState := prStatuses.State; prState {
			case "failure":
				return backoff.Permanent(fmt.Errorf("pull request %d is in failure state, skip approving", prNumber))
			case "pending":
				logger.Infof("Pull request %d is in pending state, wait for statuses to become success.", prNumber)
				for _, prStatus := range prStatuses.Statuses {
					if prStatus.State == "failure" {
						return backoff.Permanent(fmt.Errorf("pull request status check %s failed", prStatus.Context))
					} else if prStatus.State == "pending" {
						statusErr := fmt.Errorf("pull request status check %s is pending", prStatus.Context)
						logger.Debug(statusErr.Error())
						return statusErr
					}
				}
			}
			logger.Debugf("All statuses are success")
			return nil
		}
	}, backOff)
	return err
}
