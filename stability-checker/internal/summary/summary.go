package summary

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"

	"github.com/kyma-project/test-infra/stability-checker/internal/log"
	"github.com/pkg/errors"
)

// dependencies
//go:generate mockery -name=logProcessor -output=automock -outpkg=automock -case=underscore
type logProcessor interface {
	Process([]byte) (map[string]SpecificTestStats, error)
}

//go:generate mockery -name=logFetcher -output=automock -outpkg=automock -case=underscore
type logFetcher interface {
	GetLogsFromPod() (io.ReadCloser, error)
}

// Service is responsible for producing summary for test executions.
type Service struct {
	logFetcher logFetcher
	processor  logProcessor
	log        logrus.FieldLogger
}

// NewService returns Service
func NewService(logFetcher logFetcher, processor logProcessor, log logrus.FieldLogger) *Service {
	return &Service{
		logFetcher: logFetcher,
		processor:  processor,
		log:        log,
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
