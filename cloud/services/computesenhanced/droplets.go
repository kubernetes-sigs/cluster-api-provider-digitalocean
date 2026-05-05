/*
Copyright 2026 The Kubernetes Authors.

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

// Package computesenhanced contains a custom droplet service implementation for raw interactions with the droplet API
// such as anti affinity
package computesenhanced

import (
	"context"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
)

// DropletsService extends the DropletsService interface with methods for handling creating droplets
// that are anti affinity aware.
type DropletsService interface {
	godo.DropletsService
	CreateWithAffinity(tx context.Context, req *godo.DropletCreateRequest, antiAffinityKey string) (*DropletCreateWithAntiAffinityResponse, error)
}

// DropletCreateRequestWithAntiAffinityKey uses embedding with godo.DropletCreateRequest to allow for the addition of HV affinity
// for use in the HTTP client request directly. It embeds the original droplet create request and
// adds the additional anti affinity group.
type DropletCreateRequestWithAntiAffinityKey struct {
	godo.DropletCreateRequest
	AntiAffinityKey string `json:"anti_affinity_key,omitempty"`
}

// DropletCreateWithAntiAffinityResponse is a struct representing the raw response from the droplet service
// This is added here as the response object was unexported in godo https://github.internal.digitalocean.com/digitalocean/godo/blob/4c47a13b03a2c36599700938b1c60c2265103765/droplets.go#L150.
type DropletCreateWithAntiAffinityResponse struct {
	*godo.Droplet `json:"droplet"`
	RawResp       *godo.Response
}

// DropletServiceOp implements the DropletService interface to allow for creating affinity based droplets.
type DropletServiceOp struct {
	godo.DropletsService
	client  *godo.Client
	APIPath string
}

// CreateWithAffinity creates a new droplet with an optional anti affinity key specified. If the key is provided the droplet API
// will do a best effort attempt to co locate all droplets with the same key on different hypervisors. If no key is provided droplets
// will be created with no affinity considerations.
func (d *DropletServiceOp) CreateWithAffinity(ctx context.Context, req *godo.DropletCreateRequest, antiAffinityKey string) (*DropletCreateWithAntiAffinityResponse, error) {
	root := new(DropletCreateWithAntiAffinityResponse)
	reqWithAffinity := &DropletCreateRequestWithAntiAffinityKey{
		DropletCreateRequest: *req,
		AntiAffinityKey:      antiAffinityKey,
	}
	apiReq, err := d.client.NewRequest(ctx, http.MethodPost, d.APIPath, reqWithAffinity)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create droplet API request")
	}
	resp, err := d.client.Do(ctx, apiReq, root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new droplet")
	}
	root.RawResp = resp
	return root, nil
}

// NewDropletService creates a new enhanced droplet service instance with additional functions.
func NewDropletService(client *godo.Client, ds godo.DropletsService) *DropletServiceOp {
	return &DropletServiceOp{
		DropletsService: ds,
		client:          client,
		APIPath:         "v2/droplets",
	}
}
