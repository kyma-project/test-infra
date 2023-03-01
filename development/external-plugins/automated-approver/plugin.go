package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"

	consolelog "github.com/kyma-project/test-infra/development/logging"
	"github.com/kyma-project/test-infra/development/prow/externalplugin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/test-infra/prow/github"
)

type githubClient interface {
	CreatePullRequestReviewComment(org, repo string, number int, rc github.ReviewComment) error
	CreateReview(org, repo string, number int, r github.DraftReview) error
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
	GetCombinedStatus(org, repo, ref string) (*github.CombinedStatus, error)
	AddLabel(org, repo string, number int, label string) error
	CreateComment(org, repo string, number int, comment string) error
}

// handlerBackend is a backend for the plugin.
// It contains all the configuration and clients needed to handle events.
type handlerBackend struct {
	ghc                    githubClient
	logLevel               zapcore.Level
	waitForStatusesTimeout int                                                 // in seconds
	configPath             string                                              // Path to yaml config file
	conditions             map[string]map[string]map[string][]ApproveCondition `yaml:"conditions"`
	prLocks                map[string]struct{}                                 // Holds PRs head sha that are being processed.
	prMutex                sync.Mutex
}

// WatchConfig watches for changes in config file and reloads it.
func (hb *handlerBackend) watchConfig(logger *zap.SugaredLogger) {
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
					err := hb.readConfig()
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

	err = watcher.Add(hb.configPath)
	if err != nil {
		logger.Fatalf("Add failed: %s", err)
	}
	<-done
}

// lockPR locks PR for processing by adding head sha to prLocks.
// If PR is already locked, returns false.
// Because GitHub sends multiple review request events for one PR, we need to lock PR to avoid processing it multiple times.
// GitHub sends multiple events because it sends one event for each reviewer.
func (hb *handlerBackend) lockPR(logger *zap.SugaredLogger, headSha string) bool {
	// Sync access to prLocks with mutex.
	hb.prMutex.Lock()
	defer hb.prMutex.Unlock()
	defer logger.Sync()
	_, ok := hb.prLocks[headSha]
	if !ok {
		if hb.prLocks == nil {
			hb.prLocks = make(map[string]struct{})
		}
		hb.prLocks[headSha] = struct{}{}
		return true
	}
	return false
}

// unlockPR unlocks PR by removing head sha from prLocks.
func (hb *handlerBackend) unlockPR(logger *zap.SugaredLogger, headSha string) {
	// Sync access to prLocks with mutex.
	hb.prMutex.Lock()
	defer hb.prMutex.Unlock()
	defer logger.Sync()
	delete(hb.prLocks, headSha)
}

// ApproveCondition defines conditions for approving PR.
type ApproveCondition struct {
	RequiredLabels []string `yaml:"requiredLabels"`
	ChangedFiles   []string `yaml:"changedFiles"`
}

// String returns string representation of ApproveCondition.
func (ac *ApproveCondition) String() string {
	b, _ := json.Marshal(ac)
	return string(b)
}

// checkRequiredLabels checks if PR has all required labels.
func (ac *ApproveCondition) checkRequiredLabels(logger *zap.SugaredLogger, prLabels []github.Label) bool {
	defer logger.Sync()
	if ac.RequiredLabels == nil {
		logger.Debug("No required labels defined")
		// No required labels defined
		return true
	}
	pl := make(map[string]interface{})
	logger.Debugf("Checking if PR has all required labels: %v", ac.RequiredLabels)
	for _, l := range prLabels {
		pl[l.Name] = nil
	}
	for _, requiredLabel := range ac.RequiredLabels {
		if _, ok := pl[requiredLabel]; !ok {
			logger.Debugf("PR is missing required label: %s", requiredLabel)
			return false
		}
	}
	logger.Debug("All required labels are present")
	return true
}

// checkChangedFiles checks if PR changed only allowed files.
func (ac *ApproveCondition) checkChangedFiles(logger *zap.SugaredLogger, changes []github.PullRequestChange) bool {
	defer logger.Sync()
	logger.Debugf("Checking if PR changed only allowed files: %v", ac.ChangedFiles)
	for _, change := range changes {
		change := change
		logger.Debugf("Checking file: %s", change.Filename)
		matched := slices.ContainsFunc(ac.ChangedFiles, func(allowedFile string) bool {
			filesMatcher := regexp.MustCompile(allowedFile)
			matched := filesMatcher.MatchString(change.Filename)
			logger.Debugf("File %s matched %s: %t", change.Filename, allowedFile, matched)
			return matched
		})
		if !matched {
			logger.Debugf("File %s not matched", change.Filename)
			return false
		}
	}
	logger.Debug("All files matched")
	return true
}

