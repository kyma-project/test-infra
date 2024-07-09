package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/kyma-project/test-infra/pkg/github/bumper"
	"github.com/kyma-project/test-infra/pkg/github/imagebumper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"golang.org/x/oauth2/google"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	AutobumpConfig  string
	GithubTokenPath string
)

const (
	latestVersion           = "latest"
	upstreamVersion         = "upstream"
	upstreamStagingVersion  = "upstream-staging"
	tagVersion              = "vYYYYMMDD-deadbeef"
	defaultUpstreamURLBase  = "https://raw.githubusercontent.com/kubernetes/test-infra/master"
	googleImageRegistryAuth = "google"
	cloudPlatformScope      = "https://www.googleapis.com/auth/cloud-platform"
)

var rootCmd = &cobra.Command{
	Use:   "autobumper",
	Short: "Autobumper CLI",
	Long:  "Command-Line tool to update images in pipeline files and create PRs for them",
	Run: func(_ *cobra.Command, _ []string) {
		// Run autobumper if autobump config provided
		if AutobumpConfig != "" {
			err := runAutobumper(AutobumpConfig)
			if err != nil {
				log.Fatalf("failed to run bumper: %s", err)
			}
		} else {
			log.Fatalf("autobump-config is required")
		}
	},
}

var (
	tagRegexp    = regexp.MustCompile("v[0-9]{8}-[a-f0-9]{6,9}")
	imageMatcher = regexp.MustCompile(`(?s)^.+image:(.+):(v[a-zA-Z0-9_.-]+)`)
)

func init() {
	rootCmd.PersistentFlags().StringVar(&AutobumpConfig, "autobump-config", "", "path to the config for autobumper for security scanner config")
	rootCmd.PersistentFlags().StringVar(&GithubTokenPath, "github-token-path", "/etc/github/token", "path to github token for fetching inrepo config")

	rootCmd.MarkFlagRequired("autobump-config")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to run command: %s", err)
	}
}

// options is the options for autobumper operations.
type options struct {
	GitHubRepo      string   `yaml:"gitHubRepo"`
	FoldersToFilter []string `yaml:"foldersToFilter"`
	FilesToFilter   []string `yaml:"filesToFilter"`
}

// client is bumper client
type client struct {
	o        *bumper.Options
	images   map[string]string
	versions map[string][]string
}

// getVersionsAndCheckConsistency takes a list of Prefixes and a map of
// all the images found in the code before the bump : their versions after the bump
// For example {"gcr.io/k8s-prow/test1:tag": "newtag", "gcr.io/k8s-prow/test2:tag": "newtag"},
// and returns a map of new versions resulted from bumping : the images using those versions.
// It will error if one of the Prefixes was bumped inconsistently when it was not supposed to
func getVersionsAndCheckConsistency(prefixes []bumper.Prefix, images map[string]string) (map[string][]string, error) {
	// Key is tag, value is full image.
	versions := map[string][]string{}
	for _, prefix := range prefixes {
		exceptions := sets.NewString(prefix.ConsistentImageExceptions...)
		var consistencyVersion, consistencySourceImage string
		for k, v := range images {
			if strings.HasPrefix(k, prefix.Prefix) {
				image := imageFromName(k)
				if prefix.ConsistentImages && !exceptions.Has(image) {
					if consistencySourceImage != "" && (consistencyVersion != v) {
						return nil, fmt.Errorf("%s -> %s not bumped consistently for prefix %s (%s), expected version %s based on bump of %s", k, v, prefix.Prefix, prefix.Name, consistencyVersion, consistencySourceImage)
					}
					if consistencySourceImage == "" {
						consistencyVersion = v
						consistencySourceImage = k
					}
				}

				// Only add bumped images to the new versions map
				if !strings.Contains(k, v) {
					versions[v] = append(versions[v], k)
				}
			}
		}
	}
	return versions, nil
}

