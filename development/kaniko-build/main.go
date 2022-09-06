package main

import (
	"flag"
	"fmt"
	tagutil "github.com/kyma-project/test-infra/development/gcbuild/tags"
	"io"
	errutil "k8s.io/apimachinery/pkg/util/errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type options struct {
	Config
	configPath     string
	context        string
	dockerfile     string
	directory      string
	name           string
	variant        string
	logDir         string
	silent         bool
	isCI           bool
	additionalTags additionalTags
}

type additionalTags []string

func (t additionalTags) String() string {
	return strings.Join(t, ",")
}

func (t *additionalTags) Set(val string) error {
	*t = append(*t, val)
	return nil
}

// parseVariable returns a build-arg.
// Keys are set to upper-case.
func parseVariable(key, value string) string {
	k := strings.TrimSpace(strings.ToUpper(key))
	return k + "=" + strings.TrimSpace(value)
}

// run prepares command execution and handles gathering logs to file
func run(o options, name string, destinations []string, buildArgs map[string]string) error {
	args := []string{
		"--context=" + o.context,
		"--dockerfile=" + o.dockerfile,
	}
	for _, dst := range destinations {
		args = append(args, "--destination="+dst)
	}

	for k, v := range buildArgs {
		args = append(args, "--build-arg="+parseVariable(k, v))
	}

	if o.Config.Cache.Enabled {
		args = append(args, "--cache="+strconv.FormatBool(o.Cache.Enabled),
			"--cache-copy-layers="+strconv.FormatBool(o.Cache.CacheCopyLayers),
			"--cache-run-layers="+strconv.FormatBool(o.Cache.CacheRunLayers),
			"--cache-repo="+o.Cache.CacheRepo)
	}

	if o.Config.LogFormat != "" {
		args = append(args, "--log-format="+o.Config.LogFormat)
	}

	if o.Config.Reproducible {
		args = append(args, "--reproducible=true")
	}

	cmd := exec.Command("/kaniko/executor", args...)

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

func runBuildJob(o options, vs Variants) error {
	var sha, pr string
	var err error
	repo := o.Config.Registry
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

	// if sha is still not set, fail the pipeline
	if sha == "" {
		return fmt.Errorf("'sha' could not be determined")
	}

	tags, err := getTags(pr, sha, o.TagTemplate, o.additionalTags)
	if err != nil {
		return err
	}

	if len(vs) == 0 {
		// variants.yaml file not present or either empty. Run single build.
		destinations := gatherDestinations(repo, o.directory, o.name, tags)
		err = run(o, "build", destinations, make(map[string]string))
		if err != nil {
			return fmt.Errorf("build encountered error: %w", err)
		}
		fmt.Println("Successfully built image:", strings.Join(destinations, ", "))
	}

	var wg sync.WaitGroup
	wg.Add(len(vs))
	var errs []error
	for k, v := range vs {
		go func(variant string, env map[string]string) {
			defer wg.Done()
			var variantTags []string
			for _, tag := range tags {
				variantTags = append(variantTags, tag+"-"+variant)
			}
			destinations := gatherDestinations(repo, o.directory, o.name, variantTags)
			if err := run(o, variant, destinations, env); err != nil {
				errs = append(errs, fmt.Errorf("job %s ended with error: %w", variant, err))
				fmt.Printf("Job '%s' ended with error: %s.\n", variant, err)
			} else {
				fmt.Println("Successfully built image:", strings.Join(destinations, ", "))
				fmt.Printf("Job '%s' finished successfully.\n", variant)
			}
		}(k, v)
	}
	wg.Wait()
	return errutil.NewAggregate(errs)
}

func getTags(pr, sha, tagTemplate string, additionalTags []string) ([]string, error) {
	var tags []string
	// (Ressetkk): PR tag should not be hardcoded, in the future we have to find a way to parametrize it
	if pr != "" {
		// assume we are using PR number, build tag as 'PR-XXXX'
		tags = append(tags, "PR-"+pr)
	} else {
		// build a tag from commit SHA
		t, err := tagutil.NewTag(
			tagutil.CommitSHA(sha))
		if err != nil {
			return nil, fmt.Errorf("could not create tag: %w", err)
		}

		tagTmpl := `v{{ .Date }}-{{ .ShortSHA }}`
		if tagTemplate != "" {
			tagTmpl = tagTemplate
		}
		tagger := tagutil.Tagger{TagTemplate: tagTmpl}
		tag, err := tagger.BuildTag(t)
		if err != nil {
			return nil, fmt.Errorf("could not build tag: %w", err)
		}
		tags = append(tags, tag)
		tags = append(tags, additionalTags...)
	}
	return tags, nil
}

func gatherDestinations(repo, directory, name string, tags []string) []string {
	var dst []string
	for _, t := range tags {
		image := path.Join(repo, directory, name)
		dst = append(dst, image+":"+strings.ReplaceAll(t, " ", "-"))
	}
	return dst
}

// validateOptions handles options validation. All checks should be provided here
func validateOptions(o options) error {
	var errs []error
	if o.directory == "" {
		errs = append(errs, fmt.Errorf("flag '--directory' is missing"))
	}
	if o.context == "" {
		errs = append(errs, fmt.Errorf("flag '--context' is missing"))
	}
	if o.name == "" {
		errs = append(errs, fmt.Errorf("flag '--name' is missing"))
	}
	if o.dockerfile == "" {
		errs = append(errs, fmt.Errorf("flag '--dockerfile' is missing"))
	}
	return errutil.NewAggregate(errs)
}

func (o *options) gatherOptions(fs *flag.FlagSet) *flag.FlagSet {
	fs.BoolVar(&o.silent, "silent", false, "Do not push build logs to stdout")
	fs.StringVar(&o.configPath, "config", "/config/kaniko-build-config.yaml", "Path to application config file")
	fs.StringVar(&o.context, "context", ".", "Path to build directory context")
	fs.StringVar(&o.directory, "directory", "", "Destination directory where the image is be pushed. This flag will be ignored if running in presubmit job and devRegistry is provided in config.yaml")
	fs.StringVar(&o.name, "name", "", "Name of the image to be built")
	fs.StringVar(&o.dockerfile, "dockerfile", "Dockerfile", "Path to Dockerfile file relative to context")
	fs.StringVar(&o.variant, "variant", "", "If variants.yaml file is present, define which variant should be built. If variants.yaml is not present, this flag will be ignored")
	fs.StringVar(&o.logDir, "log-dir", "/logs/artifacts", "Path to logs directory where GCB logs will be stored")
	fs.Var(&o.additionalTags, "tag", "Additional tag that the image will be tagged")
	return fs
}

func main() {
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

	context, err := filepath.Abs(o.context)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	variantsFile := filepath.Join(context, filepath.Dir(o.dockerfile), "variants.yaml")
	variant, err := GetVariants(o.variant, variantsFile, os.ReadFile)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	err = runBuildJob(o, variant)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Job's done.")
}