// readConfig reads config from config file.
func (hb *handlerBackend) readConfig() error {
	c := make(map[string]map[string]map[string]map[string][]ApproveCondition)
	configFile, err := os.ReadFile(hb.configPath)
	if err == nil {
		yaml.Unmarshal(configFile, &c)
		hb.conditions = c["conditions"]
		return nil
	}
	return err
}

// checkPrStatuses checks if all statuses are in success state.
// Tide required status check is not taken into account. It will be always pending until PR is ready to merge.
// Timeout limits time waiting for statuses became success.
func (hb *handlerBackend) checkPrStatuses(logger *zap.SugaredLogger, prEvent github.PullRequestEvent) error {
	defer logger.Sync()
	// Sleep for 30 seconds to make sure all statuses are registered.
	logger.Debug("Sleeping for 30 seconds to make sure all statuses are registered")
	time.Sleep(30 * time.Second)

	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = time.Duration(hb.waitForStatusesTimeout) * time.Second
	backOff.MaxInterval = 10 * time.Minute
	backOff.InitialInterval = 5 * time.Minute
	logger.Debugf("Waiting for statuses to become success. Timeout: %d", hb.waitForStatusesTimeout)

	err := backoff.Retry(func() error {
		defer logger.Sync()
		prStatuses, err := hb.ghc.GetCombinedStatus(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA)
		if err != nil {
			gherr := fmt.Errorf("failed get pull request contexts combined status, got error %w", err)
			logger.Error(gherr.Error())
			return gherr
		}
		// logger.Sync() // Syncing logger to make sure all logs from calling GitHub API are written before logs from functions called in next steps.

		// Don't check if pr checks status is success as that means all context are success, even tide context.
		// That means a pr was already approved and is ready for merge, because tide context transition to success
		// when pr is ready for merge.
		logger.Debugf("Pull request %d status: %s", prEvent.Number, prStatuses.State)
		switch prState := prStatuses.State; prState {
		case "failure":
			return backoff.Permanent(fmt.Errorf("pull request %d is in failure state, skip approving", prEvent.Number))
		case "pending":
			logger.Infof("Pull request %d is in pending state, wait for statuses to become success.", prEvent.Number)
			for _, prStatus := range prStatuses.Statuses {
				if prStatus.State == "failure" {
					return backoff.Permanent(fmt.Errorf("pull request status check %s failed", prStatus.Context))
				} else if prStatus.State == "pending" && prStatus.Context != "tide" {
					statusErr := fmt.Errorf("pull request status check %s is pending", prStatus.Context)
					logger.Debug(statusErr.Error())
					return statusErr
				}
			}
		}
		logger.Debugf("All statuses are success")
		return nil
	}, backOff)
	return err
}

func (hb *handlerBackend) checkPrApproveConditions(logger *zap.SugaredLogger, conditions []ApproveCondition, changes []github.PullRequestChange, prLabels []github.Label) bool {
	defer logger.Sync()
	for _, condition := range conditions {
		logger.Debugw("Checking condition", "condition", condition)
		labelsMatched := condition.checkRequiredLabels(logger, prLabels)
		if !labelsMatched {
			logger.Debug("Labels not matched")
			continue
		}
		filesMatched := condition.checkChangedFiles(logger, changes)
		if !filesMatched {
			logger.Debug("Files not matched")
			continue
		}
		return true
	}
	logger.Debug("No conditions matched")
	return false
}

