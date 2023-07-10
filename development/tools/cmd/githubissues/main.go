package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	githubOrgName  = flag.String("githubOrgName", "", "Github organization name [Required]")
	githubRepoName = flag.String("githubRepoName", "", "Github repository name [Optional]")
	githubToken    = flag.String("githubToken", "", "Github token [Required]")
	githubBaseURL  = flag.String("githubBaseURL", "", "Custom Github API base URL [Optional]")
	issuesFilename = flag.String("issuesFilename", "issues.json", "name of the JSON file containign all issues [Optional]")
	bqCredentials  = flag.String("bqCredentials", "", "Path to BigQuery credentials file [Required]")
	bqProjectID    = flag.String("bqProjectID", "", "BigQuery project ID [Required]")
	bqDatasetName  = flag.String("bqDataset", "", "BigQuery dataset name [Required]")
	bqTableName    = flag.String("bqTable", "issues", "BigQuery table name [Required]")
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
	var err error
	var ghClient *github.Client
	var updatedSince time.Time
	tableExists := true

	fmt.Println("Authenticating to BigQuery")
	bqClient, err := bigquery.NewClient(ctx, *bqProjectID, option.WithCredentialsFile(*bqCredentials))
	if err != nil {
		fmt.Printf("bigquery.NewClient error: %v\n", err)
		os.Exit(1)
	}
	defer bqClient.Close()

	fmt.Println("Authenticating to Github")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: *githubToken,
			TokenType:   "token",
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	if *githubBaseURL == "" {
		ghClient = github.NewClient(tc)
	} else {
		ghClient, err = github.NewEnterpriseClient(*githubBaseURL, *githubBaseURL, tc)
		if err != nil {
			fmt.Printf("Github enterprise: %v", err)
			os.Exit(1)
		}
	}

	// check if data in table exists, if yes, then get last issue
	fmt.Println("Checking if data in table exists")
	bqTable := bqClient.Dataset(*bqDatasetName).Table(*bqTableName)
	_, err = bqTable.Metadata(ctx)
	if err != nil {
		tableExists = false
	}

	var listOptions *github.IssueListOptions
	var IssueListByRepoOptions *github.IssueListByRepoOptions

	if *githubRepoName == "" {
		listOptions = &github.IssueListOptions{Filter: "all", State: "all", ListOptions: github.ListOptions{PerPage: 100}}
	} else {
		IssueListByRepoOptions = &github.IssueListByRepoOptions{State: "all", ListOptions: github.ListOptions{PerPage: 100}}
	}

	if tableExists {
		fmt.Println("Getting most recently updated issue from the table")
		// get time of the last indexed issue and search for newer ones

		q := bqClient.Query("SELECT updated_at FROM `" + *bqProjectID + "." + *bqDatasetName + "." + *bqTableName + "` ORDER BY updated_at DESC LIMIT 1")
		q.Location = "US"

		// 2021-05-14 11:12:06 UTC
		job, err := q.Run(ctx)
		if err != nil {
			fmt.Printf("Bq issues: %v", err)
			os.Exit(1)
		}
		status, err := job.Wait(ctx)
		if err != nil {
			fmt.Printf("Bq issues: %v", err)
			os.Exit(1)
		}
		if err := status.Err(); err != nil {
			fmt.Printf("Bq issues: %v", err)
			os.Exit(1)
		}
		it, err := job.Read(ctx)

		if err != nil {
			fmt.Printf("Bq issues: %v", err)
			os.Exit(1)
		}
		for {
			var row []bigquery.Value
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Printf("Bigquery error: %v", err)
				os.Exit(1)
			}
			updatedSince = row[0].(time.Time).Add(time.Second)
			fmt.Printf("Will be looking for issues updated after %v\n", updatedSince)
			if *githubRepoName == "" {
				listOptions.Since = updatedSince
			} else {
				IssueListByRepoOptions.Since = updatedSince
			}
		}
	}

	fmt.Println("Receiving list of issues")
	var allIssues []*github.Issue
	for {
		var issues []*github.Issue
		var response *github.Response
		if *githubRepoName == "" {
			issues, response, err = ghClient.Issues.ListByOrg(ctx, *githubOrgName, listOptions)
		} else {
			issues, response, err = ghClient.Issues.ListByRepo(ctx, *githubOrgName, *githubRepoName, IssueListByRepoOptions)
		}
		if err != nil {
			fmt.Printf("Github issues: %v", err)
			os.Exit(1)
		}
		allIssues = append(allIssues, issues...)

		if response.NextPage == 0 {
			break
		}

		if *githubRepoName == "" {
			listOptions.Page = response.NextPage
		} else {
			IssueListByRepoOptions.Page = response.NextPage
		}
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

	// load all rows

	fmt.Printf("Pushing data to table \"%v:%v.%v\"\n", *bqProjectID, *bqDatasetName, *bqTableName)
	dataFile := bigquery.NewReaderSource(issuesFile)
	dataFile.SourceFormat = bigquery.JSON
	// dataFile.AutoDetect = true
	loader := bqTable.LoaderFrom(dataFile)

	job, err := loader.Run(ctx)
	if err != nil {
		fmt.Printf("BigQuery: could not add records: %v\n", err)
		for _, errMessge := range job.LastStatus().Errors {
			fmt.Printf("%s\n", errMessge.Message)
		}
		os.Exit(1)
	}

	status, err := job.Wait(ctx)
	if err != nil {
		fmt.Printf("BigQuery: could not add records: %v\n", err)
		for _, errMessge := range job.LastStatus().Errors {
			fmt.Printf("%s\n", errMessge.Message)
		}
		os.Exit(1)
	}

	if err := status.Err(); err != nil {
		fmt.Printf("BigQuery: could not add records: %v\n", err)
		for _, errMessge := range job.LastStatus().Errors {
			fmt.Printf("%s\n", errMessge.Message)
		}
		os.Exit(1)
	}

	// dedupe
	fmt.Println("removing duplicates")
	q := bqClient.Query("DELETE FROM `" + *bqProjectID + "." + *bqDatasetName + "." + *bqTableName + "` WHERE STRUCT(id, updated_at) NOT IN ( SELECT AS STRUCT id, MAX(updated_at) updated_at FROM `" + *bqProjectID + "." + *bqDatasetName + "." + *bqTableName + "` GROUP BY id)")
	q.Location = "US"

	// 2021-05-14 11:12:06 UTC
	job, err = q.Run(ctx)
	if err != nil {
		fmt.Printf("Bq issues: %v", err)
		os.Exit(1)
	}
	status, err = job.Wait(ctx)
	if err != nil {
		fmt.Printf("Bq issues: %v", err)
		os.Exit(1)
	}
	if err := status.Err(); err != nil {
		fmt.Printf("Bq issues: %v", err)
		os.Exit(1)
	}
	fmt.Println("Data was added successfully")
}
