package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hairyhenderson/go-codeowners"
	k8sowners "k8s.io/test-infra/prow/repoowners"
)

func GetWalkFunc(repoDirectory string, ownersAliases k8sowners.RepoAliases, codeOwners *[]codeowners.Codeowner) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		//pass the error further, this shouldn't ever happen
		if err != nil {
			return err
		}

		// skip directory entries, we just want files
		if info.IsDir() {
			// and completely omit everything inside vendor dir
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// we only want to check .yaml files
		if info.Name() != "OWNERS" {
			return nil
		}

		cleanPath := strings.Replace(path, repoDirectory+"/", "", -1)
		pathBase := filepath.Dir(cleanPath)
		//component = strings.Replace(component, "/values.yaml", "", -1)

		owners, err := ParseOwnersFile(path)
		if err != nil {
			return err
		}

		directoryOwnersMap := make(map[string]bool)
		for _, owner := range owners.Approvers {
			directoryOwnersMap[owner] = true
		}
		for _, owner := range owners.Reviewers {
			directoryOwnersMap[owner] = true
		}
		directoryOwners := make([]string, 0)
		for owner := range directoryOwnersMap {
			project := ""
			if ownersAliases[owner].Len() > 0 {
				// assume groups will use GH groups
				project = "kyma-project/"
			}
			directoryOwners = append(directoryOwners, "@"+project+owner)
		}

		// some OWNERS files only appends labels, skip these for now
		if len(directoryOwners) > 0 {
			owner, err := codeowners.NewCodeowner(pathBase, directoryOwners)
			if err != nil {
				return err
			}

			*codeOwners = append(*codeOwners, owner)
		} else {
			fmt.Fprintf(os.Stderr, "empty or unparsable: %s (%s)\n", cleanPath, path)
		}

		return nil
	}
}
