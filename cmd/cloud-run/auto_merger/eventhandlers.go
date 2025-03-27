package main

import (
	"context"
	"encoding/json"

	consolelog "github.com/kyma-project/test-infra/pkg/logging"
	"github.com/kyma-project/test-infra/pkg/prow/externalplugin"
	"sigs.k8s.io/prow/pkg/github"
)

// PullRequestEventHandler handles pull_request events. It checks event action and calls the appropriate handler function.
// TODO: All actions should be handled in one handler function. The event type is passed in payload.
//
//	Based on event type, the handler function should use appropriate event struct.
//	That way we can avoid code duplication.
func (hb *HandlerBackend) PullRequestEventHandler(_ *externalplugin.Plugin, payload externalplugin.Event) {
	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()
	atom.SetLevel(hb.LogLevel)
	logger = logger.With(externalplugin.EventTypeField, payload.EventType, github.EventGUID, payload.EventGUID)

	logger.Debug("Got pull_request payload")
	var prEvent github.PullRequestEvent
	if err := json.Unmarshal(payload.Payload, &prEvent); err != nil {
		logger.Errorw("Failed unmarshal json payload.", "error", err)
		return
	}
	logger = logger.With("pr-number", prEvent.Number)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	switch prEvent.Action {
	case github.PullRequestActionReviewRequested:
		hb.handleReviewRequestedAction(ctx, cancel, logger, prEvent)
	case github.PullRequestActionSynchronize:
		hb.handlePrSynchronizeAction(ctx, cancel, logger, prEvent)
	case github.PullRequestActionOpened:
		hb.handlePrOpenedAction(ctx, cancel, logger, prEvent)
	case github.PullRequestActionLabeled:
		hb.handlePrLabeledAction(ctx, cancel, logger, prEvent)
	case github.PullRequestActionUnlabeled:
		hb.handlePrUnlabeledAction(ctx, cancel, logger, prEvent)
	}
}

// PullRequestReviewEventHandler handles pull_request_review events. It checks event action and calls the appropriate handler function.
func (hb *HandlerBackend) PullRequestReviewEventHandler(_ *externalplugin.Plugin, payload externalplugin.Event) {
	logger, atom := consolelog.NewLoggerWithLevel()
	defer logger.Sync()
	atom.SetLevel(hb.LogLevel)
	logger = logger.With(externalplugin.EventTypeField, payload.EventType, github.EventGUID, payload.EventGUID)

	logger.Debug("Got pull_request_review payload")
	var reviewEvent github.ReviewEvent
	if err := json.Unmarshal(payload.Payload, &reviewEvent); err != nil {
		logger.Errorw("Failed unmarshal json payload.", "error", err)
		return
	}
	logger = logger.With("pr-number", reviewEvent.PullRequest.Number)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	switch reviewEvent.Action {
	case github.ReviewActionDismissed:
		hb.handleReviewDismissedAction(ctx, cancel, logger, reviewEvent)
	}
}
