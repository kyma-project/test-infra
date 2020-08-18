package prtagbuilder

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"k8s.io/test-infra/prow/config/secret"

	"github.com/go-yaml/yaml"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	prowflagutil "k8s.io/test-infra/prow/flagutil"
	"k8s.io/test-infra/prow/github"
)

func BuildPrTag(ghOptions prowflagutil.GitHubOptions) {
	var secretAgent *secret.Agent
	if ghOptions.TokenPath != "" {
		secretAgent = &secret.Agent{}
		if err := secretAgent.Start([]string{ghOptions.TokenPath}); err != nil {
			logrus.WithError(err).Fatal("Failed to start secret agent")
		}
	}
	githubClient, err = o.github.GitHubClient(secretAgent, false)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get GitHub client")
	}
}
