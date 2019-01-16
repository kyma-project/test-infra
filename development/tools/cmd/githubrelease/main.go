// See https://cloud.google.com/docs/authentication/.
// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
// a service account key file to authenticate to the API.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/test-infra/development/tools/pkg/common"
	"github.com/kyma-project/test-infra/development/tools/pkg/file"
	"github.com/kyma-project/test-infra/development/tools/pkg/release"
	log "github.com/sirupsen/logrus"
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

	common.Shout("Reading release version file")

	releaseVersion, err := file.ReadFile(*releaseVersionFilePath)
	if err != nil {
		log.Fatal(err)
	}

	isPreRelease := strings.Contains(releaseVersion, "rc")

	common.Shout("Release version: %s, Pre-release: %t", releaseVersion, isPreRelease)

	ga := release.NewGithubAPI(ctx, *githubAccessToken, *githubRepoOwner, *githubRepoName)

	sa, err := release.NewStorageAPI(ctx, *bucketName, releaseVersion)
	if err != nil {
		log.Fatal(err)
	}

	c := release.NewCreator(ga, sa)

	// Github release
	err = c.CreateNewRelease(ctx, releaseVersion, *targetCommit, *kymaChangelog, *kymaConfigLocal, *kymaConfigCluster, isPreRelease)
	if err != nil {
		log.Fatal(err)
	}
}