// Changes returns a slice of functions, each one does some stuff, and
// returns commit message for the changes
func (c *client) Changes() []func(context.Context) (string, []string, error) {
	return []func(context.Context) (string, []string, error){
		func(ctx context.Context) (string, []string, error) {
			var err error
			if c.images, err = updateReferencesWrapper(ctx, c.o); err != nil {
				return "", nil, fmt.Errorf("failed to update image references: %w", err)
			}

			if c.versions, err = getVersionsAndCheckConsistency(c.o.Prefixes, c.images); err != nil {
				return "", nil, err
			}

			var body string
			var prefixNames []string
			for _, prefix := range c.o.Prefixes {
				prefixNames = append(prefixNames, prefix.Name)
				body = body + generateSummary(prefix.Name, prefix.Repo, prefix.Prefix, prefix.Summarise, c.images) + "\n\n"
			}

			return fmt.Sprintf("Bumping %s\n\n%s", strings.Join(prefixNames, " and "), body), nil, nil
		},
	}
}

// updateReferencesWrapper update the references of prow-images and/or boskos-images and/or testimages
// in the files in any of "subfolders" of the includeConfigPaths but not in excludeConfigPaths
// if the file is a yaml file (*.yaml) or extraFiles[file]=true
func updateReferencesWrapper(ctx context.Context, o *bumper.Options) (map[string]string, error) {
	logrus.Info("Bumping image references...")
	var allPrefixes []string
	for _, prefix := range o.Prefixes {
		allPrefixes = append(allPrefixes, prefix.Prefix)
	}
	filterRegexp, err := regexp.Compile(strings.Join(allPrefixes, "|"))
	if err != nil {
		return nil, fmt.Errorf("bad regexp %q: %w", strings.Join(allPrefixes, "|"), err)
	}
	var client = http.DefaultClient
	if o.ImageRegistryAuth == googleImageRegistryAuth {
		var err2 error
		client, err2 = google.DefaultClient(ctx, cloudPlatformScope)
		fmt.Println("Error: ", err2)
		if err != nil {
			return nil, fmt.Errorf("failed to create authed client: %v", err)
		}
	}
	imageBumperCli := imagebumper.NewClient(client)
	return updateReferences(imageBumperCli, filterRegexp, o)
}

type imageBumper interface {
	FindLatestTag(imageHost, imageName, currentTag string) (string, error)
	UpdateFile(tagPicker func(imageHost, imageName, currentTag string) (string, error), path string, imageFilter *regexp.Regexp) error
	GetReplacements() map[string]string
	AddToCache(image, newTag string)
	TagExists(imageHost, imageName, currentTag string) (bool, error)
}

// used by updateReferences
func parseUpstreamImageVersion(upstreamAddress, prefix string) (string, error) {
	resp, err := http.Get(upstreamAddress)
	if err != nil {
		return "", fmt.Errorf("error sending GET request to %q: %w", upstreamAddress, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error %d (%q) fetching upstream config file", resp.StatusCode, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading the response body: %w", err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(body), "\n"), "\n") {
		res := imageMatcher.FindStringSubmatch(string(line))
		if len(res) > 2 && strings.Contains(res[1], prefix) {
			return res[2], nil
		}
	}
	return "", fmt.Errorf("unable to find match for %s in upstream refConfigFile", prefix)
}

