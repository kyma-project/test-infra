package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"io"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const fileNameExtension = "encrypted"

func main() {
	argBucket := flag.String("bucket", "", "")
	argKeyring := flag.String("keyring", "", "")
	argKey := flag.String("key", "", "")
	argLocation := flag.String("location", "", "")
	argKubeconfigPath := flag.String("kubeconfig", "", "")
	argProject := flag.String("project", "", "")
	argSecretsDefFile := flag.String("secrets-def-file", "", "")
	flag.Parse()
	panicOnMissingArg("bucket", argBucket)
	panicOnMissingArg("keyring", argKeyring)
	panicOnMissingArg("key", argKey)
	panicOnMissingArg("location", argLocation)
	panicOnMissingArg("kubeconfig", argKubeconfigPath)
	panicOnMissingArg("project", argProject)
	panicOnMissingArg("secrets-def-file", argSecretsDefFile)

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", *argKubeconfigPath)
	panicOnError(err)
	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	panicOnError(err)

	secretInterface := k8sCli.CoreV1().Secrets(metav1.NamespaceDefault)

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	panicOnError(err)
	client, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	panicOnError(err)

	kmsService, err := cloudkms.New(client)
	panicOnError(err)

	p := SecretsPopulator{
		secretsClient: secretInterface,
		storageClient: storageClient,
		kmsClient:     kmsService,
	}

	secrets, err := readSecretDefFile(*argSecretsDefFile)
	panicOnError(err)

	panicOnError(p.PopulateSecrets(ctx, *argProject, secrets, *argBucket, *argKeyring, *argKey, *argLocation))

}

// SecretsPopulator is responsible for populating secrets
type SecretsPopulator struct {
	secretsClient typedv1.SecretInterface
	storageClient *storage.Client
	kmsClient     *cloudkms.Service
}

// PopulateSecrets populates secrets
func (s *SecretsPopulator) PopulateSecrets(ctx context.Context, project string, secrets []SecretModel, bucket, keyring, key, location string) error {
	parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		project, location, keyring, key)

	for _, sec := range secrets {
		name := fmt.Sprintf("%s.%s", sec.Prefix, fileNameExtension)
		objHandle := s.storageClient.Bucket(bucket).Object(name)
		fmt.Printf("Reading object [%s] from bucket [%s]\n", name, bucket)
		r, err := objHandle.NewReader(ctx)
		if err != nil {
			return errors.Wrapf(err, "while reading object [%s] from bucket [%s]", name, bucket)
		}
		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, r); err != nil {
			return errors.Wrapf(err, "while coping file [%s] to buffer", name)
		}

		decryptCall := s.kmsClient.Projects.Locations.KeyRings.CryptoKeys.Decrypt(parentName, &cloudkms.DecryptRequest{
			Ciphertext: base64.StdEncoding.EncodeToString(buf.Bytes()),
		})

		resp, err := decryptCall.Do()
		if err != nil {
			return errors.Wrap(err, "while making decrypt call")
		}

		decoded, err := base64.StdEncoding.DecodeString(string(resp.Plaintext))
		if err != nil {
			return err
		}

		curr, err := s.secretsClient.Get(sec.Prefix, metav1.GetOptions{})
		switch {
		case err == nil:
			fmt.Printf("Updating k8s secret [%s]\n", curr.Name)
			curr.Data[sec.Key] = decoded
			if _, err = s.secretsClient.Update(curr); err != nil {
				return errors.Wrap(err, "while updating secret")
			}

		case k8serrors.IsNotFound(err):
			fmt.Printf("Creating k8s secret [%s] with key [%s] \n", sec.Prefix, sec.Key)
			if _, err := s.secretsClient.Create(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: sec.Prefix,
				},
				Data: map[string][]byte{
					sec.Key: decoded,
				},
			}); err != nil {
				return errors.Wrapf(err, "while creating secret [%s]", sec.Prefix)
			}
		default:
			return errors.Wrapf(err, "while getting secret [%s]", sec.Prefix)
		}

	}
	return nil
}

type RequiredSecretsData struct {
	ServiceAccounts []SASecret
	Generics        []GenericSecret
}

type SASecret struct {
	Prefix string
}

type GenericSecret struct {
	Prefix string
	Key    string
}

type SecretModel struct {
	Prefix string
	Key    string
}

func SecretsFromData(data RequiredSecretsData) []SecretModel {
	out := make([]SecretModel, 0)
	for _, sa := range data.ServiceAccounts {
		out = append(out, SecretModel{Prefix: sa.Prefix, Key: "service-account.json"})
	}
	for _, g := range data.Generics {
		out = append(out, SecretModel{Prefix: g.Prefix, Key: g.Key})
	}
	return out
}

func readSecretDefFile(fname string) ([]SecretModel, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "while opening secrets definition file [%s]", fname)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading bytes from secrets definition file [%s]", fname)
	}
	data := RequiredSecretsData{}
	if err := yaml.Unmarshal(b, &data); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling yaml file")
	}
	return SecretsFromData(data), nil
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
func panicOnMissingArg(argName string, val *string) {
	if val == nil || *val == "" {
		panic(fmt.Sprintf("missing argument [%s]", argName))
	}
}
