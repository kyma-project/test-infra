package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	secretmanager "google.golang.org/api/secretmanager/v1"
	authentication "k8s.io/api/authentication/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

// "github.com/kyma-project/test-infra/development/gardener-rotate/pkg"

var (
	log                = logrus.New()
	kubeconfigTemplate = `apiVersion: v1
    kind: Config
    current-context: garden-neighbors-sa-neighbor-robot
    contexts:
      - name: garden-neighbors-sa-neighbor-robot
        context:
          cluster: garden
          user: sa-neighbor-robot
          namespace: garden-neighbors
    clusters:
      - name: garden
        cluster:
          server: https://api.canary.gardener.cloud.sap
    users:
      - name: sa-neighbor-robot
        user:
          token: >-
            `
)

type ServiceAccount struct {
	KubernetesSA        string `json:"serviceAccount"`
	KubernetesNamespace string `json:"namespace,omitempty"`
	GCPSecret           string `json:"secret"`
	GCPProject          string `json:"project"`
	KeepOld             bool   `json:"keepOld,omitEmpty"`
	Duration            int64  `json:"duration"`
}

type ConfigFile struct {
	ServiceAccounts []ServiceAccount `json:"serviceAccounts"`
}

// Config stores command line arguments
type Config struct {
	Kubeconfig string
	ConfigFile string
	DryRun     bool
	Debug      bool
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
			ctx := context.Background()

			// Prepare Secret Manager API and gardener Kubernetes clients
			connection, err := google.DefaultClient(ctx, container.CloudPlatformScope)
			if err != nil {
				log.Fatalf("Could not get authenticated client: %v", err)
			}

			secretSvc, err := secretmanager.New(connection)
			if err != nil {
				log.Fatalf("Could not initialize Secret Manager API client: %v", err)
			}
			secretVersionsSvc := secretmanager.NewProjectsSecretsVersionsService(secretSvc)

			k8sConfig, err := clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
			if err != nil {
				log.Fatalf("Could not read Kubeconfig: %v", err)
			}

			k8sClient, err := kubernetes.NewForConfig(k8sConfig)
			if err != nil {
				log.Fatalf("Could not initialize Kubernetes API client: %v", err)
			}

			// parse config file
			yamlFile, err := ioutil.ReadFile(cfg.ConfigFile)
			if err != nil {
				log.Fatalf("error while opening %s file: %s", cfg.ConfigFile, err)
			}

			var parsedConfig ConfigFile
			err = yaml.UnmarshalStrict(yamlFile, &parsedConfig)
			if err != nil {
				log.Fatalf("error while unmarshalling %s file: %s", cfg.ConfigFile, err)
			}

			// for each service account
			for _, sa := range parsedConfig.ServiceAccounts {
				log.Debugf("Rotating token for %s service accout", sa.KubernetesSA)

				if !cfg.DryRun {
					// generate new token with duration
					namespace := "default"
					if sa.KubernetesNamespace != "" {
						namespace = sa.KubernetesNamespace
					}

					tokenRequest := authentication.TokenRequest{Spec: authentication.TokenRequestSpec{ExpirationSeconds: &sa.Duration}}

					tokenRequestResponse, err := k8sClient.CoreV1().ServiceAccounts(namespace).CreateToken(ctx, sa.KubernetesSA, &tokenRequest, meta.CreateOptions{})
					if err != nil {
						log.Fatalf("Could not get create new token request: %v", err)
					}

					// create new kubeconfig
					serviceAccountKubeconfig, err := generateKubeconfig(k8sConfig.Host, "garden", sa, tokenRequestResponse.Status.Token) //kubeconfigTemplate + tokenRequestResponse.Status.Token + "\n"
					if err != nil {
						log.Fatalf("Could not get generate kubeconfig: %v", err)
					}

					secretParent := "projects/" + sa.GCPProject + "/secrets/" + sa.GCPSecret

					// get list of all previous secret versions
					secretVersionsCall := secretSvc.Projects.Secrets.Versions.List(secretParent)
					secretVersions, err := secretVersionsCall.Do()
					if err != nil {
						log.Fatalf("Could not get list of secret versions: %v", err)
					}

					// update it in the Secret Manager
					newVersionRequest := secretmanager.AddSecretVersionRequest{Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString([]byte(serviceAccountKubeconfig))}}
					newVersionCall := secretSvc.Projects.Secrets.AddVersion(secretParent, &newVersionRequest)

					_, err = newVersionCall.Do()
					if err != nil {
						log.Fatalf("Could not create new secret version: %v", err)
					}

					// disable all previous versions
					if !sa.KeepOld {
						for _, secretVersion := range secretVersions.Versions {
							// we can only disable enabled secrets
							if secretVersion.State == "ENABLED" {
								disableRequest := secretmanager.DisableSecretVersionRequest{}
								disableCall := secretVersionsSvc.Disable(secretVersion.Name, &disableRequest)
								_, err := disableCall.Do()
								if err != nil {
									log.Fatalf("Could not disable secret version %d: %v", secretVersion.Name, err)
								}
							}
						}
					}
				}

				// TODO destroy versions older than x time ?
			}

		},
	}

	// rootCmd.PersistentFlags().StringVarP(&cfg.ServiceAccount, "service-account", "c", "", "Path to GCP service account credentials file")
	rootCmd.PersistentFlags().StringVarP(&cfg.Kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&cfg.ConfigFile, "config-file", "c", "", "Specifies the path to the YAML configuration file")
	rootCmd.PersistentFlags().BoolVar(&cfg.DryRun, "dry-run", false, "Enables the dry-run mode")
	rootCmd.PersistentFlags().BoolVar(&cfg.Debug, "debug", false, "Enables the debug mode")

	rootCmd.MarkPersistentFlagRequired("config-file")
	rootCmd.MarkPersistentFlagRequired("kubeconfig")
	// envy.ParseCobra(rootCmd, envy.CobraConfig{Prefix: "ROTATE", Persistent: true, Recursive: false})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func generateKubeconfig(server string, clusterName string, sa ServiceAccount, token string) (string, error) {
	config := api.NewConfig()
	config.CurrentContext = sa.KubernetesNamespace + "-" + sa.KubernetesSA

	context := api.NewContext()
	context.Cluster = clusterName
	context.Namespace = sa.KubernetesNamespace
	context.AuthInfo = sa.KubernetesSA
	config.Contexts[sa.KubernetesNamespace+"-"+sa.KubernetesSA] = context

	cluster := api.NewCluster()
	cluster.Server = server
	config.Clusters[clusterName] = cluster

	user := api.NewAuthInfo()
	user.Token = token
	config.AuthInfos[sa.KubernetesSA] = user

	marshalled, err := clientcmd.Write(*config)

	return string(marshalled), err
}
