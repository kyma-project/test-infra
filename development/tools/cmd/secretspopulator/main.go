package main

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
	"io"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const fileNameExtension = "encrypted"
const defaultSecretDataKey = "service-account.json"
const overrideSecretDataKeyMetadata = "override-secret-data-key"

func main() {
	argBucket := flag.String("bucket", "", "")
	argKeyring := flag.String("keyring", "", "")
	argKey := flag.String("key", "", "")
	argLocation := flag.String("location", "", "")
	argKubeconfigPath := flag.String("kubeconfig", "", "")
	flag.Parse()
	panicOnMissingArg("bucket", argBucket)
	panicOnMissingArg("keyring", argKeyring)
	panicOnMissingArg("key", argKey)
	panicOnMissingArg("location", argLocation)
	panicOnMissingArg("kubeconfig", argKubeconfigPath)

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", *argKubeconfigPath)
	panicOnError(err)
	k8sCli, err := kubernetes.NewForConfig(k8sConfig)

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
	panicOnError(p.PopulateSecrets(ctx, []string{"sa-gke-kyma-integration", "sa-vm-kyma-integration", "sa-gcs-plank", "sa-gcr-push", "kyma-bot-npm-token"},
		*argBucket, *argKeyring, *argKey, *argLocation))

}

type SecretsPopulator struct {
	secretsClient typedv1.SecretInterface
	storageClient *storage.Client
	kmsClient     *cloudkms.Service
}

func (s *SecretsPopulator) PopulateSecrets(ctx context.Context, fileNamePrefixes []string, bucket, keyring, key, location string) error {
	parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		"kyma-project", location, keyring, key)

	for _, prefix := range fileNamePrefixes {
		name := fmt.Sprintf("%s.%s", prefix, fileNameExtension)
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
		attr, err := objHandle.Attrs(ctx)
		if err != nil {
			return errors.Wrapf(err, "while reading attributes for object [%s]", name)
		}
		overrideKey, ex := attr.Metadata[overrideSecretDataKeyMetadata]
		var dataKey string
		if ex {
			dataKey = overrideKey
		} else {
			dataKey = defaultSecretDataKey
		}

		curr, err := s.secretsClient.Get(prefix, metav1.GetOptions{})
		switch {
		case err == nil:
			fmt.Printf("Updating k8s secret [%s]\n", curr.Name)
			curr.Data[dataKey] = decoded
			if _, err = s.secretsClient.Update(curr); err != nil {
				return errors.Wrap(err, "while updating secret")
			}

		case k8serrors.IsNotFound(err):
			fmt.Printf("Creating k8s secret [%s]\n", prefix)
			if _, err := s.secretsClient.Create(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: prefix,
				},
				Data: map[string][]byte{
					dataKey: decoded,
				},
			}); err != nil {
				return errors.Wrapf(err, "while creating secret [%s]", prefix)
			}
		default:
			return errors.Wrapf(err, "while getting secret [%s]", prefix)
		}

	}
	return nil
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
