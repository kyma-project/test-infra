package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/gcbuild/config"
	errutil "k8s.io/apimachinery/pkg/util/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var execCmd = exec.Command

type filesGetter func() ([]string, error)

type options struct {
	fromGit bool
	isCI    bool
	baseSha string
}

func (o *options) gatherOptions(fs *flag.FlagSet) {
	fs.BoolVar(&o.fromGit, "from-git", false, "Load changed files from git directory, assuming you run this tool in git-enabled directory")
	fs.StringVar(&o.baseSha, "base-sha", "", "When paired with --from-git, fetch between HEAD and provided SHA")

}

func fromGit(baseSha string) filesGetter {
	return func() ([]string, error) {
		cmd := execCmd("git", "diff", "--name-only", "--diff-filter=d", baseSha+"...HEAD")
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return strings.Split(string(out), "\n"), nil
	}
}

func fromArgs(fs *flag.FlagSet) filesGetter {
	return func() ([]string, error) {
		return fs.Args(), nil
	}
}

func baseRef(o options) (string, error) {
	// default value from flag
	sha := o.baseSha
	if o.isCI {
		// we work in CI, try to fetch base ref from env variables, since --base-sha flag may not be present
		sha = os.Getenv("PULL_BASE_SHA")
	}
	if sha == "" {
		return "", fmt.Errorf("could not guess 'baseSha', provide it using '--base-sha' flag or make sure, you are running in CI environment and have 'PULL_BASE_SHA' environment variable present")
	}
	return sha, nil
}

func filterFiles(files []string) []string {
	var cbs []string
	for _, f := range files {
		if strings.Contains(f, "cloudbuild.yaml") {
			cbs = append(cbs, f)
		}
	}
	return cbs
}

func getFiles(o options, fs *flag.FlagSet) ([]string, error) {
	var fg filesGetter
	if o.fromGit {
		br, err := baseRef(o)
		if err != nil {
			return nil, err
		}
		fg = fromGit(br)
	} else {
		fg = fromArgs(fs)
	}
	files, err := fg()
	if err != nil {
		return nil, err
	}
	return filterFiles(files), nil

}

func report(f string, err error) {
	switch e := err.(type) {
	case errutil.Aggregate:
		for _, er := range e.Errors() {
			fmt.Println("[", f, "]:", er)
		}
	default:
		fmt.Println(err)
	}
}

func main() {
	o := options{isCI: os.Getenv("CI") == "true"}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o.gatherOptions(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	files, err := getFiles(o, fs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(files) > 0 {
		for _, f := range files {
			dir := filepath.Dir(f)
			cb, err := config.GetCloudBuild(f, os.ReadFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			v, err := config.GetVariants("", filepath.Join(dir, "variants.yaml"), os.ReadFile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if err := config.ValidateConfig(nil, cb, v); err != nil {
				report(f, err)
				os.Exit(1)
			}
		}
	}
	fmt.Println("Job's done.")
}
