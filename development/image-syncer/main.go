package main

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	log = logrus.New()
)

// SyncDef stores synchronisation definition
type SyncDef struct {
	TargetRepoPrefix string `yaml:"targetRepoPrefix"`
	Sign             bool   `yaml:"sign,omitempty"`
	Images           []Image
}

// Image stores image location
type Image struct {
	Source string
	Sign   bool   `yaml:"sign,omitempty"`
	Tag    string `yaml:"tag,omitempty"`
}

// Config stores command line arguments
type Config struct {
	ImagesFile    string
	TargetKeyFile string
	DryRun        bool
	Debug         bool
}

//TODO: we should check for error type not match strings
func isImageNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Contains(err.Error(), "MANIFEST_UNKNOWN") {
		return true
	}
	return false
}

// cancelOnInterrupt calls cancel func when os.Interrupt or SIGTERM is received
func cancelOnInterrupt(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-c:
			cancel()
		}
	}()
}

func SyncImage(ctx context.Context, src, dest string, dryRun bool, opts ...crane.Option) error {
	log.Debug("source ", src)
	log.Debug("destination ", dest)
	srcImg, err := crane.Pull(src)
	if err != nil {
		return fmt.Errorf("source image pull error: %w", err)
	}
	destImg, err := crane.Pull(dest)
	if isImageNotFoundError(err) {
		log.Debug("Target image does not exist. Pushing image...")
		if dryRun {
			log.Info("Dry-Run enabled. Skipping push.")
			return nil
		}
		err := crane.Push(srcImg, dest, opts...)
		if err != nil {
			return fmt.Errorf("push image error: %w", err)
		}
		return nil
	} else if err != nil {
		return err
	}
	d1, err := srcImg.Digest()
	log.Debug("source digest ", d1)
	if err != nil {
		return err
	}
	d2, err := destImg.Digest()
	log.Debug("target digest ", d2)
	if err != nil {
		return err
	}

	if d1 != d2 {
		return fmt.Errorf("digests are not equal: %v, %v - probably source image has been changed", d1, d2)
	}
	log.Debug("Digests are equal. Nothing to do.")
	return nil
}

func GetTarget(source, targetRepo, targetTag string) (string, error) {
	target := targetRepo + source
	if strings.Contains(source, "@sha256:") {
		if targetTag == "" {
			return "", fmt.Errorf("sha256 digest detected, but the \"tag\" was not specified")
		}
		imageName := strings.Split(source, "@sha256:")[0]
		target = targetRepo + imageName + ":" + targetTag
	}
	return target, nil
}

func ParseImagesFile(file string) (*SyncDef, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var syncDef SyncDef
	if err := yaml.Unmarshal(f, &syncDef); err != nil {
		return nil, err
	}
	if syncDef.TargetRepoPrefix == "" {
		return nil, fmt.Errorf("targetRepoPrefix can not be empty")
	}
	return &syncDef, nil
}

func main() {
	log.Out = os.Stdout
	var cfg Config

	var rootCmd = &cobra.Command{
		Use:   "image-syncer",
		Short: "image-syncer copies images between docker registries",
		Long:  `image-syncer copies docker images. It compares checksum between source and target and protects target images against overriding`,
		Run: func(cmd *cobra.Command, args []string) {
			logLevel := logrus.InfoLevel
			if cfg.Debug {
				logLevel = logrus.DebugLevel
			}
			log.SetLevel(logLevel)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cancelOnInterrupt(ctx, cancel)

			imagesFile, err := ParseImagesFile(cfg.ImagesFile)
			if err != nil {
				log.WithError(err).Fatal("Could not parse images file")
			}
			authCfg, err := ioutil.ReadFile(cfg.TargetKeyFile)
			if err != nil {
				log.WithError(err).Fatal("Could not open target auth key JSON")
			}
			authOption := crane.WithAuth(&authn.Basic{Username: "_json_key", Password: string(authCfg)})
			var failed bool

		loop:
			for _, img := range imagesFile.Images {
				select {
				case <-ctx.Done():
					log.Error("Context cancelled")
					break loop
				default:
				}
				log.WithField("image", img.Source).Info("Start scan on image")
				target, err := GetTarget(img.Source, imagesFile.TargetRepoPrefix, img.Tag)
				if err != nil {
					log.WithError(err).Error("Failed to get target url")
					failed = true
				}
				err = SyncImage(ctx, img.Source, target, cfg.DryRun, authOption)
				if err != nil {
					log.WithError(err).Error("Failed to sync image")
					failed = true
				}
				if err == nil {
					log.WithField("target", target).Info("Image synced succesfully")
				}
			}
			if failed {
				log.Fatal("Some errors occurred during image sync")
			} else {
				log.Info("All images synced successfully")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.ImagesFile, "images-file", "i", "", "yaml file containing list of images")
	rootCmd.PersistentFlags().StringVarP(&cfg.TargetKeyFile, "target-repo-auth-key", "t", "", "JSON key file used for authorization to target repo")
	rootCmd.PersistentFlags().BoolVarP(&cfg.DryRun, "dry-run", "d", true, "dry run mode")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "enable debug mode")

	rootCmd.MarkPersistentFlagRequired("images-file")
	rootCmd.MarkPersistentFlagRequired("target-repo-auth-key")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "SYNCER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
