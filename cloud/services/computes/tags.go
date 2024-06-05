/*
Copyright 2023 The Kubernetes Authors.

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
	"strconv"

	"github.com/digitalocean/godo"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
)

// Retag checks if a droplet contains exactly the tags that this operator expects and retags if otherwise.
func (s *Service) Retag(droplet *godo.Droplet, scope *scope.MachineScope) error {
	tags := infrav1.BuildTags(infrav1.BuildTagParams{
		ClusterName: infrav1.DOSafeName(s.scope.Name()),
		ClusterUID:  s.scope.UID(),
		Name:        infrav1.DOSafeName(scope.Name()),
		Role:        scope.Role(),
		Additional:  scope.AdditionalTags(),
	})

	for _, tag := range tags {
		if !contains(droplet.Tags, tag) {
			if err := s.createTag(tag); err != nil {
				return err
			}
			_, err := s.scope.Tags.TagResources(s.ctx, tag, &godo.TagResourcesRequest{
				Resources: []godo.Resource{
					{
						ID:   strconv.Itoa(droplet.ID),
						Type: godo.DropletResourceType,
					},
				},
			})
			if err != nil {
				return err
			}
		}
	}

	for _, t := range droplet.Tags {
		if !contains(tags, t) {
			_, err := s.scope.Tags.UntagResources(s.ctx, t, &godo.UntagResourcesRequest{
				Resources: []godo.Resource{
					{
						ID:   strconv.Itoa(droplet.ID),
						Type: godo.DropletResourceType,
					},
				},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) createTag(tag string) error {
	_, _, err := s.scope.Tags.Get(s.ctx, tag)
	switch {
	case apierrors.IsNotFound(err):
		_, _, err := s.scope.Tags.Create(s.ctx, &godo.TagCreateRequest{Name: tag})
		return err
	default:
		return err
	}
}

func contains(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
