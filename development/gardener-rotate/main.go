package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	secretmanager "google.golang.org/api/secretmanager/v1"
	authentication "k8s.io/api/authentication/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

var (
	log = logrus.New()
)

type ServiceAccount struct {
	KubernetesSA        string `json:"serviceAccount"`
	KubernetesNamespace string `json:"namespace,omitempty"`
	GCPSecret           string `json:"secret"`
	GCPProject          string `json:"project"`
	KeepOld             bool   `json:"keepOld,omitempty"`
	Duration            int64  `json:"duration"`
}

type ConfigFile struct {
	ServiceAccounts []ServiceAccount `json:"serviceAccounts"`
}

// Config stores command line arguments
type Config struct {
	ServiceAccount string
	Kubeconfig     string
	ConfigFile     string
	Cluster        string
	DryRun         bool
}

func main() {
	log.Out = os.Stdout
	var cfg Config

	var rootCmd = &cobra.Command{
		Use:   "gardener-rotate",
		Short: "gardener-rotate rotates kubeconfig saved in Secret Manager",
		Long:  `gardener-rotate creates new gardener service account token and saves updated kubeconfig in Secret Manager`,
		Run: func(cmd *cobra.Command, args []string) {
			log.SetLevel(logrus.InfoLevel)
			ctx := context.Background()

			// Prepare Secret Manager API and gardener Kubernetes clients
			var serviceAccountGCP string
			if cfg.ServiceAccount != "" {
				serviceAccountGCP = cfg.ServiceAccount
			} else {
				serviceAccountGCP = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			}

			secretSvc, err := secretmanager.NewService(ctx, option.WithCredentialsFile(serviceAccountGCP))
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
				log.Infof("Rotating token for %s service accout", sa.KubernetesSA)

				// generate new token with duration
				namespace := "default"
				if sa.KubernetesNamespace != "" {
					namespace = sa.KubernetesNamespace
				}

				log.Infof("Generating new token for %s service accout with %ds duration", sa.KubernetesSA, sa.Duration)
				tokenRequest := authentication.TokenRequest{Spec: authentication.TokenRequestSpec{ExpirationSeconds: &sa.Duration}}
				var tokenRequestResponse *authentication.TokenRequest
				if !cfg.DryRun {
					tokenRequestResponse, err = k8sClient.CoreV1().ServiceAccounts(namespace).CreateToken(ctx, sa.KubernetesSA, &tokenRequest, meta.CreateOptions{})
					if err != nil {
						log.Fatalf("Could not get create new token request: %v", err)
					}
				}

				// create new kubeconfig
				log.Infof("Generating new kubeconfig for %s service accout", sa.KubernetesSA)
				serviceAccountKubeconfig := ""
				if !cfg.DryRun {
					serviceAccountKubeconfig, err = generateKubeconfig(k8sConfig.Host, cfg.Cluster, sa, tokenRequestResponse.Status.Token)
					if err != nil {
						log.Fatalf("Could not get generate kubeconfig: %v", err)
					}
				}

				secretParent := "projects/" + sa.GCPProject + "/secrets/" + sa.GCPSecret

				// get list of all previous secret versions
				secretVersionsCall := secretSvc.Projects.Secrets.Versions.List(secretParent)
				var secretVersions *secretmanager.ListSecretVersionsResponse
				if !cfg.DryRun {
					secretVersions, err = secretVersionsCall.Do()
					if err != nil {
						log.Fatalf("Could not get list of secret versions: %v", err)
					}
				}

				// update it in the Secret Manager
				log.Infof("Adding new secret version for %s service accout", sa.KubernetesSA)
				newVersionRequest := secretmanager.AddSecretVersionRequest{Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString([]byte(serviceAccountKubeconfig))}}
				newVersionCall := secretSvc.Projects.Secrets.AddVersion(secretParent, &newVersionRequest)
				if !cfg.DryRun && !sa.KeepOld {
					_, err = newVersionCall.Do()
					if err != nil {
						log.Fatalf("Could not create new secret version: %v", err)
					}
				}

				// disable all previous versions

				if !sa.KeepOld {
					log.Infof("Disabling old secret versions for %s service accout", sa.KubernetesSA)
					if !cfg.DryRun {
						for _, secretVersion := range secretVersions.Versions {
							// we can only disable enabled secrets
							if secretVersion.State == "ENABLED" {
								disableRequest := secretmanager.DisableSecretVersionRequest{}
								disableCall := secretVersionsSvc.Disable(secretVersion.Name, &disableRequest)
								_, err := disableCall.Do()
								if err != nil {
									log.Fatalf("Could not disable secret version %s: %v", secretVersion.Name, err)
								}
							}
						}
					}
				}
			}

		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfg.ServiceAccount, "service-account", "s", "", "Path to GCP service account credentials file")
	rootCmd.PersistentFlags().StringVarP(&cfg.Kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&cfg.ConfigFile, "config-file", "c", "", "Specifies the path to the YAML configuration file")
	rootCmd.PersistentFlags().BoolVar(&cfg.DryRun, "dry-run", true, "Enables the dry-run mode")
	rootCmd.PersistentFlags().StringVarP(&cfg.Cluster, "cluster-name", "n", "gardener", "Specifies the name of the cluster")

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
