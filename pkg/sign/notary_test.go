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

	"github.com/google/uuid"
)

type fakeAuthService struct {
	http.Handler
}

func (f fakeAuthService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := `{"access_token":{"claims":{"name":"sign_claim","token_ttl":"24h"},"token":"abcd1234"}}`
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func TestNotaryConfig_NewSigner(t *testing.T) {
	id := uuid.New()
	srv := httptest.NewServer(fakeAuthService{})
	tc := []struct {
		name      string
		authType  *AuthSecretConfig
		authFile  string
		expectErr bool
	}{
		{
			name:     "token type",
			authFile: "abcd1234",
			authType: &AuthSecretConfig{
				Path: filepath.Join(os.TempDir(), id.String()),
				Type: "token",
			},
			expectErr: false,
		},
		{
			name: "signify type",
			authFile: fmt.Sprintf(`endpoint: %s
payload: |
  {
    "role_id":"CD0EA3F3-C86C-4852-8092-87920F56D2D4",
    "secret_id":"70ACA8AE-81F4-48D5-BFC6-4693604DD868"
  }`, srv.URL),
			authType: &AuthSecretConfig{
				Path: filepath.Join(os.TempDir(), id.String()),
				Type: "signify",
			},
			expectErr: false,
		},
		{
			name: "backend unsupported",
			authType: &AuthSecretConfig{
				Path: filepath.Join(os.TempDir(), id.String()),
				Type: "unsupported",
			},
			expectErr: true,
		},
		{
			name:      "no auth",
			authType:  nil,
			expectErr: false,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			if c.authType != nil {
				os.WriteFile(c.authType.Path, []byte(c.authFile), 0666)
			}
			nc := NotaryConfig{
				Endpoint:     "http://localhost/sign",
				Secret:       c.authType,
				Timeout:      5 * time.Minute,
				RetryTimeout: 1,
			}
			s, err := nc.NewSigner()
			if err != nil && !c.expectErr {
				t.Errorf(err.Error())
			}
			if s != nil {
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
		})
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
			authFunc:           AuthToken("abcd1234"),
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
				if r.Header.Get("Authorization") != "Token abcd1234" {
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

func TestAuthToken(t *testing.T) {
	token := "abcd12345678" // gitleaks:allow
	expected := "Token abcd12345678"
	fc := AuthToken(token)
	req := httptest.NewRequest("POST", "http://localhost", nil)
	fc(req)
	if got := req.Header.Get("Authorization"); got != expected {
		t.Errorf("Bearer token did not apply: %s != %s", got, expected)
	}
}

func TestSignifyAuth(t *testing.T) {
	srv := httptest.NewServer(fakeAuthService{})
	expected := "Bearer abcd1234"
	jwts := SignifySecret{
		Endpoint: srv.URL,
		Payload:  `{"role_id":"CD0EA3F3-C86C-4852-8092-87920F56D2D4","secret_id":"70ACA8AE-81F4-48D5-BFC6-4693604DD868"}`, // gitleaks:allow
	}
	a, err := SignifyAuth(jwts)
	if err != nil {
		t.Fail()
	}
	req := httptest.NewRequest("POST", "http://localhost", nil)
	req = a(req)
	if got := req.Header.Get("Authorization"); got != expected {
		t.Errorf("Bearer token did not apply: %s != %s", got, expected)
	}
}
