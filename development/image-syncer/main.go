package main

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/pkg/errors"
	"github.com/sigstore/sigstore/pkg/signature"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log = logrus.New()
)

// SyncDef stores synchronisation definition
type SyncDef struct {
	TargetRepoPrefix string `yaml:"targetRepoPrefix"`
	Sign             bool   `yaml:"sign"`
	Images           []Image
}

// Image stores image location
type Image struct {
	Source string
	Tag    string `yaml:"tag,omitempty"`
	Sign   *bool  `yaml:"sign,omitempty"`
}

// Config stores command line arguments
type Config struct {
	ImagesFile    string
	TargetKeyFile string
	KeyRef        string
	DryRun        bool
	Debug         bool
}

func ifRefNotFound(err error) bool {
	if err == nil {
		return false
	}
	var e *transport.Error
	if errors.As(err, &e) {
		if e.StatusCode == 404 {
			return true
		}
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

// SyncImage syncs specific image between two registries for current architecture.
func SyncImage(ctx context.Context, src, dest string, dryRun bool, auth authn.Authenticator) (name.Reference, error) {
	log.Debug("source ", src)
	log.Debug("destination ", dest)

	sr, err := name.ParseReference(src)
	if err != nil {
		return nil, err
	}
	dr, err := name.ParseReference(dest)
	if err != nil {
		return nil, err
	}

	s, err := remote.Image(sr, remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("source image pull error: %w", err)
	}

	// lookup on target registry if image exists
	// HEAD requests usually do not count in rate limits
	if _, err := remote.Head(dr, remote.WithContext(ctx)); err != nil {
		if ifRefNotFound(err) {
			log.Debug("Target image does not exist. Pushing image...")
			if !dryRun {
				err := remote.Write(dr, s, remote.WithContext(ctx), remote.WithAuth(auth))
				if err != nil {
					return nil, fmt.Errorf("push image error: %w", err)
				}
				return dr, nil
			}
			log.Debug("Dry-Run enabled. Skipping push.")
		}
		return nil, err
	}

	d, err := remote.Image(dr, remote.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	ds, err := s.Digest()
	if err != nil {
		return nil, err
	}
	dd, err := d.Digest()
	if err != nil {
		return nil, err
	}
	log.Debug("src digest: ", ds)
	log.Debug("dest digest: ", dd)
	if ds != dd {
		return nil, fmt.Errorf("digests are not equal: %v, %v - probably source image has been changed", ds, dd)
	}
	log.Debug("Digests are equal. Nothing to do.")
	return dr, nil
}

// SyncImages is a main syncing function that takes care of copying and signing/verifying images.
func SyncImages(ctx context.Context, cfg *Config, images *SyncDef, sv signature.SignerVerifier, authCfg []byte) error {
	auth := &authn.Basic{Username: "_json_key", Password: string(authCfg)}
	for _, img := range images.Images {
		target, err := getTarget(img.Source, images.TargetRepoPrefix, img.Tag)
		if err != nil {
			return err
		}
		log.WithField("image", img.Source).Info("Start scan on image")
		targetImg, err := SyncImage(ctx, img.Source, target, cfg.DryRun, auth)
		if err != nil {
			return err
		}
		if shouldSign(images.Sign, img.Sign) {
			log.Debug("Verifying image signature")
			err = Verify(ctx, sv, targetImg, auth)
			if err != nil {
				if ifRefNotFound(err) {
					// no signature found
					log.Debug("Signature not found. Signing the image")
					err := Sign(ctx, sv, targetImg, cfg.DryRun, auth)
					if err != nil {
						return fmt.Errorf("image sign error: %w", err)
					}
				} else {
					return fmt.Errorf("image verify error: %w", err)
				}
			}
		}
		log.WithField("target", target).Info("Image synced successfully")
	}
	return nil
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

			imagesFile, err := parseImagesFile(cfg.ImagesFile)
			if err != nil {
				log.WithError(err).Fatal("Could not parse images file")
			}
			authCfg, err := ioutil.ReadFile(cfg.TargetKeyFile)
			if err != nil {
				log.WithError(err).Fatal("Could not open target auth key JSON")
			}

			if cfg.DryRun {
				log.Info("Dry-Run enabled. Program will not make any changes to the target repository.")
			}

			sv, err := NewKMSSignerVerifier(ctx, cfg.KeyRef)
			if err != nil {
				log.WithError(err).Fatal("Failed to create signer instance")
			}
			if err := SyncImages(ctx, &cfg, imagesFile, sv, authCfg); err != nil {
				log.WithError(err).Fatal("Failed to sync images")
			} else {
				log.Info("All images synced successfully")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.ImagesFile, "images-file", "i", "", "yaml file containing list of images")
	rootCmd.PersistentFlags().StringVarP(&cfg.TargetKeyFile, "target-repo-auth-key", "t", "", "JSON key file used for authorization to target repo")
	rootCmd.PersistentFlags().StringVarP(&cfg.KeyRef, "kms-key", "k", "", "path to KMS key resource (eg. gcpkms://...)")
	rootCmd.PersistentFlags().BoolVarP(&cfg.DryRun, "dry-run", "d", true, "dry run mode")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "enable debug mode")

	rootCmd.MarkPersistentFlagRequired("images-file")
	rootCmd.MarkPersistentFlagRequired("target-repo-auth-key")
	rootCmd.MarkPersistentFlagRequired("kms-key")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "SYNCER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
