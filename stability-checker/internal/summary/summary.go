package summary

import (
	"encoding/json"
	"io"

	"bufio"

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
}

// NewService returns Service
func NewService(logFetcher logFetcher, processor logProcessor) *Service {
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

	br := bufio.NewReader(readCloser)

	testIDMap := map[string]struct{}{}
	for _, id := range testIDs {
		testIDMap[id] = struct{}{}
	}
	aggregated := newStatsAggregator()

	end := false
	for !end {
		line, err := br.ReadBytes('\n')
		switch err {
		case nil:
		case io.EOF:
			end = true
		default:
			return nil, errors.Wrap(err, "while scanning stream")
		}

		var e log.Entry
		err = json.Unmarshal(line, &e)
		if err != nil {
			continue
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

	// create a buffered reader to skip non json part of the log
	//	input := bufio.NewReader(readCloser)
	//	stream := json.NewDecoder(input)
	//
	//	testIDMap := map[string]struct{}{}
	//	for _, id := range testIDs {
	//		testIDMap[id] = struct{}{}
	//	}
	//
	//	aggregated := newStatsAggregator()
	//loop:
	//	for {
	//		var e log.Entry
	//		switch err := stream.Decode(&e); err {
	//		case nil:
	//		case io.EOF:
	//			break loop
	//		default:
	//			e := c.skipNonJSON(input)
	//			switch e {
	//			case nil:
	//				stream = json.NewDecoder(input)
	//				stream.Buffered()
	//			case io.EOF:
	//				break loop
	//			default:
	//				return nil, errors.Wrap(err, "while decoding stream")
	//			}
	//			continue
	//		}
	//
	//		_, contains := testIDMap[e.Log.TestRunID]
	//		if contains {
	//			tm, err := c.processor.Process([]byte(e.Log.Message))
	//			if err != nil {
	//				return nil, errors.Wrap(err, "while processing test output")
	//			}
	//			aggregated.Merge(tm)
	//		}
	//
	//	}

	return aggregated.ToList(), nil
}

func (c *Service) skipNonJSON(br *bufio.Reader) error {

	for {
		// read characters until an error or beginning of a new json
		ch, _, err := br.ReadRune()
		if err != nil {
			return err
		}
		if ch == '{' {
			err := br.UnreadRune()
			return err
		}
	}
}
