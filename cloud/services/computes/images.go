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

func (s *Service) GetImage(imageSpec intstr.IntOrString) (*godo.Image, error) {
	var image *godo.Image
	var reterr error

	if imageSpec.IntValue() != 0 { // nolint
		image, _, reterr = s.scope.Images.GetByID(s.ctx, imageSpec.IntValue())
	} else if imageSpec.String() != "" && imageSpec.String() != "0" {
		image, _, reterr = s.scope.Images.GetBySlug(s.ctx, imageSpec.String())
	} else {
		reterr = errors.New("Unable to get image")
	}

	if reterr != nil {
		return nil, errors.Wrap(reterr, "Unable to get image")
	}

	return image, nil
}
