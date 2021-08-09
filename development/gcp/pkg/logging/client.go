package logging

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"google.golang.org/api/option"
)

const (
	errorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
	LogsProjectID      = "sap-kyma-prow"
)

func NewClient(ctx context.Context, credentialsFilePath, logName string) (*Client, error) {
	client := &Client{
		LogName: logName,
	}
	c, err := logging.NewClient(ctx, LogsProjectID, option.WithCredentialsFile(credentialsFilePath))
	if err != nil {
		return nil, fmt.Errorf("got error while creating google cloud logging client, error: %v", err)
	}
	client.Client = c
	return client, nil
}

func (c *Client) NewProwjobLogger() *logging.Logger {
	logger := c.Logger(c.LogName, logging.CommonLabels())
}
