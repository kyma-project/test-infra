package summary

import (
	"encoding/json"
	"io"

	"github.com/kyma-project/test-infra/stability-checker/internal/log"
	"github.com/pkg/errors"
)

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=LogProcessor -output=automock -outpkg=automock -case=underscore

// LogProcessor interface to process logs
type LogProcessor interface {
	Process([]byte) (map[string]SpecificTestStats, error)
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=LogFetcher -output=automock -outpkg=automock -case=underscore

// LogFetcher interface to receive logs from pod
type LogFetcher interface {
	GetLogsFromPod() (io.ReadCloser, error)
}

// Service is responsible for producing summary for test executions.
type Service struct {
	logFetcher LogFetcher
	processor  LogProcessor
}

// NewService returns Service
func NewService(logFetcher LogFetcher, processor LogProcessor) *Service {
	return &Service{
		logFetcher: logFetcher,
		processor:  processor,
	}
}

// GetTestSummaryForExecutions analyzes logs from test executions and produces summary for specific tests.
func (c *Service) GetTestSummaryForExecutions(testIDs []string) ([]SpecificTestStats, error) {
	readCloser, err := c.logFetcher.GetLogsFromPod()
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()
	stream := json.NewDecoder(readCloser)

	testIDMap := map[string]struct{}{}
	for _, id := range testIDs {
		testIDMap[id] = struct{}{}
	}

	aggregated := newStatsAggregator()
loop:
	for {
		var e log.Entry
		switch err := stream.Decode(&e); err {
		case nil:
		case io.EOF:
			break loop
		default:
			return nil, errors.Wrap(err, "while decoding stream")
		}

		_, contains := testIDMap[e.Log.TestRunID]
		if contains {
			tm, err := c.processor.Process([]byte(e.Log.Message))
			if err != nil {
				return nil, errors.Wrap(err, "while processing test output")
			}
			aggregated.Merge(tm)
		}

	}

	return aggregated.ToList(), nil
}