func updateReferences(imageBumperCli imageBumper, filterRegexp *regexp.Regexp, o *bumper.Options) (map[string]string, error) {
	var tagPicker func(string, string, string) (string, error)

	switch o.TargetVersion {
	case latestVersion:
		tagPicker = imageBumperCli.FindLatestTag
	case upstreamVersion, upstreamStagingVersion:
		var err error
		if tagPicker, err = upstreamImageVersionResolver(o, o.TargetVersion, parseUpstreamImageVersion, imageBumperCli); err != nil {
			return nil, fmt.Errorf("failed to resolve the %s image version: %w", o.TargetVersion, err)
		}
	default:
		tagPicker = func(_, _, _ string) (string, error) { return o.TargetVersion, nil }
	}

	updateFile := func(name string) error {
		logrus.WithField("file", name).Info("Updating file")
		if err := imageBumperCli.UpdateFile(tagPicker, name, filterRegexp); err != nil {
			return fmt.Errorf("failed to update the file: %w", err)
		}
		return nil
	}
	updateYAMLFile := func(name string) error {
		if (strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")) && !isUnderPath(name, o.ExcludedConfigPaths) {
			return updateFile(name)
		}
		return nil
	}

	// Updated all .yaml and .yml files under the included config paths but not under excluded config paths.
	for _, path := range o.IncludedConfigPaths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get the file info for %q: %w", path, err)
		}
		if info.IsDir() {
			err := filepath.Walk(path, func(subpath string, _ os.FileInfo, _ error) error {
				return updateYAMLFile(subpath)
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update yaml files under %q: %w", path, err)
			}
		} else {
			if err := updateYAMLFile(path); err != nil {
				return nil, fmt.Errorf("failed to update the yaml file %q: %w", path, err)
			}
		}
	}

	// Update the extra files in any case.
	for _, file := range o.ExtraFiles {
		if err := updateFile(file); err != nil {
			return nil, fmt.Errorf("failed to update the extra file %q: %w", file, err)
		}
	}

	return imageBumperCli.GetReplacements(), nil
}

// used by upstreamImageVersionResolver
func upstreamConfigVersions(upstreamVersionType string, o *bumper.Options, parse func(upstreamAddress, prefix string) (string, error)) (versions map[string]string, err error) {
	versions = make(map[string]string)
	var upstreamAddress string
	for _, prefix := range o.Prefixes {
		if upstreamVersionType == upstreamVersion {
			upstreamAddress = o.UpstreamURLBase + "/" + prefix.RefConfigFile
		} else if upstreamVersionType == upstreamStagingVersion {
			upstreamAddress = o.UpstreamURLBase + "/" + prefix.StagingRefConfigFile
		} else {
			return nil, fmt.Errorf("unsupported upstream version type: %s, must be one of %v",
				upstreamVersionType, []string{upstreamVersion, upstreamStagingVersion})
		}
		version, err := parse(upstreamAddress, prefix.Prefix)
		if err != nil {
			return nil, err
		}
		versions[prefix.Prefix] = version
	}

	return versions, nil
}

// used by updateReferences
func upstreamImageVersionResolver(
	o *bumper.Options, upstreamVersionType string, parse func(upstreamAddress, prefix string) (string, error), imageBumperCli imageBumper) (func(imageHost, imageName, currentTag string) (string, error), error) {
	upstreamVersions, err := upstreamConfigVersions(upstreamVersionType, o, parse)
	if err != nil {
		return nil, err
	}

	return func(imageHost, imageName, currentTag string) (string, error) {
		imageFullPath := imageHost + "/" + imageName + ":" + currentTag
		for prefix, version := range upstreamVersions {
			if !strings.HasPrefix(imageFullPath, prefix) {
				continue
			}
			if exists, err := imageBumperCli.TagExists(imageHost, imageName, version); err != nil {
				return "", err
			} else if exists {
				imageBumperCli.AddToCache(imageFullPath, version)
				return version, nil
			}
			imageBumperCli.AddToCache(imageFullPath, currentTag)
			return "", fmt.Errorf("Unable to bump to %s, image tag %s does not exist for %s", imageFullPath, version, imageName)
		}
		return currentTag, nil
	}, nil
}

// makeCommitSummary takes a list of Prefixes and a map of new tags resulted
// from bumping : the images using those tags and returns a summary of what was
// bumped for use in the commit message
func makeCommitSummary(prefixes []bumper.Prefix, versions map[string][]string) string {
	var allPrefixes []string
	for _, prefix := range prefixes {
		allPrefixes = append(allPrefixes, prefix.Name)
	}
	if len(versions) == 0 {
		return fmt.Sprintf("Update %s images as necessary", strings.Join(allPrefixes, ", "))
	}
	var inconsistentBumps []string
	var consistentBumps []string
	for _, prefix := range prefixes {
		tag, bumped := isBumpedPrefix(prefix, versions)
		if !prefix.ConsistentImages && bumped {
			inconsistentBumps = append(inconsistentBumps, prefix.Name)
		} else if prefix.ConsistentImages && bumped {
			consistentBumps = append(consistentBumps, fmt.Sprintf("%s to %s", prefix.Name, tag))
		}
	}
	var msgs []string
	if len(consistentBumps) != 0 {
		msgs = append(msgs, strings.Join(consistentBumps, ", "))
	}
	if len(inconsistentBumps) != 0 {
		msgs = append(msgs, fmt.Sprintf("%s as needed", strings.Join(inconsistentBumps, ", ")))
	}
	return fmt.Sprintf("Update %s", strings.Join(msgs, " and "))

}

