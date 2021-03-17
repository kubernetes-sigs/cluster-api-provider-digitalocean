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
	"fmt"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func (s *Service) GetImageID(imageSpec intstr.IntOrString) (int, error) {
	var image *godo.Image

	if imageSpec.IntValue() != 0 { // nolint
		return imageSpec.IntValue(), nil
	}

	imageSpecStr := imageSpec.String()
	if imageSpecStr == "" || imageSpecStr == "0" {
		return 0, fmt.Errorf("invalid image spec string %q", imageSpecStr)
	}

	image, _, err := s.scope.Images.GetBySlug(s.ctx, imageSpecStr)
	if err != nil {
		return 0, errors.Wrap(err, "Unable to get image")
	}

	return image.ID, nil
}
