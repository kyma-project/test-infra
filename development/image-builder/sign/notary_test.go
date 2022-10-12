package sign

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNotaryConfig_NewSigner(t *testing.T) {
	tmpFile := filepath.Join(os.TempDir(), "test-secret")
	os.WriteFile(tmpFile, []byte("abcd1234"), 0666)
	nc := NotaryConfig{
		Endpoint: "http://localhost/sign",
		Secret: &AuthSecretConfig{
			Path: tmpFile,
			Type: "bearer",
		},
		Timeout:      5 * time.Minute,
		RetryTimeout: 1,
	}
	s, err := nc.NewSigner()
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	ns := s.(NotarySigner)
	if ns.retryTimeout != 1 {
		t.Errorf("incorrect retryTimeout")
	}
	if ns.url != "http://localhost/sign" {
		t.Errorf("incorrect url")
	}
	if ns.c.Timeout != 5*time.Minute {
		t.Errorf("incorrect timeout")
	}
}

func TestNotarySigner_SignImages(t *testing.T) {
	tc := []struct {
		name               string
		expectErr          bool
		expectSignedImages int
		authFunc           AuthFunc
		internalErr        bool
	}{
		{
			name:               "passed signing",
			expectErr:          false,
			expectSignedImages: 2,
			authFunc:           BearerToken("abcd1234"),
		},
		{
			name:               "unauthorized",
			expectErr:          true,
			expectSignedImages: 0,
			authFunc:           nil,
		},
		{
			name:               "retries reached on internal error",
			expectErr:          true,
			expectSignedImages: 0,
			authFunc:           nil,
			internalErr:        true,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			images := []string{
				"europe-docker.pkg.dev/kyma-project/prod/keda-manager:v20221012-fc16657e",
				"europe-docker.pkg.dev/kyma-project/prod/test-infra/buildpack-golang:v20221017-733bfd36",
			}
			var signed int
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if c.internalErr {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if r.Header.Get("Authorization") != "Bearer abcd1234" {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Unauthorized!"))
					return
				}
				var sr []SigningRequest
				err := json.NewDecoder(r.Body).Decode(&sr)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
				}
				for _, i := range sr {
					fmt.Println("signing", i.NotaryGun)
					signed++
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK!"))
			}))
			ns := NotarySigner{
				url:          s.URL,
				c:            http.Client{Timeout: 5 * time.Minute},
				authFunc:     c.authFunc,
				retryTimeout: 1,
			}
			err := ns.Sign(images)
			if err != nil && !c.expectErr {
				t.Errorf("Sign() error: %v", err)
			}
			if signed != c.expectSignedImages {
				t.Errorf("signed images mismatch %v != %v", signed, c.expectSignedImages)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	token := "abcd12345678"
	expected := "Bearer abcd12345678"
	fc := BearerToken(token)
	req := httptest.NewRequest("POST", "http://localhost", nil)
	fc(req)
	if got := req.Header.Get("Authorization"); got != expected {
		t.Errorf("Bearer token did not apply: %s != %s", got, expected)
	}
}
