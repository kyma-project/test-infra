package rendertemplates

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	"github.com/forestgiant/sliceutil"
	"github.com/imdario/mergo"
	"github.com/jinzhu/copier"
)

// Config represents configuration of all templates to render along with global values
type Config struct {
	TemplatesConfigs []*TemplateConfig `yaml:"templates,omitempty"`
	Global           map[string]interface{}
	GlobalSets       map[string]ConfigSet `yaml:"globalSets,omitempty"`
}

// TemplateConfig specifies template to use and files to render
type TemplateConfig struct {
	FromTo        []FromTo        `yaml:"fromTo,omitempty"`
	From          string          `yaml:"from,omitempty"`
	RenderConfigs []*RenderConfig `yaml:"render,omitempty"`
}

// FromTo defines what template should be used and where to store the render output
type FromTo struct {
	From string
	To   string
}

// RenderConfig specifies where to render template and values to use
type RenderConfig struct {
	To         string
	Values     map[string]interface{}
	LocalSets  map[string]ConfigSet `yaml:"localSets,omitempty"`
	JobConfigs []Repo               `yaml:"jobConfigs,omitempty"`
}

// ConfigSet hold set of data for generating prowjob from template
type ConfigSet map[string]interface{}

// Repo represent github repository with associated prowjobs data
type Repo struct {
	RepoName string `yaml:"repoName,omitempty"`
	Jobs     []Job  `yaml:"jobs,omitempty"`
}

// InheritedConfigsElement specify named configs to use for generating component prowjob from template
type InheritedConfigsElement struct {
	Global []string `yaml:"global,omitempty"`
	Local  []string `yaml:"local,omitempty"`
}

// InheritedConfigs specify named configs to use for generating prowjob from template
type InheritedConfigs struct {
	Global      []string                `yaml:"global,omitempty"`
	Local       []string                `yaml:"local,omitempty"`
	PreConfigs  InheritedConfigsElement `yaml:"preConfigs,omitempty"`
	PostConfigs InheritedConfigsElement `yaml:"postConfigs,omitempty"`
}

// Job holds data for generating prowjob from template
type Job struct {
	InheritedConfigs InheritedConfigs `yaml:"inheritedConfigs,omitempty"`
	JobConfig        ConfigSet        `yaml:"jobConfig,omitempty"`
	JobConfigPre     ConfigSet        `yaml:"jobConfigPre,omitempty"`
	JobConfigPost    ConfigSet        `yaml:"jobConfigPost,omitempty"`
}

// Map performs a deep copy of the given map m.
func Map(m map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	var deepcopy map[string]interface{}
	err = dec.Decode(&deepcopy)
	if err != nil {
		return nil, err
	}
	return deepcopy, nil
}

// Merge merges all jobconfigs using local / globalsets defined in the configuration
func (cfg *Config) Merge(mergoConfig mergo.Config) {
	cfg.TemplatesConfigs = generateFromTo(cfg.TemplatesConfigs)

	for _, templateConfig := range cfg.TemplatesConfigs {
		templateConfig.generateRenderConfigs(cfg, mergoConfig)
	}

	cfg.TemplatesConfigs = mergeRenderDestinations(cfg.TemplatesConfigs)
}

// generateFromTo for datafiles without FromTo a new FromTo will be created from From and To fields
func generateFromTo(templatesConfigs []*TemplateConfig) []*TemplateConfig {
	var tmpls []*TemplateConfig
	for _, templateConfig := range templatesConfigs {
		if len(templateConfig.FromTo) == 0 {
			if templateConfig.From != "" {
				for _, renderConfig := range templateConfig.RenderConfigs {
					fromTo := FromTo{
						From: templateConfig.From,
						To:   renderConfig.To,
					}
					var rc RenderConfig
					err := copier.CopyWithOption(&rc, &renderConfig, copier.Option{DeepCopy: true})
					if err != nil {
						log.Fatalf("Cannot deepcopy object: %s", err)
					}
					tmpls = append(tmpls, &TemplateConfig{
						FromTo:        []FromTo{fromTo},
						RenderConfigs: []*RenderConfig{&rc},
					})
				}
			}
		} else {
			tmpls = append(tmpls, templateConfig)
		}
	}
	return tmpls
}

