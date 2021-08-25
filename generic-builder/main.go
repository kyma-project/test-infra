package main

import (
	"flag"
	"fmt"
	"github.com/containers/buildah"
	"github.com/containers/storage/pkg/unshare"
	"github.com/kyma-project/test-infra/generic-builder/pkg/config"
	"github.com/openshift/imagebuilder"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if buildah.InitReexec() {
		return
	}
	unshare.MaybeReexecUsingUserNamespace(false)

	buildConfig := flag.String("build-config", "", "Configuration file for image build rules")
	flag.Parse()
	c, err := config.NewFromFile(*buildConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)

	//buildStoreOptions, err := storage.DefaultStoreOptions(unshare.IsRootless(), unshare.GetRootlessUID())
	//_, err = storage.GetStore(buildStoreOptions)
	//if err != nil {
	//	panic(err)
	//}

	for _, stage := range c.Stages {
		var dockerfiles []io.ReadCloser
		for _, im := range stage.Images {
			f, _ := os.Open(filepath.Join(im.Context, "Dockerfile"))
			dockerfiles = append(dockerfiles, f)
			n, _ := imagebuilder.ParseDockerfile(f)
			fmt.Printf("%+v\n", n)
		}
		fmt.Printf("%+v\n", dockerfiles)
	}
}
