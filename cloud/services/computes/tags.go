package computes

import (
	"github.com/digitalocean/godo"
	"k8s.io/apimachinery/pkg/api/errors"
	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"strconv"
)

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

	// do we ever add tags to CPC droplets in another way or manually that might be deleted by the below logic?
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
	case errors.IsNotFound(err):
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
