package main

import (
	"context"

	"go.uber.org/zap"
	"sigs.k8s.io/prow/pkg/github"
)

func (hb *HandlerBackend) handleReviewRequestedAction(ctx context.Context, cancel context.CancelFunc, logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	if locked := hb.lockPR(cancel, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number); !locked {
		logger.Infof("Review request for pull request head sha %s already in process.", prEvent.PullRequest.Head.SHA)
		return
	}
	logger.Debug("Got pull request review requested action")
	logger.Sync()
	hb.reviewPullRequest(ctx, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.User.Login, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number, prEvent.PullRequest.Labels)
}

func (hb *HandlerBackend) handlePrSynchronizeAction(ctx context.Context, cancel context.CancelFunc, logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	// Cancel context for review for previous commit.
	hb.cancelPR(logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number)
	if locked := hb.lockPR(cancel, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number); !locked {
		logger.Infof("Pull request head sha %s already in process.", prEvent.PullRequest.Head.SHA)
		return
	}
	logger.Debug("Got pull request synchronize action")
	logger.Sync()
	hb.reviewPullRequest(ctx, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.User.Login, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number, prEvent.PullRequest.Labels)
}

func (hb *HandlerBackend) handleReviewDismissedAction(ctx context.Context, cancel context.CancelFunc, logger *zap.SugaredLogger, reviewEvent github.ReviewEvent) {
	if locked := hb.lockPR(cancel, logger, reviewEvent.Repo.Owner.Login, reviewEvent.Repo.Name, reviewEvent.PullRequest.Head.SHA, reviewEvent.PullRequest.Number); !locked {
		logger.Infof("Pull request head sha %s already in process.", reviewEvent.PullRequest.Head.SHA)
		return
	}
	logger.Debug("Got pull request review dismissed action")
	logger.Sync()
	hb.reviewPullRequest(ctx, logger, reviewEvent.Repo.Owner.Login, reviewEvent.Repo.Name, reviewEvent.PullRequest.User.Login, reviewEvent.PullRequest.Head.SHA, reviewEvent.PullRequest.Number, reviewEvent.PullRequest.Labels)
}

// handlePrOpenedAction function executes logic required for GitHub pull request opened event.
// TODO (dekiel): do we need lock or cancel here?
func (hb *HandlerBackend) handlePrOpenedAction(ctx context.Context, cancel context.CancelFunc, logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	if locked := hb.lockPR(cancel, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number); !locked {
		logger.Infof("Pull request head sha %s already in process.", prEvent.PullRequest.Head.SHA)
		return
	}
	logger.Debug("Got pull request opened action")
	logger.Sync()
	err := hb.setPullRequestAutoMerge(ctx, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number, prEvent.PullRequest.Labels)
	if err != nil {
		logger.Errorf("Failed to set pull request auto merge: %v", err)
	}
}

// handlePrOpenedAction function executes logic required for GitHub pull request opened event.
// TODO (dekiel): do we need lock or cancel here?
func (hb *HandlerBackend) handlePrLabeledAction(ctx context.Context, cancel context.CancelFunc, logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	if locked := hb.lockPR(cancel, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number); !locked {
		logger.Infof("Pull request head sha %s already in process.", prEvent.PullRequest.Head.SHA)
		return
	}
	logger.Debug("Got pull request opened action")
	logger.Sync()
	err := hb.setPullRequestAutoMerge(ctx, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number, prEvent.PullRequest.Labels)
	if err != nil {
		logger.Errorf("Failed to set pull request auto merge: %v", err)
	}
}

// handlePrOpenedAction function executes logic required for GitHub pull request opened event.
// TODO (dekiel): do we need lock or cancel here?
func (hb *HandlerBackend) handlePrUnlabeledAction(ctx context.Context, cancel context.CancelFunc, logger *zap.SugaredLogger, prEvent github.PullRequestEvent) {
	if locked := hb.lockPR(cancel, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number); !locked {
		logger.Infof("Pull request head sha %s already in process.", prEvent.PullRequest.Head.SHA)
		return
	}
	logger.Debug("Got pull request opened action")
	logger.Sync()
	err := hb.setPullRequestAutoMerge(ctx, logger, prEvent.Repo.Owner.Login, prEvent.Repo.Name, prEvent.PullRequest.Head.SHA, prEvent.PullRequest.Number, prEvent.PullRequest.Labels)
	if err != nil {
		logger.Errorf("Failed to set pull request auto merge: %v", err)
	}
}
