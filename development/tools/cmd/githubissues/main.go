package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

var (
	githubOrgName  = flag.String("githubOrgName", "", "Github organization name [Required]")
	githubToken    = flag.String("githubToken", "", "Github token [Required]")
	issuesFilename = flag.String("issuesFilename", "issues.json", "name of the JSON file containign all issues [optional]")
	bqCredentials  = flag.String("bqCredentials", "", "Path to BigQuery credentials file [required]")
	bqProjectID    = flag.String("bqProjectID", "", "BigQuery project ID [required]")
	bqDatasetName  = flag.String("bqDataset", "", "BigQuery dataset name [required]")
	bqTableName    = flag.String("bqTable", "issues", "BigQuery table name [optional]")
)

func main() {
	flag.Parse()

	if *githubOrgName == "" {
		fmt.Fprintln(os.Stderr, "missing -githubOrgName flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubToken == "" {
		fmt.Fprintln(os.Stderr, "missing -githubToken flag")
		flag.Usage()
		os.Exit(2)
	}

	if *bqCredentials == "" {
		fmt.Fprintln(os.Stderr, "missing -bqCredentials flag")
		flag.Usage()
		os.Exit(2)
	}

	if *bqProjectID == "" {
		fmt.Fprintln(os.Stderr, "missing -bqProjectID flag")
		flag.Usage()
		os.Exit(2)
	}

	if *bqDatasetName == "" {
		fmt.Fprintln(os.Stderr, "missing -bqDataset flag")
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()

	fmt.Println("Authenticating to Github")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)
	listOptions := &github.IssueListOptions{Filter: "all", State: "open", ListOptions: github.ListOptions{PerPage: 100}}

	fmt.Println("Receiving list of issues")
	var allIssues []*github.Issue
	for {
		issues, response, err := ghClient.Issues.ListByOrg(ctx, *githubOrgName, listOptions)
		if err != nil {
			fmt.Printf("Github issues: %v", err)
			os.Exit(1)
		}
		allIssues = append(allIssues, issues...)

		if response.NextPage == 0 {
			break
		}
		listOptions.Page = response.NextPage
	}

	fmt.Printf("Saving %d issues to \"%v\"\n", len(allIssues), *issuesFilename)
	issuesFile, err := os.OpenFile(*issuesFilename, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("could nor open file: %v", err)
		os.Exit(1)
	}
	defer issuesFile.Close()

	for _, issue := range allIssues {
		// BigQuery won't process fields "+1" and "-1", removing reactions is easiest way to "fix" this
		issue.Reactions = nil
		marshalledIssue, err := json.Marshal(issue)
		if err != nil {
			fmt.Printf("Could not marshall issue: %v\n", err)
			os.Exit(1)
		}
		issuesFile.Write(marshalledIssue)
		issuesFile.WriteString("\n")
	}

	// let's re-use this reader for BigQuery
	issuesFile.Seek(0, 0)

	//bigquery
	fmt.Println("Authenticating to BigQuery")
	bqClient, err := bigquery.NewClient(ctx, *bqProjectID, option.WithCredentialsFile(*bqCredentials))
	if err != nil {
		fmt.Printf("bigquery.NewClient error: %v\n", err)
		os.Exit(1)
	}
	defer bqClient.Close()

	fmt.Printf("Deleting table \"%v:%v.%v\"\n", *bqProjectID, *bqDatasetName, *bqTableName)
	bqTable := bqClient.Dataset(*bqDatasetName).Table(*bqTableName)
	if err := bqTable.Delete(ctx); err != nil {
		fmt.Printf("BigQuery: could not delete table \"%v:%v.%v\", skipping deletion\n", *bqProjectID, *bqDatasetName, *bqTableName)
	}

	fmt.Printf("Pushing data to table \"%v:%v.%v\"\n", *bqProjectID, *bqDatasetName, *bqTableName)
	dataFile := bigquery.NewReaderSource(issuesFile)
	dataFile.SourceFormat = bigquery.JSON
	dataFile.AutoDetect = true
	loader := bqTable.LoaderFrom(dataFile)

	job, err := loader.Run(ctx)
	if err != nil {
		fmt.Printf("BigQuery: could not add records: %v\n", err)
		os.Exit(1)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		fmt.Printf("BigQuery: could not add records: %v\n", err)
		os.Exit(1)
	}

	if err := status.Err(); err != nil {
		fmt.Printf("BigQuery: could not add records: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Data was added succesfully")
}
