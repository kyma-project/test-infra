package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyma-project/test-infra/pkg/imagesync"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/pkg/errors"

	"github.com/jamiealquiza/envy"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log = logrus.New()
)

// Config stores command line arguments
type Config struct {
	ImagesFile    string
	TargetKeyFile string
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

func isImageIndex(ctx context.Context, src string) (bool, error) {

	sr, err := name.ParseReference(src)
	if err != nil {
		return false, err
	}

	s, err := remote.Head(sr, remote.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("source image index pull error: %w", err)
	}
	return s.MediaType.IsIndex(), nil
}

// SyncIndex syncs specific image index between two registries.
func SyncIndex(ctx context.Context, src, dest string, dryRun bool, auth authn.Authenticator) (name.Reference, error) {
	log.Debug("Source ", src)
	log.Debug("Destination ", dest)

	sr, err := name.ParseReference(src)
	if err != nil {
		return nil, err
	}
	dr, err := name.ParseReference(dest)
	if err != nil {
		return nil, err
	}

	s, err := remote.Index(sr, remote.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("source index index pull error: %w", err)
	}

	d, err := remote.Index(dr, remote.WithContext(ctx), remote.WithAuth(auth))
	if err != nil {
		if ifRefNotFound(err) {
			log.Debug("Target index does not exist. Pushing index...")
			if !dryRun {
				err := remote.WriteIndex(dr, s, remote.WithContext(ctx), remote.WithAuth(auth))
				if err != nil {
					return nil, fmt.Errorf("push index error: %w", err)
				}
			} else {
				log.Debug("Dry-Run enabled. Skipping push.")
			}
			return dr, nil
		}
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
	log.Debug("Src digest: ", ds)
	log.Debug("Dest digest: ", dd)
	if ds != dd {
		return nil, fmt.Errorf("digests are not equal: %v, %v - probably source index has been changed", ds, dd)
	}
	log.Debug("Digests are equal - nothing to sync")
	return dr, nil
}

// SyncImage syncs specific image between two registries for amd64/linux architecture.
func SyncImage(ctx context.Context, src, dest string, dryRun bool, auth authn.Authenticator) (name.Reference, error) {
	log.Debug("Source ", src)
	log.Debug("Destination ", dest)

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

	d, err := remote.Image(dr, remote.WithContext(ctx))
	if err != nil {
		if ifRefNotFound(err) {
			log.Debug("Target image does not exist. Pushing image...")
			if !dryRun {
				err := remote.Write(dr, s, remote.WithContext(ctx), remote.WithAuth(auth))
				if err != nil {
					return nil, fmt.Errorf("push image error: %w", err)
				}
			} else {
				log.Debug("Dry-Run enabled. Skipping push.")
			}
			return dr, nil
		}
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
	log.Debug("Src digest: ", ds)
	log.Debug("Dest digest: ", dd)
	if ds != dd {
		return nil, fmt.Errorf("digests are not equal: %v, %v - probably source image has been changed", ds, dd)
	}
	log.Debug("Digests are equal - nothing to sync")
	return dr, nil
}

// SyncImages is a main syncing function that takes care of copying images.
func SyncImages(ctx context.Context, cfg *Config, images *imagesync.SyncDef, authCfg []byte) error {
	auth := &authn.Basic{Username: "_json_key", Password: string(authCfg)}
	for _, img := range images.Images {
		// Update target repo prefix and source image to reflect namespace handling
		target, err := getTarget(img.Source, images.TargetRepoPrefix, img.Tag)
		imageType := "Index"
		if err != nil {
			return err
		}
		log.WithField("image", img.Source).Info("Start sync")
		if img.AMD64Only {
			// sync whole index if possible, otherwise sync singular image
			// we force users to set explicit info about single-arch images
			var isIndex bool
			isIndex, err = isImageIndex(ctx, img.Source)
			if err != nil {
				return err
			}
			if isIndex {
				_, err = SyncIndex(ctx, img.Source, target, cfg.DryRun, auth)
			} else {
				imageType = "Image"
				_, err = SyncImage(ctx, img.Source, target, cfg.DryRun, auth)
			}
		} else {
			// sync whole index
			_, err = SyncIndex(ctx, img.Source, target, cfg.DryRun, auth)
		}
		if err != nil {
			return err
		}
		log.WithField("target", target).Infof("%s synced successfully", imageType)
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
		//nolint:revive
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
			authCfg, err := os.ReadFile(cfg.TargetKeyFile)
			if err != nil {
				log.WithError(err).Fatal("Could not open target auth key JSON")
			}

			if cfg.DryRun {
				log.Info("Dry-Run enabled. Program will not make any changes to the target repository.")
			}

			if err != nil {
				log.WithError(err).Fatal("Failed to create signer instance")
			}
			if err := SyncImages(ctx, &cfg, imagesFile, authCfg); err != nil {
				log.WithError(err).Fatal("Failed to sync images")
			} else {
				log.Info("All images synced successfully")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.ImagesFile, "images-file", "i", "", "Specifies the path to the YAML file that contains list of images")
	rootCmd.PersistentFlags().StringVarP(&cfg.TargetKeyFile, "target-repo-auth-key", "t", "", "Specifies the JSON key file used for authorization to the target repository")
	rootCmd.PersistentFlags().BoolVar(&cfg.DryRun, "dry-run", false, "Enables the dry-run mode")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Enables the debug mode")

	rootCmd.MarkPersistentFlagRequired("images-file")
	rootCmd.MarkPersistentFlagRequired("target-repo-auth-key")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "SYNCER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
