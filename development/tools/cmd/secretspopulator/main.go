package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	var (
		argBucket         = flag.String("bucket", "", "")
		argKeyring        = flag.String("keyring", "", "")
		argKey            = flag.String("key", "", "")
		argLocation       = flag.String("location", "", "")
		argKubeconfigPath = flag.String("kubeconfig", "", "")
		argProject        = flag.String("project", "", "")
		argSecretsDefFile = flag.String("secrets-def-file", "", "")
	)
	flag.Parse()

	err := checkRequiredFlags(flag.CommandLine, "bucket", "keyring", "key", "location", "kubeconfig", "project", "secrets-def-file")
	fatalOnError(err)

	logger := logrus.StandardLogger()

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", *argKubeconfigPath)
	fatalOnError(err)
	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err)

	secretInterface := k8sCli.CoreV1().Secrets(metav1.NamespaceDefault)

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	fatalOnError(err)
	client, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	fatalOnError(err)

	kmsService, err := cloudkms.New(client)
	fatalOnError(err)

	p := SecretsPopulator{
		secretsClient: secretInterface,
		decryptor:     &decryptorWrapper{wrapped: kmsService},
		storageReader: &storageReaderWrapper{wrapped: storageClient},
		logger:        logger,
	}

	secrets, err := readSecretDefFile(*argSecretsDefFile)
	fatalOnError(err)

	fatalOnError(p.PopulateSecrets(ctx, *argProject, secrets, *argBucket, *argKeyring, *argKey, *argLocation))

}

// SecretsPopulator is responsible for populating secrets
type SecretsPopulator struct {
	secretsClient typedv1.SecretInterface
	storageReader StorageReader
	decryptor     Decryptor
	logger        logrus.FieldLogger
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=Decryptor -output=automock -outpkg=automock -case=underscore

// Decryptor decrypts data
type Decryptor interface {
	Decrypt(key string, bytes []byte) (*cloudkms.DecryptResponse, error)
}

type decryptorWrapper struct {
	wrapped *cloudkms.Service
}

func (w *decryptorWrapper) Decrypt(decryptKey string, bytes []byte) (*cloudkms.DecryptResponse, error) {
	decryptCall := w.wrapped.Projects.Locations.KeyRings.CryptoKeys.Decrypt(decryptKey, &cloudkms.DecryptRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(bytes),
	})
	return decryptCall.Do()
}

//go:generate go run github.com/vektra/mockery/cmd/mockery -name=StorageReader -output=automock -outpkg=automock -case=underscore

// StorageReader provide interface for reading from object storage
type StorageReader interface {
	Read(ctx context.Context, bucket, name string) (io.Reader, error)
}

type storageReaderWrapper struct {
	wrapped *storage.Client
}

func (w *storageReaderWrapper) Read(ctx context.Context, bucket, name string) (io.Reader, error) {
	objHandle := w.wrapped.Bucket(bucket).Object(name)
	return objHandle.NewReader(ctx)
}

// PopulateSecrets populates secrets
func (s *SecretsPopulator) PopulateSecrets(ctx context.Context, project string, secrets []SecretModel, bucket, keyring, key, location string) error {
	parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		project, location, keyring, key)

	for _, sec := range secrets {
		name := fmt.Sprintf("%s.%s", sec.Name, fileNameExtension)
		s.logger.Infof("Reading object [%s] from bucket [%s]", name, bucket)
		r, err := s.storageReader.Read(ctx, bucket, name)
		if err != nil {
			return errors.Wrapf(err, "while reading object [%s] from bucket [%s]", name, bucket)
		}
		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, r); err != nil {
			return errors.Wrapf(err, "while coping file [%s] to buffer", name)
		}

		resp, err := s.decryptor.Decrypt(parentName, buf.Bytes())
		if err != nil {
			return errors.Wrap(err, "while making decrypt call")
		}

		decoded, err := base64.StdEncoding.DecodeString(string(resp.Plaintext))
		if err != nil {
			return err
		}

		curr, err := s.secretsClient.Get(sec.Name, metav1.GetOptions{})
		switch {
		case err == nil:
			if bytes.Equal(curr.Data[sec.Key], decoded) {
				s.logSecretAction(curr, "Unchanged")
				continue
			}

			curr.Data[sec.Key] = decoded
			if _, err = s.secretsClient.Update(curr); err != nil {
				return errors.Wrap(err, "while updating secret")
			}
			s.logSecretAction(curr, "Updated")

		case k8serrors.IsNotFound(err):
			curr, err := s.secretsClient.Create(&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: sec.Name,
				},
				Data: map[string][]byte{
					sec.Key: decoded,
				},
			})
			if err != nil {
				return errors.Wrapf(err, "while creating secret [%s]", sec.Name)
			}
			s.logSecretAction(curr, "Created")
		default:
			return errors.Wrapf(err, "while getting secret [%s]", sec.Name)
		}

	}
	return nil
}

// logSecretAction receives a Secret, formats it and logs.
func (s *SecretsPopulator) logSecretAction(secret *corev1.Secret, action string) {
	var keys []string
	for k := range secret.Data {
		keys = append(keys, k)
	}
	s.logger.Infof("%s Secret [%s] with key(s) [%s]", action, secret.Name, strings.Join(keys, ", "))
}

// RequiredSecretsData represents secrets required by Prow cluster
type RequiredSecretsData struct {
	ServiceAccounts []SASecret
	Generics        []GenericSecret
}

// SASecret represents Service Account Secret
type SASecret struct {
	Prefix string
}

// GenericSecret represents other than Service Account secrets
type GenericSecret struct {
	Prefix string
	Key    string
}

// SecretModel defines secret to be stored in k8s
type SecretModel struct {
	Name string
	Key  string
}

// SecretsFromData converts input data to SecretModels
func SecretsFromData(data RequiredSecretsData) []SecretModel {
	out := make([]SecretModel, 0)
	for _, sa := range data.ServiceAccounts {
		out = append(out, SecretModel{Name: sa.Prefix, Key: "service-account.json"})
	}
	for _, g := range data.Generics {
		out = append(out, SecretModel{Name: g.Prefix, Key: g.Key})
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

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err)
	}
}

func checkRequiredFlags(flags *flag.FlagSet, requiredFlags ...string) error {
	required := map[string]struct{}{}
	for _, name := range requiredFlags {
		required[name] = struct{}{}
	}

	var unSetFlags []string

	flags.VisitAll(func(flag *flag.Flag) {
		_, flagRequired := required[flag.Name]

		if flagRequired && flag.Value.String() == "" {
			unSetFlags = append(unSetFlags, flag.Name)
		}
	})

	if len(unSetFlags) > 0 {
		return fmt.Errorf(`missing required flag(s): "%s"`, strings.Join(unSetFlags, `", "`))
	}

	return nil
}