// mergeRenderDestinations merges and deduplicates renderconfigurations for destinations so that a file can be used as a target in multiple data files
func mergeRenderDestinations(templatesConfigs []*TemplateConfig) []*TemplateConfig {
	tmpl := make(map[FromTo]*TemplateConfig)
	for _, templateConfig := range templatesConfigs {
		for _, fromTo := range templateConfig.FromTo {
			if template, ok := tmpl[fromTo]; ok {
				reposDst, ok := template.RenderConfigs[0].Values["JobConfigs"].([]Repo)
				if !ok {
					log.Fatalf("dst JobConfigs not of Type []Repo")
				}
				reposSrc, ok := templateConfig.RenderConfigs[0].Values["JobConfigs"].([]Repo)
				if !ok {
					log.Fatalf("src JobConfigs not of Type []Repo")
				}
				reposDst = mergeRepos(reposDst, reposSrc)

				template.RenderConfigs[0].Values["JobConfigs"] = reposDst
			} else {
				var tplCfg TemplateConfig
				if err := copier.CopyWithOption(&tplCfg, templateConfig, copier.Option{DeepCopy: true}); err != nil {
					log.Fatalf("cannot deepcopy object: %s", err)
				}
				tmpl[fromTo] = &tplCfg
				tmpl[fromTo].FromTo = []FromTo{fromTo}
			}
		}
	}
	v := make([]*TemplateConfig, 0)
	for _, value := range tmpl {
		v = append(v, value)
	}
	return v
}

func mergeRepos(dst []Repo, src []Repo) []Repo {
	repos := make(map[string]Repo)

	for _, item := range dst {
		repos[item.RepoName] = item
	}

	for _, item := range src {
		if _, ok := repos[item.RepoName]; !ok {
			repos[item.RepoName] = item
		} else {
			repo := repos[item.RepoName]
			repo.Jobs = append(repo.Jobs, item.Jobs...)
			repos[item.RepoName] = repo
		}
	}

	v := make([]Repo, 0)
	for _, value := range repos {
		v = append(v, value)
	}
	return v
}

func (ft FromTo) String() string {
	return fmt.Sprintf("%s -> %s", ft.From, ft.To)
}

// generateRenderConfigs merges parts, generates component jobs and appends all jobs to the list of values
func (tplCfg *TemplateConfig) generateRenderConfigs(config *Config, mergoConfig mergo.Config) {
	for _, render := range tplCfg.RenderConfigs {
		// merge all parts of a config
		render.generateJobConfigs(config.GlobalSets, mergoConfig)
		// generate component jobs
		render.GenerateComponentJobs(config.Global)
		// append all jobs to the list of values for the template
		render.AppendJobs()
	}
}

