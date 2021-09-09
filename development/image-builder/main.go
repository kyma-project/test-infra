package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/containers/buildah"
	"github.com/containers/buildah/define"
	"github.com/containers/buildah/imagebuildah"
	"github.com/containers/common/pkg/config"
	"github.com/containers/storage/pkg/archive"
	bc "github.com/kyma-project/test-infra/development/image-builder/pkg/config"

	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
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
	c, err := bc.NewFromFile(*buildConfig)
	if err != nil {
		panic(err)
	}

	buildStoreOptions, err := storage.DefaultStoreOptions(unshare.IsRootless(), unshare.GetRootlessUID())
	buildStore, err := storage.GetStore(buildStoreOptions)
	if err != nil {
		panic(err)
	}

	for _, step := range c.Steps {
		fmt.Println(step.Name)
		for _, im := range step.Images {
			absCtx, err := filepath.Abs(im.Context)
			if err != nil {
				panic(err)
			}
			dockerfilePath := filepath.Join(absCtx, im.Dockerfile)
			bo := define.BuildOptions{
				CommonBuildOpts:  &define.CommonBuildOptions{},
				ContextDirectory: absCtx,
				PullPolicy:       define.PullIfMissing,
				Compression:      archive.Gzip,
				CNIConfigDir:     define.DefaultCNIConfigDir,
				CNIPluginPath:    define.DefaultCNIPluginPath,
				AdditionalTags:   []string{im.Tag},
				Out:              os.Stdout,
				Err:              os.Stderr,
				ReportWriter:     os.Stdout,
				Args:             map[string]string{"commit": "dupa"},
				//Labels: []string{"test-label"},
				//Annotations: []string{"test-annotation"},
				//Layers: true,
				//NoCache: false,
				Excludes: nil,
				Runtime:  Runtime(),
			}
			imageID, ref, err := imagebuildah.BuildDockerfiles(context.TODO(), buildStore, bo, dockerfilePath)
			if err != nil {
				panic(err)
			}
			fmt.Println(imageID, ref)
		}
	}
	//bo := buildah.BuilderOptions{
	//	FromImage: "alpine:latest",
	//	Isolation:        define.IsolationChroot,
	//	CommonBuildOpts:  &define.CommonBuildOptions{},
	//	ConfigureNetwork: define.NetworkDefault,
	//}
	//bu, err := buildah.NewBuilder(context.TODO(), buildStore, bo)
	//if err != nil {
	//	panic(err)
	//}

	//bu.Add("/bin/", false, buildah.AddAndCopyOptions{}, "echo", "sleep", "sh")
	//bu.SetEnv("PATH", "/bin")
	//bu.Run([]string{"echo", `"dupa123"`}, buildah.RunOptions{})
	//bu.SetCmd([]string{"sleep", "60"})
	//ir, _ := is.Transport.ParseStoreReference(buildStore, "test:latest")
	//bu.Commit(context.TODO(), ir, buildah.CommitOptions{})
	//ims, _ := buildStore.Images()
	//fmt.Println(ims)
}

func Runtime() string {
	runtime := os.Getenv("BUILDAH_RUNTIME")
	if runtime != "" {
		return runtime
	}

	conf, err := config.Default()
	if err != nil {
		return define.DefaultRuntime
	}
	return conf.Engine.OCIRuntime
}
