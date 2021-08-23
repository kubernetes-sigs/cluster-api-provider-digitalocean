/*
Copyright 2020 The Kubernetes Authors.

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

package computes

import (
	"github.com/digitalocean/godo"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/util/intstr"
)

// GetSSHKey return the public ssh key stored in DO.
func (s *Service) GetSSHKey(sshkey intstr.IntOrString) (*godo.Key, error) {
	var keys *godo.Key
	var reterr error

	if sshkey.IntValue() != 0 { // nolint
		keys, _, reterr = s.scope.Keys.GetByID(s.ctx, sshkey.IntValue())
	} else if sshkey.String() != "" && sshkey.String() != "0" {
		keys, _, reterr = s.scope.Keys.GetByFingerprint(s.ctx, sshkey.String())
	} else {
		reterr = errors.New("Missing key id or fingerprint")
	}

	if reterr != nil {
		return nil, reterr
	}

	return keys, nil
}