// generateJobConfigs merges values from GlobalSets, LocalSets and local values for each job
func (r *RenderConfig) generateJobConfigs(globalConfigSets map[string]ConfigSet, mergoConfig mergo.Config) {
	if present := len(r.JobConfigs); present > 0 {
		r.Values = make(map[string]interface{})
		for repoIndex, repo := range r.JobConfigs {
			for jobIndex, job := range repo.Jobs {

				jobConfig := ConfigSet{}
				jobConfigPre := ConfigSet{}
				jobConfigPost := ConfigSet{}
				generatePresubmitJob := false
				generatePostsubmitJob := false

				// merge "default" global inheritedConfig to jobConfig
				if sliceutil.Contains(job.InheritedConfigs.Global, "default") {
					if err := jobConfig.mergeConfigSet(globalConfigSets["default"], mergoConfig); err != nil {
						log.Fatalf("Failed merge Global default configSet: %s", err)
					}

				}
				// merge "default" local inheritedConfig to jobConfig
				if sliceutil.Contains(job.InheritedConfigs.Local, "default") {
					if err := jobConfig.mergeConfigSet(r.LocalSets["default"], mergoConfig); err != nil {
						log.Fatalf("Failed merge Local default configSet: %s", err)
					}
				}
				// merge global inheritedConfigs to jobConfig
				for _, v := range job.InheritedConfigs.Global {
					if v != "default" {
						if err := jobConfig.mergeConfigSet(globalConfigSets[v], mergoConfig); err != nil {
							log.Fatalf("Failed merge global %s named configset: %s", v, err)
						}
					}
				}

				// check if we should generate jobConfigPre/jobConfigPost and create then from jobConfig, so they'll already have default and global inheritedConfigs
				if len(job.InheritedConfigs.PreConfigs.Global) > 0 || len(job.InheritedConfigs.PreConfigs.Local) > 0 || len(job.JobConfigPre) > 0 {
					generatePresubmitJob = true
					jobConfigPre = deepCopyConfigSet(jobConfig)
				}
				if len(job.InheritedConfigs.PostConfigs.Global) > 0 || len(job.InheritedConfigs.PostConfigs.Local) > 0 || len(job.JobConfigPost) > 0 {
					generatePostsubmitJob = true
					jobConfigPost = deepCopyConfigSet(jobConfig)
				}

				// if global precommit InheritedConfigs exist, merge them to jobConfigPre
				if len(job.InheritedConfigs.PreConfigs.Global) > 0 {
					for _, v := range job.InheritedConfigs.PreConfigs.Global {
						if err := jobConfigPre.mergeConfigSet(globalConfigSets[v], mergoConfig); err != nil {
							log.Fatalf("Failed merge global %s named configset: %s", v, err)
						}
					}
				}

				// if global postcommit InheritedConfigs exist, merge them to jobConfigPost
				if len(job.InheritedConfigs.PostConfigs.Global) > 0 {
					for _, v := range job.InheritedConfigs.PostConfigs.Global {
						if err := jobConfigPost.mergeConfigSet(globalConfigSets[v], mergoConfig); err != nil {
							log.Fatalf("Failed merge global %s named configset: %s", v, err)
						}
					}
				}

				// merge local inheritedConfigs to jobConfig
				for _, v := range job.InheritedConfigs.Local {
					if v != "default" {
						if err := jobConfig.mergeConfigSet(r.LocalSets[v], mergoConfig); err != nil {
							log.Fatalf("Failed merge local %s named configset: %s", v, err)
						}
					}
				}

				if generatePresubmitJob {
					// merge local inheritedConfigs to jobConfigPre
					for _, v := range job.InheritedConfigs.Local {
						if v != "default" {
							if err := jobConfigPre.mergeConfigSet(r.LocalSets[v], mergoConfig); err != nil {
								log.Fatalf("Failed merge local %s named configset: %s", v, err)
							}
						}
					}
					// merge local precommit inheritedConfigs to jobConfigPre
					if len(job.InheritedConfigs.PreConfigs.Local) > 0 {
						for _, v := range job.InheritedConfigs.PreConfigs.Local {
							if err := jobConfigPre.mergeConfigSet(r.LocalSets[v], mergoConfig); err != nil {
								log.Fatalf("Failed merge local %s named configset: %s", v, err)
							}
						}
					}
				}

				if generatePostsubmitJob {
					// merge local inheritedConfigs to jobConfigPost
					for _, v := range job.InheritedConfigs.Local {
						if v != "default" {
							if err := jobConfigPost.mergeConfigSet(r.LocalSets[v], mergoConfig); err != nil {
								log.Fatalf("Failed merge local %s named configset: %s", v, err)
							}
						}
					}
					// merge local postcommit inheritedConfigs to jobConfigPost
					if len(job.InheritedConfigs.PostConfigs.Local) > 0 {
						for _, v := range job.InheritedConfigs.PostConfigs.Local {
							if err := jobConfigPost.mergeConfigSet(r.LocalSets[v], mergoConfig); err != nil {
								log.Fatalf("Failed merge local %s named configset: %s", v, err)
							}
						}
					}
				}

				// merge jobconfig to jobConfig
				if len(job.JobConfig) > 0 {
					if err := jobConfig.mergeConfigSet(job.JobConfig, mergoConfig); err != nil {
						log.Fatalf("Failed merge job configset %s", err)
					}
				}

				if generatePresubmitJob {
					// merge jobconfig to jobConfigPre
					if len(job.JobConfig) > 0 {
						if err := jobConfigPre.mergeConfigSet(job.JobConfig, mergoConfig); err != nil {
							log.Fatalf("Failed merge job configset: %s", err)
						}
					}
					// merge jobconfigPre to jobConfigPre
					if len(job.JobConfigPre) > 0 {
						if err := jobConfigPre.mergeConfigSet(job.JobConfigPre, mergoConfig); err != nil {
							log.Fatalf("Failed merge job configsetpre: %s", err)
						}
					}
				}

				if generatePostsubmitJob {
					// merge jobconfig to jobConfigPost
					if len(job.JobConfig) > 0 {
						if err := jobConfigPost.mergeConfigSet(job.JobConfig, mergoConfig); err != nil {
							log.Fatalf("Failed merge job configset: %s", err)
						}
					}
					// merge post jobconfigPost to jobConfigPost
					if len(job.JobConfigPost) > 0 {
						if err := jobConfigPost.mergeConfigSet(job.JobConfigPost, mergoConfig); err != nil {
							log.Fatalf("Failed merge job configsetpost: %s", err)
						}
					}
				}

				// add all generated jobs to the list of JobConfigs
				r.JobConfigs[repoIndex].Jobs[jobIndex].JobConfig = jobConfig
				if generatePresubmitJob {
					r.JobConfigs[repoIndex].Jobs[jobIndex].JobConfigPre = jobConfigPre
				}
				if generatePostsubmitJob {
					r.JobConfigs[repoIndex].Jobs[jobIndex].JobConfigPost = jobConfigPost
				}
			}
		}
	}
}

