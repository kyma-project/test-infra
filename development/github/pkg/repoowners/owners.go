package repoowners

import (
	"fmt"
	"io/ioutil"

	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"

	toolstypes "github.com/kyma-project/test-infra/development/types"
	"k8s.io/test-infra/prow/github"
	k8sowners "k8s.io/test-infra/prow/repoowners"
)

type AllOwners map[string]struct{}

func (a AllOwners) addOwners(approvers []string) {
	for _, approver := range approvers {
		a[approver] = struct{}{}
	}
}

func (c *OwnersClient) LoadRepoAliases(basedir, filename string) (k8sowners.RepoAliases, error) {
	path := filepath.Join(basedir, filename)
	c.Logger.Debugf("Loading repo aliases from path %s", path)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return k8sowners.ParseAliasesConfig(b)
}

func (c *OwnersClient) GetOwnersForChanges(changes []github.PullRequestChange, repoBase string, owners k8sowners.RepoOwner) (AllOwners, error) {
	var (
		conf k8sowners.Config
		err  error
	)
	allOwners := AllOwners{}
	for _, change := range changes {
		approversFile := owners.FindApproverOwnersForFile(change.Filename)
		reviewersFile := owners.FindReviewersOwnersForFile(change.Filename)
		conf, err = c.ParseOwnersFile(approversFile, repoBase, owners)
		if err != nil {
			return nil, err
		}
		allOwners.addOwners(conf.Approvers)
		if approversFile != reviewersFile {
			conf, err = c.ParseOwnersFile(reviewersFile, repoBase, owners)
			if err != nil {
				return nil, err
			}
		}
		allOwners.addOwners(conf.Reviewers)
	}
	return allOwners, nil
}

func (c *OwnersClient) ParseOwnersFile(ownersFilePath, repoBase string, owners k8sowners.RepoOwner) (k8sowners.Config, error) {
	var (
		fullconfig   k8sowners.FullConfig
		simpleconfig k8sowners.SimpleConfig
		err          error
	)
	ownersFile := filepath.Join(repoBase, ownersFilePath, "OWNERS")

	simpleconfig, err = owners.ParseSimpleConfig(ownersFile)
	if err != nil {
		return k8sowners.Config{}, fmt.Errorf("failed parsing %s owners file to simpleconfig, error: %w", ownersFilePath, err)
	}
	if simpleconfig.Empty() {
		fullconfig, err = owners.ParseFullConfig(ownersFile)
		if err != nil {
			return k8sowners.Config{}, fmt.Errorf("failed parsing %s owners file to fullconfig, error: %w", ownersFilePath, err)
		}
		return fullconfig.Filters[".*"], nil
	}
	return simpleconfig.Config, nil
}

func (c *OwnersClient) ResolveSlackNames(allOwners AllOwners, aliases []toolstypes.Alias, users []toolstypes.User, repoAliases k8sowners.RepoAliases) (sets.String, error) {
	aliasesMap := make(map[string]toolstypes.Alias)
	usersMap := make(map[string]toolstypes.User)
	targets := sets.NewString()
	for _, alias := range aliases {
		aliasesMap[alias.ComGithubAliasname] = alias
	}
	for _, user := range users {
		usersMap[user.ComGithubUsername] = user
	}
	for owner := range allOwners {
		if !c.checkIfNotifyAlias(owner, &targets, aliasesMap) {
			if !c.checkIfNotifyUser(owner, &targets, usersMap) {
				userOwners := repoAliases.ExpandAlias(owner)
				if userOwners.Len() > 0 {
					for userOwner, _ := range userOwners {
						_ = c.checkIfNotifyUser(userOwner, &targets, usersMap)
					}
				}
			}
		}
	}
	return targets, nil
}

func (c *OwnersClient) checkIfNotifyAlias(owner string, targets *sets.String, aliasesMap map[string]toolstypes.Alias) bool {
	if alias, ok := aliasesMap[owner]; ok {
		if alias.AutomergeNotifications {
			targets.Insert(alias.ComEnterpriseSlackChannelsnames...)
			return true
		}
		return false
	}
	return false
}

func (c *OwnersClient) checkIfNotifyUser(owner string, targets *sets.String, usersMap map[string]toolstypes.User) bool {
	if user, ok := usersMap[owner]; ok {
		if user.AutomergeNotifications {
			targets.Insert(user.ComEnterpriseSlackUsername)
			return true
		}
		return false
	}
	return false
}
