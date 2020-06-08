package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jamiealquiza/envy"
	parser "github.com/novln/docker-parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	log = logrus.New()
)

type SyncDef struct {
	TargetRepoPrefix string `yaml:"targetRepoPrefix"`
	Images           []Image
}

type Image struct {
	Source string
}

type Config struct {
	ImagesFile    string
	TargetKeyFile string
	DryRun        bool
}

func getAuthString(user, password string) (string, error) {
	authConfig := types.AuthConfig{
		Username: user,
		Password: password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encodedJSON), nil
}

func getImageIdAndRepoDigest(ctx context.Context, cli *client.Client, image string) (string, string, error) {
	if reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{}); err != nil {
		return "", "", fmt.Errorf("image pull failed: %w", err)
	} else {
		w := log.WriterLevel(logrus.DebugLevel)
		defer w.Close()
		io.Copy(w, reader)
	}

	details, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return "", "", fmt.Errorf("image inspection failed: %w", err)
	}

	ref, err := parser.Parse(image)
	imageRepository := ref.Repository()

	for _, url := range details.RepoDigests {
		ref, err := parser.Parse(url)
		if err != nil {
			return "", "", fmt.Errorf("url parsing failed: %w", err)
		}
		if imageRepository == ref.Repository() {
			return details.ID, ref.Tag(), nil
		}
	}

	return "", "", errors.New(fmt.Sprintf("Unable to find digest for '%s'", image))
}

func safeCopyImage(ctx context.Context, cli *client.Client, authString, source, target string, dryRun bool) error {
	if source == "" {
		return errors.New("source image can not be empty")
	}
	log.Infof("Source image: %s", source)
	sourceId, sourceDigest, err := getImageIdAndRepoDigest(ctx, cli, source)
	if err != nil {
		return err
	}
	log.Infof("Source ID: %s", sourceId)
	log.Infof("Source repo digest: %s", sourceDigest)

	log.Infof("Target image: %s", target)
	targetId, targetDigest, err := getImageIdAndRepoDigest(ctx, cli, target)
	if isImageNotFoundError(err) {
		log.Info("Target image does not exist")
		if dryRun {
			log.Info("Dry-run mode - tagging and pushing skipped")
			return nil
		}
		if err = cli.ImageTag(ctx, source, target); err != nil {
			return err
		}
		log.Info("Image re-tagged")
		log.Info("Pushing to target repo")
		if reader, err := cli.ImagePush(ctx, target, types.ImagePushOptions{RegistryAuth: authString}); err != nil {
			return err
		} else {
			w := log.WriterLevel(logrus.DebugLevel)
			defer w.Close()
			io.Copy(w, reader)
		}
		log.Info("Image pushed successfully")
	} else if err != nil {
		return err
	} else {
		log.Infof("Target ID: %s", targetId)
		log.Infof("Target repo digest: %s", targetDigest)
		if sourceId != targetId {
			return errors.New("source and target IDs are different - probably source image has been changed")
		}
		log.Info("Source and target IDs are the same, nothing to do")
	}

	return nil
}

func isImageNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if strings.Index(err.Error(), "not found: manifest unknown") != -1 {
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

func copyImages(cfg Config) error {

	log.SetLevel(logrus.InfoLevel)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	cancelOnInterrupt(ctx, cancelFunc)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	imagesFile, err := ioutil.ReadFile(cfg.ImagesFile)
	if err != nil {
		return err
	}
	var syncDef SyncDef
	if err := yaml.Unmarshal(imagesFile, &syncDef); err != nil {
		return err
	}
	if syncDef.TargetRepoPrefix == "" {
		return errors.New("TargetRepoPrefix can not be empty")
	}

	jsonKey, err := ioutil.ReadFile(cfg.TargetKeyFile)
	if err != nil {
		return err
	}

	authString, err := getAuthString("_json_key", string(jsonKey))
	if err != nil {
		return err
	}

	for _, image := range syncDef.Images {
		err = safeCopyImage(ctx, cli, authString, image.Source, syncDef.TargetRepoPrefix+image.Source, cfg.DryRun)
		if err != nil {
			return err
		}
		log.Info("------------------------")
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
			if err := copyImages(cfg); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.ImagesFile, "images-file", "i", "", "yaml file containing list of images")
	rootCmd.PersistentFlags().StringVarP(&cfg.TargetKeyFile, "target-key-file", "t", "", "JSON key file used for authorization to target repo")
	rootCmd.PersistentFlags().BoolVarP(&cfg.DryRun, "dry-run", "d", true, "dry run mode")

	rootCmd.MarkPersistentFlagRequired("images-file")
	rootCmd.MarkPersistentFlagRequired("target-key-file")
	envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "SYNCER", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