func deepCopyConfigSet(configSet ConfigSet) ConfigSet {
	dst, err := Map(configSet)
	if err != nil {
		log.Fatalf("Failed ConfigSet deepCopy with error: %s", err)
	}
	return dst
}

func (j *ConfigSet) mergeConfigSet(configSet ConfigSet, mergoConfig mergo.Config) error {
	if len(configSet) == 0 {
		return errors.New("configSet not found")
	}
	if mergoConfig.AppendSlice {
		if err := mergo.Merge(j, deepCopyConfigSet(configSet), mergo.WithOverride, mergo.WithAppendSlice); err != nil {
			return err
		}
	} else {
		if err := mergo.Merge(j, deepCopyConfigSet(configSet), mergo.WithOverride); err != nil {
			return err
		}
	}
	return nil
}

// MatchingReleases filters releases list against since and until releases
func MatchingReleases(allReleases []interface{}, since interface{}, until interface{}) []interface{} {
	result := make([]interface{}, 0)
	for _, rel := range allReleases {
		if ReleaseMatches(rel, since, until) {
			result = append(result, rel)
		}
	}
	return result
}

// ReleaseMatches checks if the release falls between since and until releases
func ReleaseMatches(rel interface{}, since interface{}, until interface{}) bool {
	relVer := semver.MustParse(rel.(string))
	if since != nil && relVer.Compare(semver.MustParse(since.(string))) < 0 {
		return false
	}
	if until != nil && relVer.Compare(semver.MustParse(until.(string))) > 0 {
		return false
	}
	return true
}

// AppendJobs appends data of presubmit/postsubmit/common jobs to the values list
func (r *RenderConfig) AppendJobs() {
	if present := len(r.JobConfigs); present > 0 {
		for repoIndex, repo := range r.JobConfigs {
			var jobs []Job

			for _, job := range repo.Jobs {

				// append the common job to the list
				if len(job.JobConfig) > 0 {
					jobs = append(jobs, job)
				}

				// append the presubmit job to the list
				if len(job.JobConfigPre) > 0 {
					preSubmit := Job{}
					preSubmit.JobConfig = deepCopyConfigSet(job.JobConfigPre)
					jobs = append(jobs, preSubmit)
				}

				// append the postsubmit job to the list
				if len(job.JobConfigPost) > 0 {
					postSubmit := Job{}
					postSubmit.JobConfig = deepCopyConfigSet(job.JobConfigPost)
					jobs = append(jobs, postSubmit)
				}
			}
			r.JobConfigs[repoIndex].Jobs = jobs
		}
		// copy the jobs to the values used by the templates
		r.Values["JobConfigs"] = r.JobConfigs
	}
}
