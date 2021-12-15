package main

import (
	"bytes"
	"context"
	"crypto"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/cosign/pkg/cosign"
	cremote "github.com/sigstore/cosign/pkg/cosign/remote"
	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/kms"
	"github.com/sigstore/sigstore/pkg/signature/options"
	"github.com/sigstore/sigstore/pkg/signature/payload"
)

//dummy

// NewKMSSignerVerifier returns signature.SignerVerifier for a provider based on a keyRef.
func NewKMSSignerVerifier(ctx context.Context, keyRef string) (signature.SignerVerifier, error) {
	for prefix := range kms.ProvidersMux().Providers() {
		if strings.HasPrefix(keyRef, prefix) {
			return kms.Get(ctx, keyRef, crypto.SHA256)
		}
	}
	return nil, fmt.Errorf("unsupported keyRef: %v", keyRef)
}

// Sign generates signature for the image ref using provides signature.SignerVerifier and pushes it to the registry.
func Sign(ctx context.Context, sv signature.SignerVerifier, ref name.Reference, dryRun bool, auth authn.Authenticator) error {
	get, err := remote.Get(ref, remote.WithContext(ctx))
	if err != nil {
		if ifRefNotFound(err) && dryRun {
			log.Debug("Skipped signature signing in dry-run mode")
			return nil
		}
		return err
	}
	repo := ref.Context()
	img := repo.Digest(get.Digest.String())
	pl, err := payload.Cosign{Image: img}.MarshalJSON()
	if err != nil {
		return fmt.Errorf("payload %w", err)
	}
	sig, err := sv.SignMessage(bytes.NewReader(pl), options.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("signing %w", err)
	}

	imgHash, err := v1.NewHash(img.Identifier())
	if err != nil {
		return err
	}
	sigRef := cosign.AttachedImageTag(img.Context(), imgHash, cosign.SignatureTagSuffix)

	uo := cremote.UploadOpts{
		DupeDetector: sv,
		RemoteOpts:   []remote.Option{remote.WithAuth(auth), remote.WithContext(ctx)},
	}
	if !dryRun {
		log.Debug("Uploading signature to the registry")
		_, err = cremote.UploadSignature(sig, pl, sigRef, uo)
		if err != nil {
			return fmt.Errorf("upload signature %w", err)
		}
	}

	return nil
}

// Verify checks the provided image signature against the provided signature.SignerVerifier. It returns error if the check fails.
func Verify(ctx context.Context, sv signature.SignerVerifier, sigRef name.Reference, auth authn.Authenticator) error {
	co := cosign.CheckOpts{
		SigVerifier:        sv,
		RegistryClientOpts: []remote.Option{remote.WithContext(ctx), remote.WithAuth(auth)},
	}
	_, err := cosign.Verify(ctx, sigRef, &co)
	if err != nil {
		return err
	}
	return nil
}
