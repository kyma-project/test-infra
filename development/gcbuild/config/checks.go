package config

import (
	"fmt"
	errutil "k8s.io/apimachinery/pkg/util/errors"
	"regexp"
	"strings"
)

const (
	ErrMissingRepository = "missing-repository"
	ErrMissingTag        = "missing-tag"
	ErrMissingVariants   = "missing-variants"
)

var defaultErrs = []string{
	ErrMissingTag,
	ErrMissingRepository,
	ErrMissingVariants,
}

func ValidateConfig(enabledErrs []string, c *CloudBuild, vs Variants) error {
	checks := defaultErrs
	if len(enabledErrs) > 0 {
		checks = enabledErrs
	}

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
		case ErrMissingVariants:
			errs = append(errs, validateVariants(c, vs))
		}
	}

	return errutil.NewAggregate(errs)
}

// validateTag checks parsed config file if it includes $_TAG substitution in the yaml.
// The tool requires that cloudbuild.yaml uses _TAG substitution when tagging image.
// This check ensures that _TAG is present in at least one of the steps as argument
// and in the 'images' field.
func validateTag(c *CloudBuild) error {
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
		errs = append(errs, fmt.Errorf("images: missing _TAG substitution in 'images', add image with tag as _TAG to the 'images' field"))
	}
	return errutil.NewAggregate(errs)
}

// validateRepository checks, if the _REPOSITORY substitution is present in the 'substitutions' field
// in the parsed cloudbuild.yaml file.
// The tool requires this substitution as a default value defined in the config.
func validateRepository(c *CloudBuild) error {
	if len(c.Substitutions) == 0 {
		return fmt.Errorf("'substitutions' field in cloudbuild.yaml file is empty, in order for gcbuild to run properly you must define this field with at least _REPOSITORY variable")
	}
	if _, ok := c.Substitutions["_REPOSITORY"]; !ok {
		return fmt.Errorf("missing _REPOSITORY in 'substitutions' field in cloudbuild.yaml file, add this variable to the 'substitutions' field with default push repository")
	}
	return nil
}

// validateVariants checks, if the config file contains $_VARIANTS substitution
// if the 'variants.yaml' file is present
// If the 'variants.yaml' file is present, but no $_VARIANT substitution is available,
// then variants will be pushed under the same tag, overriding the image.
func validateVariants(c *CloudBuild, vs Variants) error {
	var hasVariant, fileNotExists bool
	fileNotExists = vs == nil
	for _, i := range c.Images {
		if strings.Contains(i, "$_VARIANT") || strings.Contains(i, "${_VARIANT}") {
			hasVariant = true
			break
		}
	}
	if !hasVariant && !fileNotExists {
		return fmt.Errorf("your directory has 'variants.yaml' file present, but there is no $_VARIANT substitution provided in config, add $_VARIANT substitution to the config to allow building variants of the image")
	}
	if hasVariant && fileNotExists {
		return fmt.Errorf("your config has $_VARIANT substitution provided, but you do not use any 'variants.yaml' file")
	}
	return nil
}
