package main

import (
	"flag"
	"fmt"
	"github.com/kyma-project/test-infra/development/gcbuild/tags"
	"gopkg.in/yaml.v3"
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

type Cloudbuild struct {
	Steps         []Step            `yaml:"steps"`
	Substitutions map[string]string `yaml:"substitutions"`
	Images        []string          `yaml:"images"`
}

func getCloudbuild(f string) (*Cloudbuild, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var cb Cloudbuild
	if err := yaml.Unmarshal(b, &cb); err != nil {
		return nil, fmt.Errorf("cloudbuild.yaml parse error: %w", err)
	}
	return &cb, nil
}

type Step struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"args"`
}

type options struct {
	Config
	configPath   string
	buildDir     string
	cloudbuild   string
	variantsFile string
	variant      string
	logDir       string
	silent       bool
	isCI         bool
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
		"--config", o.cloudbuild,
		"--substitutions", strings.Join(s, ","),
	}

	if o.Project != "" {
		args = append(args, "--project", o.Project)
	}
	// TODO (@Ressetkk): custom staging bucket implementation and re-using source code in builds
	//if o.stagingBucket != "" {
	//
	//}

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

// getVariants fetches variants from variants.yaml file.
// If variant flag is used, it fetches the requested variant.
func getVariants(o options) (variants, error) {
	var v variants
	b, err := ioutil.ReadFile(o.variantsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// variant file not found, skipping
		return nil, nil
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	if o.variant != "" {
		va, ok := v[o.variant]
		if !ok {
			return nil, fmt.Errorf("requested variant '%s', but it's not present in variants.yaml file", o.variant)
		}
		return variants{o.variant: va}, nil
	}
	return v, nil
}

// getImageNames returns aggregated string with a list of images that were built
// It works by replacing substitution variables with actual values calculated by the tool
func getImageNames(repo, tag, variant string, images []string) string {
	r := strings.NewReplacer("$_REPOSITORY", repo, "$_TAG", tag, "$_VARIANT", variant,
		"${_REPOSITORY}", repo, "${_TAG}", tag, "${_VARIANT}", variant)
	var res []string
	for _, i := range images {
		res = append(res, r.Replace(i))
	}
	return strings.Join(res, " ")
}

func runBuildJob(o options) error {
	config, err := getCloudbuild(o.cloudbuild)
	if err != nil {
		return err
	}

	if err := validateConfig(o, config); err != nil {
		return fmt.Errorf("config validation ended with error: %w", err)
	}
	// TODO (Ressetkk): decouple this function so we can test it
	repo := config.Substitutions["_REPOSITORY"]
	var sha, pr string
	if o.isCI {
		presubmit := os.Getenv("JOB_TYPE") == "presubmit"
		if presubmit {
			if o.DevRegistry != "" {
				repo = o.DevRegistry
			}
			if n := os.Getenv("PULL_NUMBER"); n != "" {
				pr = n
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

	var tag string
	// (Ressetkk): PR tag should not be hardcoded, in the future we have to find a way to parametrize it
	if pr != "" {
		// assume we are using PR number, build tag as 'PR-XXXX'
		tag = "PR-" + pr
	} else {
		// build a tag from commit SHA
		t, err := tags.NewTag(
			tags.CommitSHA(sha))
		if err != nil {
			return fmt.Errorf("could not create tag: %w", err)
		}

		tagTmpl := `v{{ .Date }}-{{ .ShortSHA }}`
		if o.TagTemplate != "" {
			tagTmpl = o.TagTemplate
		}
		tagger := tags.Tagger{TagTemplate: tagTmpl}
		tag, err = tagger.BuildTag(t)
		if err != nil {
			return fmt.Errorf("could not build tag: %w", err)
		}
	}

	// TODO (@Ressetkk): custom staging bucket implementation and re-using source code in builds
	//if o.stagingBucket != "" {
	//
	//}

	vs, err := getVariants(o)
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
		img := getImageNames(repo, tag, "", config.Images)
		fmt.Println("Successfully built image:", img)
	}

	var wg sync.WaitGroup
	wg.Add(len(vs))
	var errs []error
	for k, v := range vs {
		go func(variant string, env map[string]string) {
			defer wg.Done()
			env["_VARIANT"] = variant
			if err := run(o, variant, repo, tag, env); err != nil {
				errs = append(errs, fmt.Errorf("job %s ended with error: %w", variant, err))
				fmt.Printf("Job %s ended with error: %s.\n", variant, err)
			} else {
				img := getImageNames(repo, tag, variant, config.Images)
				fmt.Println("Successfully build image:", img)
				fmt.Printf("Job %s finished successfully.\n", variant)
			}
		}(k, v)
	}
	wg.Wait()
	return errutil.NewAggregate(errs)
}

// validateOptions handles options validation. All checks should be provided here
func validateOptions(o options) error {
	// TODO(Ressetkk): These checks are great candidate to be in checks.go as separate check
	var errs []error

	if o.cloudbuild == "" {
		errs = append(errs, fmt.Errorf("'cloudbuild' option is missing, please define this option using flag --cloudbuild-file"))
	}
	if o.buildDir == "" {
		errs = append(errs, fmt.Errorf("'buildDir' option is missing, please define this option using flag --build-dir"))
	}
	if o.Project == "" {
		errs = append(errs, fmt.Errorf("'project' option is missing in the config file"))
	}
	if o.variant != "" && o.variantsFile == "" {
		errs = append(errs, fmt.Errorf("variant option is defined, but variantsFile option is missing"))
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

func (o *options) gatherOptions(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&o.silent, "silent", false, "Do not push build logs to stdout")
	fs.StringVar(&o.configPath, "config", "", "Path to application config file")
	fs.StringVar(&o.buildDir, "build-dir", ".", "Path to build directory")
	fs.StringVar(&o.cloudbuild, "cloudbuild-file", "cloudbuild.yaml", "Path to cloudbuild.yaml file relative to build-dir")
	fs.StringVar(&o.variantsFile, "variants-file", "", "Name of variants file relative to build-dir")
	fs.StringVar(&o.variant, "variant", "", "Define which variant should be built")
	fs.StringVar(&o.logDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	// (Ressetkk): Maybe treat these flags as overrides for config file?
	//fs.StringVar(&o.devRegistry, "dev-registry", "", "Registry URL where development/dirty images should land. If not set then the default registry is used. This flag is only valid when running in CI (CI env variable is set to `true`)")
	//fs.StringVar(&o.project, "project", "", "GCP project name where build jobs will run")
	//fs.StringVar(&o.stagingBucket, "staging-bucket", "", "Full name to the Google Cloud Storage bucket, where the source will be pushed beforehand. If not set, rely on Google Cloud Build")
	//fs.StringVar(&o.logsBucket, "logs-bucket", "", "Full name to the Google Cloud Storage bucket, where the logs will be pushed after build finishes. If not set, rely on Google Cloud Build")
	return fs
}

func main() {
	if err := checkDependencies(); err != nil {
		panic(err)
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{}
	o.gatherOptions(fs)
	// for CI purposes
	o.isCI = os.Getenv("CI") == "true"
	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	c, err := ioutil.ReadFile(o.configPath)
	if err != nil {
		panic(err)
	}

	if err := o.ParseConfig(c); err != nil {
		panic(err)
	}

	// validate if options provided by flags and config file are fine
	if err := validateOptions(o); err != nil {
		panic(err)
	}

	buildDir, err := filepath.Abs(o.buildDir)
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(buildDir); err != nil {
		panic(err)
	}
	err = runBuildJob(o)
	if err != nil {
		panic(err)
	}
	fmt.Println("Job's done.")
}
