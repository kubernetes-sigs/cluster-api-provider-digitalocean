package sshutil

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"

	"golang.org/x/crypto/ssh"

	"github.com/digitalocean/godo"
	"github.com/pborman/uuid"
)

const privateRSAKeyBitSize = 4096

// PubKey is used to create temporary SSH keypairs. It is used as a way to disable root passwords emails on Droplet creation.
// The reason for not hardcoding a random public key is that it would look like a backdoor
type PubKey struct {
	Name           string
	PublicKey      string
	FingerprintMD5 string
}

// NewKey creates a new public key. The private key is discarded.
func NewKey() (*PubKey, error) {
	tmpRSAKeyPair, err := rsa.GenerateKey(rand.Reader, privateRSAKeyBitSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create private RSA key: %v", err)
	}

	if err := tmpRSAKeyPair.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate private RSA key: %v", err)
	}

	pubKey, err := ssh.NewPublicKey(&tmpRSAKeyPair.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ssh public key: %v", err)
	}

	return &PubKey{
		Name:           uuid.New(),
		PublicKey:      string(ssh.MarshalAuthorizedKey(pubKey)),
		FingerprintMD5: ssh.FingerprintLegacyMD5(pubKey),
	}, nil
}

// Create uploads the public key to DigitalOcean.
func (p *PubKey) Create(ctx context.Context, keysService godo.KeysService) error {
	existingkey, res, err := keysService.GetByFingerprint(ctx, p.FingerprintMD5)
	if err == nil && existingkey != nil && res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusAccepted {
		return fmt.Errorf("failed to create ssh public key, the key already exists")
	}

	_, _, err = keysService.Create(ctx, &godo.KeyCreateRequest{
		PublicKey: p.PublicKey,
		Name:      p.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to create ssh public key: %v", err)
	}
	return nil
}

// Delete deletes the public key from DigitalOcean.
func (p *PubKey) Delete(ctx context.Context, keysService godo.KeysService) error {
	_, err := keysService.DeleteByFingerprint(ctx, p.FingerprintMD5)
	if err != nil {
		return fmt.Errorf("failed to remove a temporary ssh key with fingerprint %s: %v", p.FingerprintMD5, err)
	}
	return nil
}