func (hb *handlerBackend) handleReviewRequestedAction(logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	defer logger.Sync()
	defer hb.unlockPR(logger, prEvent.PullRequest.Head.SHA)
	logger.Debugf("Checking if conditions for PR author %s exists: %t", prEvent.PullRequest.User.Login, hb.conditions[prEvent.Repo.Owner.Login][prEvent.Repo.Name][prEvent.PullRequest.User.Login] != nil)
	if conditions, ok := hb.conditions[prEvent.Repo.Owner.Login][prEvent.Repo.Name][prEvent.PullRequest.User.Login]; ok {
		logger.Debugf("Checking if PR %d meets approval conditions: %v", prEvent.Number, conditions)

		// Get changes from pull request.
		changes, err := hb.ghc.GetPullRequestChanges(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.Number)
		if err != nil {
			logger.Errorw("failed get pull request changes", "error", err.Error())
		}
		logger.Sync() // Syncing logger to make sure all logs from calling GitHub API are written before logs from functions called in next steps.
		conditionsMatched := hb.checkPrApproveConditions(logger, conditions, changes, prEvent.PullRequest.Labels)
		if !conditionsMatched {
			return
		}
		// Sleep for 30 seconds to make sure all statuses are registered.
		// logger.Debug("Sleeping for 30 seconds to make sure all statuses are registered")
		// time.Sleep(30 * time.Second)

		// prStatuses, err := hb.ghc.GetCombinedStatus(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA)
		// if err != nil {
		// 	logger.Errorw("failed get pull request contexts combined status", "error", err.Error())
		// }
		// logger.Sync() // Syncing logger to make sure all logs from calling GitHub API are written before logs from functions called in next steps.

		// Don't check if pr checks status is success as that means all context are success, even tide context.
		// That means a pr was already approved and is ready for merge, because tide context transition to success
		// when pr is ready for merge.
		// logger.Debugf("Pull request %d status: %s", prEvent.Number, prStatuses.State)
		// switch prState := prStatuses.State; prState {
		// case "failure":
		// 	logger.Infof("Pull request %d is in failure state, skip approving.", prEvent.Number)
		// 	return
		// case "pending":
		// 	logger.Infof("Pull request %d is in pending state, wait for statuses to become success.", prEvent.Number)
		err = hb.checkPrStatuses(logger, prEvent)
		if err != nil {
			logger.Errorf("pull request %s/%s#%d has non success statuses, got error: %s",
				prEvent.Repo.Owner.Login,
				prEvent.Repo.Name,
				prEvent.Number,
				err)
			return
		}
		review := github.DraftReview{
			CommitSHA: prEvent.PullRequest.Head.SHA,
			Body:      "",
			Action:    "APPROVE",
			Comments:  nil,
		}
		err = hb.ghc.CreateReview(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.Number, review)
		if err != nil {
			logger.Errorf("failed create review for pull request %s/%s#%d sha: %s, got error: %s",
				prEvent.Repo.Owner.Login,
				prEvent.Repo.Name,
				prEvent.Number,
				prEvent.PullRequest.Head.SHA,
				err)
			return
		}
		logger.Infof("Pull request %s/%s#%d was approved.", prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.Number)
		err = hb.ghc.AddLabel(prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.Number, "auto-approved")
		if err != nil {
			logger.Errorf("failed add label to pull request %s/%s#%d, got error: %s",
				prEvent.Repo.Owner.Login,
				prEvent.Repo.Name,
				prEvent.Number,
				err)
		}
		logger.Infof("Label auto-approved was added to pull request %s/%s#%d.", prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.Number)
	} else {
		logger.Infof("Pull request %s/%s#%d doesn't meet conditions to be auto approved, pr author %s doesn't have conditions defined.",
			prEvent.Repo.Owner.Login,
			prEvent.Repo.Name,
			prEvent.Number,
			prEvent.PullRequest.User.Login)
	}
}

func (hb *handlerBackend) pullRequestEventHandler(_ *externalplugin.Plugin, payload externalplugin.Event) {
	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()
	atom.SetLevel(hb.logLevel)
	logger = logger.With(externalplugin.EventTypeField, payload.EventType, github.EventGUID, payload.EventGUID)

	logger.Debug("Got pull_request payload")
	var prEvent github.PullRequestEvent
	if err := json.Unmarshal(payload.Payload, &prEvent); err != nil {
		logger.Errorw("Failed unmarshal json payload.", "error", err)
		return
	}

	if prEvent.Action == github.PullRequestActionReviewRequested {
		logger = logger.With("pr-number", prEvent.Number)
		if locked := hb.lockPR(logger, prEvent.PullRequest.Head.SHA); !locked {
			logger.Infof("Reeview request for pull request head sha %s already in process.", prEvent.PullRequest.Head.SHA)
			return
		}
		logger.Debug("Got pull request review requested action")
		logger.Sync()
		hb.handleReviewRequestedAction(logger, prEvent)
	}
}
