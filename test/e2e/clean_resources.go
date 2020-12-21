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

package e2e

import (
	"context"
	"os"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

// CleanDOResources clean any resource leftover from the tests.
func CleanDOResources(clusterName string) error {
	token := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if token == "" {
		return errors.New("missing DO token")
	}

	ctx := context.Background()
	client := godo.NewFromToken(token)

	droplets, err := dropletList(ctx, client)
	if err != nil {
		return errors.Wrap(err, "failed to list droplets")
	}

	for _, droplet := range droplets {
		if strings.Contains(droplet.Name, clusterName) {
			_, err := client.Droplets.Delete(ctx, droplet.ID)
			if err != nil {
				return errors.Wrapf(err, "failed to delete droplet %d/%s", droplet.ID, droplet.Name)
			}
		}
	}

	lbs, err := lbList(ctx, client)
	if err != nil {
		return errors.Wrap(err, "failed to list droplets")
	}

	for _, lb := range lbs {
		if strings.Contains(lb.Name, clusterName) {
			_, err := client.LoadBalancers.Delete(ctx, lb.ID)
			if err != nil {
				return errors.Wrapf(err, "failed to delete loadbalancer %s/%s", lb.ID, lb.Name)
			}
		}
	}

	return nil
}

func dropletList(ctx context.Context, client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		list = append(list, droplets...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return list, nil
}

func lbList(ctx context.Context, client *godo.Client) ([]godo.LoadBalancer, error) {
	list := []godo.LoadBalancer{}

	opt := &godo.ListOptions{}
	for {
		lbs, resp, err := client.LoadBalancers.List(ctx, opt)
		if err != nil {
			return nil, err
		}

		list = append(list, lbs...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return list, nil
}