// PRTitleBody returns the body of the PR, this function runs after each commit
func (c *client) PRTitleBody() (string, string, error) {
	body := generatePRBody(c.images, c.o.Prefixes) + "\n"
	if c.o.AdditionalPRBody != "" {
		body += c.o.AdditionalPRBody + "\n"
	}
	return makeCommitSummary(c.o.Prefixes, c.versions), body, nil
}

func generatePRBody(images map[string]string, prefixes []bumper.Prefix) (body string) {
	body = ""
	for _, prefix := range prefixes {
		body = body + generateSummary(prefix.Name, prefix.Repo, prefix.Prefix, prefix.Summarise, images) + "\n\n"
	}
	return body + "\n"
}

// Generate PR summary for github
func generateSummary(_, repo, prefix string, summarise bool, images map[string]string) string {
	type delta struct {
		oldCommit string
		newCommit string
		oldDate   string
		newDate   string
		variant   string
		component string
	}
	versions := map[string][]delta{}
	for image, newTag := range images {
		if !strings.HasPrefix(image, prefix) {
			continue
		}
		if strings.HasSuffix(image, ":"+newTag) {
			continue
		}
		oldDate, oldCommit, oldVariant := imagebumper.DeconstructTag(tagFromName(image))
		newDate, newCommit, _ := imagebumper.DeconstructTag(newTag)
		oldCommit = commitToRef(oldCommit)
		newCommit = commitToRef(newCommit)
		k := oldCommit + ":" + newCommit
		d := delta{
			oldCommit: oldCommit,
			newCommit: newCommit,
			oldDate:   oldDate,
			newDate:   newDate,
			variant:   formatVariant(oldVariant),
			component: componentFromName(image),
		}
		versions[k] = append(versions[k], d)
	}

	switch {
	case len(versions) == 0:
		return fmt.Sprintf("No %s changes.", prefix)
	case len(versions) == 1 && summarise:
		for k, v := range versions {
			s := strings.Split(k, ":")
			return fmt.Sprintf("%s changes: %s/compare/%s...%s (%s → %s)", prefix, repo, s[0], s[1], formatTagDate(v[0].oldDate), formatTagDate(v[0].newDate))
		}
	default:
		changes := make([]string, 0, len(versions))
		for k, v := range versions {
			s := strings.Split(k, ":")
			names := make([]string, 0, len(v))
			for _, d := range v {
				names = append(names, d.component+d.variant)
			}
			sort.Strings(names)
			changes = append(changes, fmt.Sprintf("%s/compare/%s...%s | %s&nbsp;&#x2192;&nbsp;%s | %s",
				repo, s[0], s[1], formatTagDate(v[0].oldDate), formatTagDate(v[0].newDate), strings.Join(names, ", ")))
		}
		sort.Slice(changes, func(i, j int) bool { return strings.Split(changes[i], "|")[1] < strings.Split(changes[j], "|")[1] })
		return fmt.Sprintf("Multiple distinct %s changes:\n\nCommits | Dates | Images\n--- | --- | ---\n%s\n", prefix, strings.Join(changes, "\n"))
	}
	panic("unreachable!")
}

