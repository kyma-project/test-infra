// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/google/go-github/github"
	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/file"
	"github.com/kyma-project/test-infra/development/tools/pkg/githubrelease"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	targetCommit           = flag.String("targetCommit", "", "Target commitish [Required]")
	bucketName             = flag.String("bucketName", "kyma-prow-artifacts", "Google bucket name where artifacts are stored [Optional]")
	kymaConfigCluster      = flag.String("kymaConfigCluster", "kyma-config-cluster.yaml", "Filename for cluster artifacts [Optional]")
	kymaConfigLocal        = flag.String("kymaConfigLocal", "kyma-config-local.yaml", "Filename for local artifacts [Optional]")
	kymaChangelog          = flag.String("kymaChangelog", "release-changelog.md", "Filename for release changelog [Optional]")
	githubRepoOwner        = flag.String("githubRepoOwner", "", "Github repository owner [Required]")
	githubRepoName         = flag.String("githubRepoName", "", "Github repository name [Required]")
	githubAccessToken      = flag.String("githubAccessToken", "", "Github access token [Required]")
	releaseVersionFilePath = flag.String("releaseVersionFilePath", "", "Full path to a file containing release version [Required]")
	kymaArtifactsDir       = "kyma-artifacts"
)

func main() {
	flag.Parse()

	if *targetCommit == "" {
		fmt.Fprintln(os.Stderr, "missing -targetCommit flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubRepoOwner == "" {
		fmt.Fprintln(os.Stderr, "missing -githubRepoOwner flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubRepoName == "" {
		fmt.Fprintln(os.Stderr, "missing -githubRepoName flag")
		flag.Usage()
		os.Exit(2)
	}

	if *githubAccessToken == "" {
		fmt.Fprintln(os.Stderr, "missing -githubAccessToken flag")
		flag.Usage()
		os.Exit(2)
	}

	if *releaseVersionFilePath == "" {
		fmt.Fprintln(os.Stderr, "missing -releaseVersionFilePath flag")
		flag.Usage()
		os.Exit(2)
	}

	ctx := context.Background()

	artifactsDir, err := ioutil.TempDir("", kymaArtifactsDir)
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(artifactsDir)

	common.Shout("Reading release version file")

	releaseVersion, err := file.ReadFile(*releaseVersionFilePath)
	if err != nil {
		log.Fatal(err)
	}

	isPreRelease := strings.Contains(releaseVersion, "rc")

	common.Shout("Release version: %s, Pre-release: %b", releaseVersion, isPreRelease)

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	saw := &githubrelease.StorageAPIWrapper{Context: ctx, StorageClient: storageClient, BucketName: *bucketName, FolderName: releaseVersion, TmpDir: artifactsDir}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	gap := &githubrelease.GithubAPIWrapper{Context: ctx, Client: client, RepoOwner: *githubRepoOwner, RepoName: *githubRepoName}

	gr := &githubrelease.Release{Gap: gap, Saw: saw}

	// Github Release
	err = gr.CreateRelease(releaseVersion, *targetCommit, *kymaChangelog, *kymaConfigLocal, *kymaConfigCluster, isPreRelease)
	if err != nil {
		log.Fatal(err)
	}
}
