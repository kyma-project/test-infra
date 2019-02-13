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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	assert.Contains(t, actual, SecretModel{Key: "givenKey", Name: "givenPrefix"})
	assert.Contains(t, actual, SecretModel{Key: "service-account.json", Name: "givenSAPrefix"})
}

func TestPopulateSecretsSuccess(t *testing.T) {
	tests := map[string]struct {
		secretAlreadyInK8s v1.Secret

		decodedDataFromBucket string
		secretToProcess       SecretModel
		expWriteK8sActions    int
	}{
		"Should create not existing Secret": {
			secretAlreadyInK8s: v1.Secret{}, /* no matching secrets */
			secretToProcess: SecretModel{
				Name: "not-existing-secret",
				Key:  "key-no1",
			},
			decodedDataFromBucket: "decode-data",

			expWriteK8sActions: 1,
		},
		"Should update existing Secret when data is different": {
			secretAlreadyInK8s: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing-secret",
					Namespace: metav1.NamespaceDefault,
				},
				Data: map[string][]byte{
					"key-no1": []byte("old-data"),
				},
			},
			secretToProcess: SecretModel{
				Name: "existing-secret",
				Key:  "key-no1",
			},
			decodedDataFromBucket: "new-data",

			expWriteK8sActions: 1,
		},
		"Should not update existing Secret when data wasn't changed": {
			secretAlreadyInK8s: v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing-secret",
					Namespace: metav1.NamespaceDefault,
				},
				Data: map[string][]byte{
					"key-no1": []byte("same-data"),
				},
			},
			secretToProcess: SecretModel{
				Name: "existing-secret",
				Key:  "key-no1",
			},
			decodedDataFromBucket: "same-data",

			expWriteK8sActions: 0,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			mockDecryptor := &automock.Decryptor{}
			defer mockDecryptor.AssertExpectations(t)
			mockStorageReader := &automock.StorageReader{}
			defer mockStorageReader.AssertExpectations(t)

			fakeCli := fake.NewSimpleClientset(&tc.secretAlreadyInK8s)

			mockStorageReader.On("Read", mock.Anything, "bucket", tc.secretToProcess.Name+".encrypted").
				Return(bytes.NewBufferString("encrypted"), nil).
				Once()

			parentName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", "project", "location", "keyring", "key")
			mockDecryptor.On("Decrypt", parentName, []byte("encrypted")).
				Return(&cloudkms.DecryptResponse{Plaintext: base64.StdEncoding.EncodeToString([]byte(tc.decodedDataFromBucket))}, nil).
				Once()

			sut := SecretsPopulator{
				decryptor:     mockDecryptor,
				storageReader: mockStorageReader,
				secretsClient: fakeCli.CoreV1().Secrets(metav1.NamespaceDefault),
				logger:        logrus.StandardLogger(),
			}

			// WHEN
			err := sut.PopulateSecrets(context.Background(), "project",
				[]SecretModel{tc.secretToProcess},
				"bucket", "keyring",
				"key", "location")

			// THEN
			require.NoError(t, err)

			performedActions := filterOutReadonlyActions(fakeCli.Actions())
			assert.Len(t, performedActions, tc.expWriteK8sActions)

			secretsList, err := fakeCli.CoreV1().Secrets(metav1.NamespaceDefault).List(metav1.ListOptions{})
			require.NoError(t, err)
			assert.Len(t, secretsList.Items, 1)
			assert.Equal(t, tc.secretToProcess.Name, secretsList.Items[0].Name)
			assert.Equal(t, []byte(tc.decodedDataFromBucket), secretsList.Items[0].Data[tc.secretToProcess.Key])
		})
	}
}

func filterOutReadonlyActions(actions []k8s_testing.Action) []k8s_testing.Action {
	// based on API request verb, see:
	// https://kubernetes.io/docs/reference/access-authn-authz/authorization/#review-your-request-attributes
	readonlyAction := map[string]struct{}{
		"get":      {},
		"list":     {},
		"watch":    {},
		"proxy":    {},
		"redirect": {},
	}
	var ret []k8s_testing.Action
	for _, action := range actions {
		if _, readonly := readonlyAction[action.GetVerb()]; readonly {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}