func validateOptions(o bumper.Options) error {
	if len(o.Prefixes) == 0 {
		return errors.New("must have at least one Prefix specified")
	}
	for _, prefix := range o.Prefixes {
		if len(prefix.ConsistentImageExceptions) > 0 && !prefix.ConsistentImages {
			return fmt.Errorf("consistentImageExceptions requires consistentImages to be true, found in prefix %q", prefix.Name)
		}
	}
	if len(o.IncludedConfigPaths) == 0 {
		return errors.New("includedConfigPaths is mandatory")
	}
	if o.TargetVersion != latestVersion && o.TargetVersion != upstreamVersion &&
		o.TargetVersion != upstreamStagingVersion && !tagRegexp.MatchString(o.TargetVersion) {
		logrus.WithField("allowed", []string{latestVersion, upstreamVersion, upstreamStagingVersion, tagVersion}).Warn(
			"Warning: targetVersion mot in allowed so it might not work properly.")
	}
	if o.TargetVersion == upstreamVersion {
		for _, prefix := range o.Prefixes {
			if prefix.RefConfigFile == "" {
				return fmt.Errorf("targetVersion can't be %q without refConfigFile for each prefix. %q is missing one", upstreamVersion, prefix.Name)
			}
		}
	}
	if o.TargetVersion == upstreamStagingVersion {
		for _, prefix := range o.Prefixes {
			if prefix.StagingRefConfigFile == "" {
				return fmt.Errorf("targetVersion can't be %q without stagingRefConfigFile for each prefix. %q is missing one", upstreamStagingVersion, prefix.Name)
			}
		}
	}
	if (o.TargetVersion == upstreamVersion || o.TargetVersion == upstreamStagingVersion) && o.UpstreamURLBase == "" {
		o.UpstreamURLBase = defaultUpstreamURLBase
		logrus.Warnf("targetVersion can't be 'upstream' or 'upstreamStaging` without upstreamURLBase set. Default upstreamURLBase is %q", defaultUpstreamURLBase)
	}

	if o.ImageRegistryAuth != "" && o.ImageRegistryAuth != googleImageRegistryAuth {
		return fmt.Errorf("imageRegistryAuth has incorrect value: %q. Only \"\" and %q are allowed", o.ImageRegistryAuth, googleImageRegistryAuth)
	}

	return nil
}

func parseOptions() (*options, *bumper.Options, error) {
	var config string
	var labelsOverride []string
	var skipPullRequest bool
	var signoff bool

	var o options
	flag.StringVar(&config, "config", "", "The path to the config file for the autobumber.")
	flag.StringSliceVar(&labelsOverride, "labels-override", nil, "Override labels to be added to PR.")
	flag.BoolVar(&skipPullRequest, "skip-pullrequest", false, "")
	flag.BoolVar(&signoff, "signoff", false, "Signoff the commits.")
	flag.Parse()

	var pro bumper.Options
	data, err := os.ReadFile(config)
	if err != nil {
		return nil, nil, fmt.Errorf("read %q: %w", config, err)
	}

	if err = yaml.Unmarshal(data, &o); err != nil {
		return nil, nil, fmt.Errorf("unmarshal %q: %w", config, err)
	}

	if err := yaml.Unmarshal(data, &pro); err != nil {
		return nil, nil, fmt.Errorf("unmarshal %q: %w", config, err)
	}

	if labelsOverride != nil {
		pro.Labels = labelsOverride
	}
	pro.SkipPullRequest = skipPullRequest
	return &o, &pro, nil
}

// runAutobumper is wrapper for bumper API -> ACL
func runAutobumper(autoBumperCfg string) error {
	data, err := os.ReadFile(autoBumperCfg)
	if err != nil {
		return fmt.Errorf("open autobumper config: %s", err)
	}

	_, pro, _ := parseOptions()

	var bumperClientOpt bumper.Options
	err = yaml.Unmarshal(data, &bumperClientOpt)
	if err != nil {
		return fmt.Errorf("decode autobumper config: %s", err)
	}

	var opts bumper.Options
	err = yaml.Unmarshal(data, &opts)
	if err != nil {
		return fmt.Errorf("decode bumper options: %s", err)
	}

	if err := validateOptions(opts); err != nil {
		logrus.WithError(err).Fatalf("Failed validating flags")
	}

	ctx := context.Background()
	return bumper.Run(ctx, pro, &client{o: &bumperClientOpt})
}
