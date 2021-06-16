package rendertemplates

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"

	"github.com/Masterminds/semver"
	"github.com/forestgiant/sliceutil"
	"github.com/imdario/mergo"
)

// Config represents configuration of all templates to render along with global values
type Config struct {
	Templates  []TemplateConfig
	Global     map[string]interface{}
	GlobalSets map[string]ConfigSet `yaml:"globalSets,omitempty"`
}

// TemplateConfig specifies template to use and files to render
type TemplateConfig struct {
	From   string
	Render []RenderConfig
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

// InheritedConfigs specify named configs to use for generating prowjob from template
type InheritedConfigs struct {
	Global      []string `yaml:"global,omitempty"`
	Local       []string `yaml:"local,omitempty"`
	PreConfigs  []string `yaml:"preConfigs,omitempty"`
	PostConfigs []string `yaml:"postConfigs,omitempty"`
}

// Job holds data for generating prowjob from template
type Job struct {
	InheritedConfigs InheritedConfigs `yaml:"inheritedConfigs,omitempty"`
	JobConfig        ConfigSet        `yaml:"jobConfig,omitempty"`
	jobConfigPre     ConfigSet
	jobConfigPost    ConfigSet
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
	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}
	return copy, nil
}

// MergeConfigs merges values from GlobalSets, LocalSets and local values for each job
func (r *RenderConfig) MergeConfigs(config *Config) {
	globalConfigSets := config.GlobalSets
	if present := len(r.JobConfigs); present > 0 {
		r.Values = make(map[string]interface{})
		for repoIndex, repo := range r.JobConfigs {
			for jobIndex, job := range repo.Jobs {
				jobConfig := ConfigSet{}
				jobConfigPre := ConfigSet{}
				jobConfigPost := ConfigSet{}

				if sliceutil.Contains(job.InheritedConfigs.Global, "default") {
					if err := jobConfig.mergeConfigSet(deepCopyConfigSet(globalConfigSets["default"])); err != nil {
						log.Fatalf("Failed merge Global default configSet: %s", err)
					}

				}
				if sliceutil.Contains(job.InheritedConfigs.Local, "default") {
					if err := jobConfig.mergeConfigSet(deepCopyConfigSet(r.LocalSets["default"])); err != nil {
						log.Fatalf("Failed merge Local default configSet: %s", err)
					}
				}
				for _, v := range job.InheritedConfigs.Global {
					if v != "default" {
						if err := jobConfig.mergeConfigSet(deepCopyConfigSet(globalConfigSets[v])); err != nil {
							log.Fatalf("Failed merge global %s named configset: %s", v, err)
						}
					}
				}

				if len(job.InheritedConfigs.PreConfigs) > 0 {
					jobConfigPre = deepCopyConfigSet(jobConfig)
					for _, v := range job.InheritedConfigs.PreConfigs {
						if err := jobConfigPre.mergeConfigSet(deepCopyConfigSet(globalConfigSets[v])); err != nil {
							log.Fatalf("Failed merge global %s named configset: %s", v, err)
						}
					}
				}
				if len(job.InheritedConfigs.PostConfigs) > 0 {
					jobConfigPost = deepCopyConfigSet(jobConfig)
					for _, v := range job.InheritedConfigs.PostConfigs {
						if err := jobConfigPost.mergeConfigSet(deepCopyConfigSet(globalConfigSets[v])); err != nil {
							log.Fatalf("Failed merge global %s named configset: %s", v, err)
						}
					}
				}

				for _, v := range job.InheritedConfigs.Local {
					if v != "default" {
						if err := jobConfig.mergeConfigSet(deepCopyConfigSet(r.LocalSets[v])); err != nil {
							log.Fatalf("Failed merge local %s named configset: %s", v, err)
						}
					}
				}
				if err := jobConfig.mergeConfigSet(job.JobConfig); err != nil {
					log.Fatalf("Failed merge job configset %s", err)
				}

				if len(jobConfigPre) > 0 {
					if err := jobConfigPre.mergeConfigSet(job.JobConfig); err != nil {
						log.Fatalf("Failed merge job configset %s", err)
					}
				}
				if len(jobConfigPost) > 0 {
					if err := jobConfigPost.mergeConfigSet(job.JobConfig); err != nil {
						log.Fatalf("Failed merge job configset %s", err)
					}
				}

				r.JobConfigs[repoIndex].Jobs[jobIndex].JobConfig = jobConfig
				if len(jobConfigPre) > 0 {
					r.JobConfigs[repoIndex].Jobs[jobIndex].jobConfigPre = jobConfigPre
				}
				if len(jobConfigPost) > 0 {
					r.JobConfigs[repoIndex].Jobs[jobIndex].jobConfigPost = jobConfigPost
				}
			}
		}
		r.Values["JobConfigs"] = r.JobConfigs
	}
}

func deepCopyConfigSet(configSet ConfigSet) ConfigSet {
	dst, err := Map(configSet)
	if err != nil {
		log.Fatalf("Failed ConfigSet deepCopy with error: %s", err)
	}
	return dst
}

func (j *ConfigSet) mergeConfigSet(configSet ConfigSet) error {
	if len(configSet) == 0 {
		return errors.New("configSet not found")
	}
	if err := mergo.Merge(j, configSet, mergo.WithOverride); err != nil {
		return err
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
