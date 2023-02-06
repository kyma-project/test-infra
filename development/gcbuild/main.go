package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/kyma-project/test-infra/development/gcbuild/config"
	"github.com/kyma-project/test-infra/development/pkg/tags"
	errutil "k8s.io/apimachinery/pkg/util/errors"
)

type options struct {
	config.Config
	configPath string
	buildDir   string
	cloudbuild string
	variant    string
	logDir     string
	silent     bool
	isCI       bool
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

func runBuildJob(o options, cb *config.CloudBuild, vs config.Variants) error {
	if err := config.ValidateConfig(nil, cb, vs); err != nil {
		return fmt.Errorf("config validation ended with error: %w", err)
	}
	// TODO (Ressetkk): decouple this function so we can test it
	repo := cb.Substitutions["_REPOSITORY"]
	var sha, pr string
	var err error
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
		tagTmpl := `v{{ .Date }}-{{ .ShortSHA }}`
		if o.TagTemplate != "" {
			tagTmpl = o.TagTemplate
		}
		// build a tag from commit SHA
		tagger, err := tags.NewTagger([]tags.Tag{{Name: "TagTemplate", Value: tagTmpl}}, tags.CommitSHA(sha))
		if err != nil {
			return fmt.Errorf("get tagger: %w", err)
		}
		p, err := tagger.ParseTags()
		if err != nil {
			return fmt.Errorf("build tag: %w", err)
		}
		// we'll always get one tag in this slice
		tag = p[0].Value
	}

	// TODO (@Ressetkk): custom staging bucket implementation and re-using source code in builds
	//if o.stagingBucket != "" {
	//
	//}

	if len(vs) == 0 {
		// variants.yaml file not present or either empty. Run single build.
		// subs act as overrides for YAML substitutions
		subs := make(map[string]string)

		err = run(o, "build", repo, tag, subs)
		if err != nil {
			return fmt.Errorf("build encountered error: %w", err)
		}
		img := getImageNames(repo, tag, "", cb.Images)
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
				img := getImageNames(repo, tag, variant, cb.Images)
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
	var errs []error
	if o.cloudbuild == "" {
		errs = append(errs, fmt.Errorf("'cloudbuild' option is missing, please define this option using '--cloudbuild-file' flag"))
	}
	if o.buildDir == "" {
		errs = append(errs, fmt.Errorf("'buildDir' option is missing, please define this option using '--build-dir' flag"))
	}
	if o.Project == "" {
		errs = append(errs, fmt.Errorf("'project' option is missing in the config file"))
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
	fs.StringVar(&o.variant, "variant", "", "If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored")
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
		fmt.Println(err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	o := options{isCI: os.Getenv("CI") == "true"}
	o.gatherOptions(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if o.configPath == "" {
		fmt.Println("'--config' flag is missing or has empty value, please provide the path to valid 'config.yaml' file")
		os.Exit(1)
	}
	c, err := os.ReadFile(o.configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := o.ParseConfig(c); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// validate if options provided by flags and config file are fine
	if err := validateOptions(o); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	buildDir, err := filepath.Abs(o.buildDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// jump into buildDir directory and run stuff from there
	if err := os.Chdir(buildDir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cbDir := filepath.Dir(o.cloudbuild)
	variantsPath := filepath.Join(cbDir, "variants.yaml")

	cb, err := config.GetCloudBuild(o.cloudbuild, os.ReadFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	variant, err := config.GetVariants(o.variant, variantsPath, os.ReadFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	err = runBuildJob(o, cb, variant)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Job's done.")
}
