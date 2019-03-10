/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sshutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/klog"

	"golang.org/x/crypto/ssh"

	"github.com/digitalocean/godo"
	"github.com/pborman/uuid"
)

// PubKey is used to create temporary SSH keypairs. It is used as a way to disable root passwords emails on Droplet creation.
// The reason for not hardcoding a random public key is that it would look like a backdoor
type PubKey struct {
	Name           string
	PublicKey      string
	FingerprintMD5 string
}

// NewKeyFromString converts provided public key string to public key object.
func NewKeyFromString(publicKey string) (*PubKey, error) {
	sshKeyPair, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		return nil, err
	}

	return &PubKey{
		Name:           uuid.New(),
		PublicKey:      string(publicKey),
		FingerprintMD5: ssh.FingerprintLegacyMD5(sshKeyPair),
	}, nil
}

func NewKeyFromFile(publicKeyPath string) (*PubKey, error) {
	key, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	sshKeyPair, _, _, _, err := ssh.ParseAuthorizedKey(key)
	if err != nil {
		return nil, err
	}

	return &PubKey{
		Name:           uuid.New(),
		PublicKey:      string(key),
		FingerprintMD5: ssh.FingerprintLegacyMD5(sshKeyPair),
	}, nil
}

// Create uploads the public key to DigitalOcean.
func (p *PubKey) Create(ctx context.Context, keysService godo.KeysService) error {
	existingkey, res, err := keysService.GetByFingerprint(ctx, p.FingerprintMD5)
	if err == nil && existingkey != nil && res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusAccepted {
		klog.Info("Failed to create ssh public key, the key already exists.")
		return nil
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
