package component

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	componentarchiveremote "github.com/gardener/component-cli/pkg/commands/componentarchive/remote"
	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/kyma-project/test-infra/development/image-url-helper/pkg/list"
	"github.com/mandelsoft/vfs/pkg/osfs"
)

type ComponentOptions struct {
	Provider         string
	ComponentName    string // github.com/kyma-project/kyma
	ComponentVersion string
	AppName          string
	OutputDir        string
	RepoContext      string
	GitCommit        string
	GitBranch        string
}

func GenerateComponentDescriptor(options ComponentOptions, images list.ImageMap) (*v2.ComponentDescriptor, error) {
	component := &v2.ComponentDescriptor{}

	err := createComponent(component, options)
	if err != nil {
		return component, err
	}

	err = addSources(component, options)
	if err != nil {
		return component, err
	}

	err = addResources(component, options, images)
	if err != nil {
		return component, err
	}

	//addComponentReferences()

	return component, nil
}

func createComponent(component *v2.ComponentDescriptor, options ComponentOptions) error {
	// set default values
	v2.DefaultComponent(component)
	component.Metadata.Version = "v2"

	if options.Provider == "internal" {
		component.Provider = v2.InternalProvider
	} else if options.Provider == "external" {
		component.Provider = v2.ExternalProvider
	} else {
		return fmt.Errorf("unknown provider value: %s", options.Provider)
	}
	component.SetName(options.ComponentName)
	component.SetVersion(options.ComponentVersion)
	return nil
}

func addSources(component *v2.ComponentDescriptor, options ComponentOptions) error {

	source := v2.Source{}
	// TODO DO NOT hardcode this
	source.Name = "kyma-project_kyma"
	source.Version = options.ComponentVersion
	source.Type = "git"
	accessData := make(map[string]interface{})
	// TODO again, do not hardcode
	accessData["repoUrl"] = "https://github.com/kyma-project/kyma"
	accessData["ref"] = options.GitBranch
	accessData["commit"] = options.GitCommit

	source.Access = v2.NewUnstructuredType("github", accessData)

	component.Sources = append(component.Sources, source)

	return nil
}

func addResources(component *v2.ComponentDescriptor, options ComponentOptions, images list.ImageMap) error {
	for _, image := range images {
		// TODO ugly hack, since istio is still in wrong format
		if strings.HasPrefix(image.FullImageURL(), "eu.gcr.io/kyma-project/external/istio") {
			continue
		}

		resource := v2.Resource{}

		resource.Version = options.ComponentVersion
		resource.Type = "ociImage"
		resource.Relation = v2.LocalRelation

		imageReference, err := name.ParseReference(image.FullImageURL())
		if err != nil {
			return err
		}

		imageInfo, err := remote.Image(imageReference)
		if err != nil {
			return err
		}

		imageHash, err := imageInfo.Digest()
		if err != nil {
			return err
		}

		resource.Name = strings.Replace(imageReference.Context().RepositoryStr(), "/", "_", -1)
		resource.Name = strings.Replace(resource.Name, ".", "_", -1)

		accessData := make(map[string]interface{})
		accessData["imageReference"] = imageReference.Context().RegistryStr() + "/" + imageReference.Context().RepositoryStr() + "@" + imageHash.String()

		resource.Access = v2.NewUnstructuredType("ociRegistry", accessData)

		component.Resources = append(component.Resources, resource)
	}

	return nil
}

func SanityCheck(encodedComponentDescriptor []byte) error {
	var decoded v2.ComponentDescriptor
	err := codec.Decode(encodedComponentDescriptor, &decoded)
	return err
}

func PushDescriptor(encodedComponentDescriptor []byte, repoContext string) error {
	// create temporary dir, so pushing can be separate from saving YAML file
	dirPath, err := ioutil.TempDir(os.TempDir(), "component_descriptor")
	if err != nil {
		return err
	}

	err = os.MkdirAll(dirPath, 0666)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dirPath)

	filePath := dirPath + "/component-descriptor.yaml"
	tempFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	_, err = tempFile.Write(encodedComponentDescriptor)
	if err != nil {
		return err
	}
	tempFile.Close()

	pushOptions := componentarchiveremote.PushOptions{}
	pushOptions.ComponentArchivePath = dirPath
	pushOptions.BuilderOptions.BaseUrl = repoContext

	err = pushOptions.Run(context.Background(), logr.Discard(), osfs.New())
	if err != nil {
		return err
	}

	return err
}
