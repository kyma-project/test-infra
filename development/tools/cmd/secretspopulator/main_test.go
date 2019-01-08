package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/kyma-project/test-infra/development/tools/cmd/secretspopulator/automock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/cloudkms/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8s_testing "k8s.io/client-go/testing"
)

func TestSecretsFromData(t *testing.T) {
	// GIVEN
	givenData := RequiredSecretsData{
		Generics: []GenericSecret{
			{
				Key:    "givenKey",
				Prefix: "givenPrefix",
			},
		},
		ServiceAccounts: []SASecret{
			{
				Prefix: "givenSAPrefix",
			},
		},
	}
	// WHEN
	actual := SecretsFromData(givenData)
	// THEN
	assert.Len(t, actual, 2)
	assert.Contains(t, actual, SecretModel{Key: "givenKey", Prefix: "givenPrefix"})
	assert.Contains(t, actual, SecretModel{Key: "service-account.json", Prefix: "givenSAPrefix"})
}

const (
	existingSecretName = "existing"
	newSecretName      = "new"
)

func TestPopulateSecrets(t *testing.T) {
	// GIVEN
	mockDecryptor := &automock.Decryptor{}
	defer mockDecryptor.AssertExpectations(t)
	mockStorageReader := &automock.StorageReader{}
	defer mockStorageReader.AssertExpectations(t)
	fakeClientset := fake.NewSimpleClientset(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      existingSecretName,
			Namespace: metav1.NamespaceDefault,
		},
		Data: make(map[string][]byte),
	})

	fakeClientset.PrependReactor("get", "secrets", func(action k8s_testing.Action) (handled bool, ret runtime.Object, err error) {
		impl := action.(k8s_testing.GetActionImpl)
		if impl.Name == newSecretName {
			return true, nil, errors.NewNotFound(schema.GroupResource{}, "")
		}
		return false, nil, nil

	})

	mockStorageReader.On("Read", mock.Anything, "bucket", "existing.encrypted").Return(bytes.NewBufferString("existingEncrypted"), nil)
	mockStorageReader.On("Read", mock.Anything, "bucket", "new.encrypted").Return(bytes.NewBufferString("newEncrypted"), nil)

	parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		"project", "location", "keyring", "key")
	mockDecryptor.On("Decrypt", parentName, []byte("existingEncrypted")).Return(&cloudkms.DecryptResponse{Plaintext: base64.StdEncoding.EncodeToString([]byte("existingPlain"))}, nil)
	mockDecryptor.On("Decrypt", parentName, []byte("newEncrypted")).Return(&cloudkms.DecryptResponse{Plaintext: base64.StdEncoding.EncodeToString([]byte("newPlain"))}, nil)

	sut := SecretsPopulator{
		decryptor:     mockDecryptor,
		storageReader: mockStorageReader,
		secretsClient: fakeClientset.CoreV1().Secrets(metav1.NamespaceDefault),
		logger:        logrus.StandardLogger(),
	}

	// WHEN
	err := sut.PopulateSecrets(context.Background(), "project",
		[]SecretModel{{Prefix: newSecretName, Key: "token"}, {Prefix: existingSecretName, Key: "service-account.json"}},
		"bucket",
		"keyring",
		"key",
		"location")

	// THEN
	require.NoError(t, err)

	secretsList, err := fakeClientset.CoreV1().Secrets(metav1.NamespaceDefault).List(metav1.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, secretsList.Items, 2)
	assert.Equal(t, existingSecretName, secretsList.Items[0].Name)
	assert.Equal(t, []byte("existingPlain"), secretsList.Items[0].Data["service-account.json"])
	assert.Equal(t, newSecretName, secretsList.Items[1].Name)
	assert.Equal(t, []byte("newPlain"), secretsList.Items[1].Data["token"])

}
