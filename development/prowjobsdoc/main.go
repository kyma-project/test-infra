package main

import (
	"flag"
	"fmt"
	jobs "github.com/kyma-project/test-infra/development/prowjobsdoc/inventory"
	"gopkg.in/yaml.v2"
	configflagutil "k8s.io/test-infra/prow/flagutil/config"
	"log"
	"os"
	"path/filepath"
)

type opts struct {
	config configflagutil.ConfigOptions
	output string
}

func gatherOptions() opts {
	var o opts
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&o.output, "output", "inventory.yaml", "Path to the output file")
	o.config.AddFlags(fs)
	fs.Parse(os.Args[1:])
	return o
}

func main() {
	o := gatherOptions()
	fmt.Println(o.config.ConfigPath, o.config.JobConfigPath)
	ca, err := o.config.ConfigAgent()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	conf := ca.Config()
	repos := conf.AllRepos.List()

	inv := jobs.Inventory{
		Jobs: make(map[string]jobs.OrgRepoJobs),
	}

	for _, r := range repos {
		orgRepoJobs := jobs.OrgRepoJobs{}
		pre := conf.AllStaticPresubmits([]string{r})
		for _, j := range pre {
			d := jobs.OrgRepoJob{
				Name:        j.Name,
				Description: j.Annotations["description"],
				Branches:    j.Branches,
			}
			for _, ref := range j.ExtraRefs {
				d.ExtraRefs = append(d.ExtraRefs, jobs.Refs{
					BaseRef: ref.BaseRef,
					Repo:    ref.Repo,
				})
			}

			orgRepoJobs.Presubmits = append(orgRepoJobs.Presubmits, d)
		}
		post := conf.AllStaticPostsubmits([]string{r})
		for _, j := range post {
			d := jobs.OrgRepoJob{
				Name:        j.Name,
				Description: j.Annotations["description"],
				Branches:    j.Branches,
			}
			for _, ref := range j.ExtraRefs {
				d.ExtraRefs = append(d.ExtraRefs, jobs.Refs{
					BaseRef: ref.BaseRef,
					Repo:    ref.Repo,
				})
			}
			orgRepoJobs.Postsubmits = append(orgRepoJobs.Postsubmits, d)
		}
		repoTotal := len(pre) + len(post)
		orgRepoJobs.Total = repoTotal

		inv.Jobs[r] = orgRepoJobs
		inv.Total += repoTotal
	}

	periodics := conf.AllPeriodics()
	var per jobs.Periodics
	for _, j := range periodics {
		d := jobs.OrgRepoJob{
			Name:        j.Name,
			Description: j.Annotations["description"],
		}
		for _, ref := range j.ExtraRefs {
			d.ExtraRefs = append(d.ExtraRefs, jobs.Refs{
				BaseRef: ref.BaseRef,
				Repo:    ref.Repo,
			})
		}
		per.Jobs = append(per.Jobs, d)
	}
	per.Total = len(periodics)
	inv.Total += per.Total
	inv.Periodics = &per

	fpath := filepath.FromSlash(o.output)
	f, err := os.Create(fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	ye := yaml.NewEncoder(f)
	err = ye.Encode(inv)
	if err != nil {
		panic(err)
	}
}
