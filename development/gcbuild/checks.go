package main

import (
	"fmt"
	errutil "k8s.io/apimachinery/pkg/util/errors"
	"regexp"
)

const (
	ErrMissingRepository = "missing-repository"
	ErrMissingTag        = "missing-tag"
)

var defaultErrs = []string{
	ErrMissingTag,
	ErrMissingRepository,
}

func validateConfig(c *Config) error {

	checks := defaultErrs

	var errs []error
	//
	// sanity checks for cloudbuild yaml file
	//
	for _, ch := range checks {
		switch ch {
		case ErrMissingTag:
			errs = append(errs, validateTag(c))
		case ErrMissingRepository:
			errs = append(errs, validateRepository(c))
		}
	}

	return errutil.NewAggregate(errs)
}

// validateTag checks parsed config file if it includes $_TAG substitution in the yaml.
// The tool requires that cloudbuild.yaml uses _TAG substitution when tagging image.
// This check ensures that _TAG is present in at least one of the steps as argument
// and in the 'images' field.
func validateTag(c *Config) error {
	var presentInArgs bool
	r := regexp.MustCompile(`(\$_TAG)|(\${_TAG})`)
	for _, step := range c.Steps {
		for _, s := range step.Args {
			if r.MatchString(s) {
				presentInArgs = true
				break
			}
		}
	}
	var presentInImages bool
	for _, i := range c.Images {
		if r.MatchString(i) {
			presentInImages = true
			break
		}
	}

	var errs []error
	if !presentInArgs {
		errs = append(errs, fmt.Errorf("steps: missing _TAG substitution in 'args', define at least one step that build image with tag as _TAG substitution"))
	}
	if !presentInImages {
		errs = append(errs, fmt.Errorf("images: missing _TAG substitution in 'images', add image with tag as _TAG to the 'images' feld"))
	}
	return errutil.NewAggregate(errs)
}

// validateRepository checks, if the _REPOSITORY substitution is present in the 'substitutions' field
// in the parsed cloudbuild.yaml file.
// The tool requires this substitution as a default value defined in the config.
func validateRepository(c *Config) error {
	if len(c.Substitutions) == 0 {
		return fmt.Errorf("'substitutions' field is empty")
	}
	if _, ok := c.Substitutions["_REPOSITORY"]; !ok {
		return fmt.Errorf("missing _REPOSITORY in 'substitutions' field")
	}
	return nil
}
