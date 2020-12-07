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

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/digitalocean/godo"
)

const timeToCleanInHours = 12

func main() {
	log.Println("Starting DO Janitor")
	token := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	if token == "" {
		log.Fatal("missing DO token")
	}

	ctx := context.Background()
	client := godo.NewFromToken(token)

	droplets, err := dropletList(ctx, client)
	if err != nil {
		log.Fatalf("fail to list droplets: %+v", err.Error())
	}

	for _, droplet := range droplets {
		dropletCreated, err := time.Parse(time.RFC3339, droplet.Created)
		if err != nil {
			log.Fatalf("failt parsing time: %+v", err.Error())
		}

		hours := time.Since(dropletCreated).Hours()
		if hours >= timeToCleanInHours {
			log.Printf("%s is older than %d hours will terminate\n", droplet.Name, timeToCleanInHours)
			_, err := client.Droplets.Delete(ctx, droplet.ID)
			if err != nil {
				log.Printf("fail to delete droplet %s: %+v\n", droplet.Name, err.Error())
				continue
			}

			log.Printf("droplet %s terminated\n", droplet.Name)
		}
	}

	lbs, err := lbList(ctx, client)
	if err != nil {
		log.Fatalf("fail to failed list LoadBalancer: %+v", err.Error())
	}

	for _, lb := range lbs {
		lbCreated, err := time.Parse(time.RFC3339, lb.Created)
		if err != nil {
			log.Fatalf("failt parsing time: %+v", err.Error())
		}

		hours := time.Since(lbCreated).Hours()
		if hours >= timeToCleanInHours {
			log.Printf("%s is older than %d hours will terminate\n", lb.Name, timeToCleanInHours)
			_, err := client.LoadBalancers.Delete(ctx, lb.ID)
			if err != nil {
				log.Printf("fail to delete droplet %s: %+v\n", lb.Name, err.Error())
				continue
			}

			log.Printf("droplet %s terminated\n", lb.Name)
		}
	}

	log.Println("Done DO Janitor")
	os.Exit(0)
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
