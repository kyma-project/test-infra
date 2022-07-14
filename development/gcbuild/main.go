package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/gcbuild/tags"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	errutil "k8s.io/apimachinery/pkg/util/errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type FileGetter interface {
	GetFile(f string) io.ReadCloser
}

type Config struct {
	Steps         []Step            `yaml:"steps"`
	Substitutions map[string]string `yaml:"substitutions"`
	Images        []string          `yaml:"images"`
}

func getConfig(f string) (*Config, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var cb Config
	if err := yaml.Unmarshal(b, &cb); err != nil {
		return nil, fmt.Errorf("cloudbuild.yaml parse error: %w", err)
	}
	return &cb, nil
}

type Step struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"args"`
}

type BuildCtx struct {
	variants variants
	dir      string
	name     string
}

type options struct {
	buildDir           string
	cloudBuildYAMLFile string
	variantsFile       string
	logDir             string
	devRegistry        string
	project            string
	silent             bool
	isCI               bool
	tagger             tags.Tagger
}

func (o *options) gatherOptions(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&o.silent, "silent", false, "Do not push build logs to stdout")
	fs.StringVar(&o.buildDir, "build-dir", ".", "Path to build directory")
	fs.StringVar(&o.cloudBuildYAMLFile, "cloudbuild-file", "cloudbuild.yaml", "Path to cloudbuild.yaml file relative to build-dir")
	fs.StringVar(&o.variantsFile, "variants-file", "variants.yaml", "Name of variants file relative to build-dir")
	fs.StringVar(&o.logDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	fs.StringVar(&o.devRegistry, "dev-registry", "", "Registry URL where development/dirty images should land. If not set then the default registry is used. This flag is only valid when running in CI (CI env variable is set to `true`)")
	fs.StringVar(&o.project, "project", "", "GCP project name where build jobs will run")
	o.tagger.AddFlags(fs)
	return fs
}

// parseVariable returns a gcloud compatible substitution option.
// Keys are set to upper-case and prefix "_" is set if not present.
func parseVariable(key, value string) string {
	k := strings.TrimSpace(strings.ToUpper(key))
	if !strings.HasPrefix(k, "_") {
		k = "_" + k
	}
	return k + "=" + strings.TrimSpace(value)
}

func getSHAFromGit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "head")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	validTagRegexp, err := regexp.Compile("[^-_.a-zA-Z0-9]+")
	if err != nil {
		return "", err
	}
	sanitizedOutput := validTagRegexp.ReplaceAllString(string(out), "")
	return sanitizedOutput, nil
}

// run prepares command execution and handles gathering logs to file
func run(o options, name, repo, tag string, subs map[string]string) error {
	var s []string

	s = append(s,
		parseVariable("_TAG", tag),
		parseVariable("_REPOSITORY", repo),
	)

	// parse additional substitutions
	for k, v := range subs {
		s = append(s, parseVariable(k, v))
	}
	args := []string{
		"builds", "submit",
		"--config", o.cloudBuildYAMLFile,
		"--substitutions", strings.Join(s, ","),
	}

	if o.project != "" {
		args = append(args, "--project", o.project)
	}

	cmd := exec.Command("gcloud", args...)

	var outw []io.Writer
	var errw []io.Writer

	if !o.silent {
		outw = append(outw, os.Stdout)
		errw = append(errw, os.Stderr)
	}

	f, err := os.Create(filepath.Join(o.logDir, strings.TrimSpace("build_"+strings.TrimSpace(name)+".log")))
	if err != nil {
		return fmt.Errorf("could not create log file: %w", err)
	}

	outw = append(outw, f)
	errw = append(errw, f)

	cmd.Stdout = io.MultiWriter(outw...)
	cmd.Stderr = io.MultiWriter(errw...)

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

type variants map[string]map[string]string

func getVariants(f string) (variants, error) {
	var v variants
	b, err := ioutil.ReadFile(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// variant file not found, skipping
		return nil, nil
	}
	if err := yaml.UnmarshalStrict(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func getImageNames(repo, tag, variant string, images []string) string {
	r := strings.NewReplacer("$_REPOSITORY", repo, "$_TAG", tag, "$_VARIANT", variant)
	var res string
	for _, i := range images {
		res = res + r.Replace(i) + " "
	}
	return res
}

func runBuildJob(o options) error {
	buildDir, err := filepath.Abs(o.buildDir)
	if err != nil {
		return err
	}
	if err := os.Chdir(buildDir); err != nil {
		return err
	}
	/*
		REQUIRED SUBSTITUTIONS:
		- _TAG
		- _REPOSITORY
	*/

	f, err := getConfig(o.cloudBuildYAMLFile)
	if err != nil {
		return err
	}

	if err := validateConfig(f); err != nil {
		return fmt.Errorf("could not validate config: %w", err)
	}

	repo := f.Substitutions["_REPOSITORY"]
	var sha string
	if o.isCI {
		presubmit := os.Getenv("JOB_TYPE") == "presubmit"
		if presubmit {
			if o.devRegistry != "" {
				repo = o.devRegistry
			}
		}

		if c := os.Getenv("PULL_BASE_SHA"); c != "" {
			sha = c
		}
	}

	// if sha is still not set, use git to discover it
	if sha == "" {
		sha, err = getSHAFromGit()
		if err != nil {
			return err
		}
	}

	// build a tag
	t, err := tags.NewTag(
		tags.CommitSHA(sha))
	if err != nil {
		return fmt.Errorf("could not create tag: %w", err)
	}
	tag, err := o.tagger.BuildTag(t)
	if err != nil {
		return fmt.Errorf("could not build tag: %w", err)
	}

	vs, err := getVariants(o.variantsFile)
	if err != nil {
		return err
	}

	if len(vs) == 0 {
		// variants.yaml file not present or either empty. Run single build.
		// subs act as overrides here for YAML substitutions
		subs := make(map[string]string)

		err = run(o, "build", repo, tag, subs)
		if err != nil {
			return fmt.Errorf("build encountered error: %w", err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(vs))
	var errs []error
	for k, v := range vs {
		go func(variant string, env map[string]string) {
			defer wg.Done()
			if err := run(o, variant, repo, tag, env); err != nil {
				errs = append(errs, fmt.Errorf("job %s ended with error: %w", variant, err))
				fmt.Printf("Job %s ended with error: %s.\n", variant, err)
			} else {
				img := getImageNames(repo, tag, variant, f.Images)
				fmt.Println("Successfully build image: ", img)
				fmt.Printf("Job %s finished successfully.\n", variant)
			}
		}(k, v)
	}
	wg.Wait()
	return errutil.NewAggregate(errs)
}

func validateOptions(o options) error {
	var errs []error
	if o.project == "" {
		errs = append(errs, fmt.Errorf("--project flag is missing"))
	}
	return errutil.NewAggregate(errs)
}

// checkDependencies checks if required binaries are present in PATH
// If they are not present, then returns proper error message.
func checkDependencies() error {
	deps := []string{
		"gcloud",
		"gsutil",
		"git",
	}
	var errs []error
	for _, c := range deps {
		if _, err := exec.LookPath(c); err != nil {
			errs = append(errs, err)
		}
	}
	return errutil.NewAggregate(errs)
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{}
	o.gatherOptions(fs)
	// for CI purposes
	o.isCI = os.Getenv("CI") == "true"

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
	if err := validateOptions(o); err != nil {
		panic(err)
	}
	if err := checkDependencies(); err != nil {
		panic(err)
	}
	err := runBuildJob(o)
	if err != nil {
		panic(err)
	}
	fmt.Println("Job's done.")
}
