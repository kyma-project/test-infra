package repoowners

import (
	"fmt"
	toolstypes "github.com/kyma-project/test-infra/pkg/types"

	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/prow/prow/github"
	k8sowners "sigs.k8s.io/prow/prow/repoowners"
)

// AllOwners holds repository owners as map.
// Owners are keep without duplicates and provide easy way for checking owner presence.
type AllOwners map[string]struct{}

// addOwners add list of owners to AllOwners as map keys.
func (a AllOwners) addOwners(approvers []string) {
	for _, approver := range approvers {
		a[approver] = struct{}{}
	}
}

// LoadRepoAliases parse and load repository aliases from file.
// A file is a local file at basedir/filename.
func (c *OwnersClient) LoadRepoAliases(basedir, filename string) (k8sowners.RepoAliases, error) {
	path := filepath.Join(basedir, filename)
	c.Logger.Debugf("Loading repo aliases from path %s", path)
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return k8sowners.ParseAliasesConfig(b)
}

// GetOwnersForChanges find owners for pullrequest changes.
// repoBase is a local repository root path.
func (c *OwnersClient) GetOwnersForChanges(changes []github.PullRequestChange, repoBase string, owners k8sowners.RepoOwner) (AllOwners, error) {
	var (
		conf k8sowners.Config
		err  error
	)
	allOwners := AllOwners{}
	// Go over pull request changes.
	for _, change := range changes {
		// Find OWNERS file with approvers for changed file.
		approversFile := owners.FindApproverOwnersForFile(change.Filename)
		// Find OWNERS file with reviewers for changed file.
		reviewersFile := owners.FindReviewersOwnersForFile(change.Filename)
		// Parse found OWNERS file with approvers.
		conf, err = c.ParseOwnersFile(approversFile, repoBase, owners)
		if err != nil {
			return nil, err
		}
		allOwners.addOwners(conf.Approvers)
		// Parse OWNERS file with reviewers if it's a different one than file with approvers.
		if approversFile != reviewersFile {
			// Parse found OWNERS file with reviewers.
			conf, err = c.ParseOwnersFile(reviewersFile, repoBase, owners)
			if err != nil {
				return nil, err
			}
		}
		allOwners.addOwners(conf.Reviewers)
	}
	return allOwners, nil
}

// ParseOwnersFile parse OWNERS file. A file is a local file at repoBase/ownersFilePath.
// ownersFilePath is a path to OWNERS file relative from repository root.
// repoBase is a local absolute path.
func (c *OwnersClient) ParseOwnersFile(ownersFilePath, repoBase string, owners k8sowners.RepoOwner) (k8sowners.Config, error) {
	var (
		fullconfig   k8sowners.FullConfig
		simpleconfig k8sowners.SimpleConfig
		err          error
	)
	ownersFile := filepath.Join(repoBase, ownersFilePath, "OWNERS")

	// Parse OWNERS file as SimpleConfig.
	simpleconfig, err = owners.ParseSimpleConfig(ownersFile)
	if err != nil {
		return k8sowners.Config{}, fmt.Errorf("failed parsing %s owners file to simpleconfig, error: %w", ownersFilePath, err)
	}
	// Check if parsed SimpleConfig is empty and parse OWNERS file as FullConfig when true.
	if simpleconfig.Empty() {
		fullconfig, err = owners.ParseFullConfig(ownersFile)
		if err != nil {
			return k8sowners.Config{}, fmt.Errorf("failed parsing %s owners file to fullconfig, error: %w", ownersFilePath, err)
		}
		return fullconfig.Filters[".*"], nil
	}
	return simpleconfig.Config, nil
}

// ResolveSlackNames find all owners from allOwners, in repository owners aliases and users lists which have enabled pr automerge notifications.
func (c *OwnersClient) ResolveSlackNames(allOwners AllOwners, aliases []toolstypes.Alias, users []toolstypes.User, repoAliases k8sowners.RepoAliases) (sets.Set[string], error) {
	// Convert lists of aliases and users to maps for easy searching.
	// Using gitHub alias and user names as searching keys.
	aliasesMap := make(map[string]toolstypes.Alias)
	usersMap := make(map[string]toolstypes.User)
	targets := sets.New[string]()
	for _, alias := range aliases {
		aliasesMap[alias.ComGithubAliasname] = alias
	}
	for _, user := range users {
		usersMap[user.ComGithubUsername] = user
	}

	// Search in maps, to find all owners with enabled notification for pr automerge.
	for owner := range allOwners {
		// Search in aliases map.
		if !c.checkIfNotifyAlias(owner, &targets, aliasesMap) {
			// Search in users map.
			if !c.checkIfNotifyUser(owner, &targets, usersMap) {
				// Check if owner was an alias and expand it to GitHub users.
				// Search in users map for all expanded users.
				// Expanding owners aliases to users in case owners aliases is not present in aliases map, or it has disabled pr automerge notification.
				// In such case owners users with enabled pr automerge notification will be search and notified individually.
				userOwners := repoAliases.ExpandAlias(owner)
				if userOwners.Len() > 0 {
					for userOwner := range userOwners {
						_ = c.checkIfNotifyUser(userOwner, &targets, usersMap)
					}
				}
			}
		}
	}
	return targets, nil
}

// checkIfNotifyAlias checks in aliases map if owner is a repository owner alias and has enabled notifications for pr automerge events.
func (c *OwnersClient) checkIfNotifyAlias(owner string, targets *sets.Set[string], aliasesMap map[string]toolstypes.Alias) bool {
	if alias, ok := aliasesMap[owner]; ok {
		if alias.AutomergeNotifications {
			// Add owner to notification targets.
			targets.Insert(alias.ComEnterpriseSlackChannelsnames...)
			return true
		}
		return false
	}
	return false
}

// checkIfNotifyUser checks in users map  if owner is a repository owner user and has enabled notifications for pr automerge events.
func (c *OwnersClient) checkIfNotifyUser(owner string, targets *sets.Set[string], usersMap map[string]toolstypes.User) bool {
	if user, ok := usersMap[owner]; ok {
		if user.AutomergeNotifications {
			// Add owner to notification targets.
			targets.Insert(user.ComEnterpriseSlackUsername)
			return true
		}
		return false
	}
	return false
}
