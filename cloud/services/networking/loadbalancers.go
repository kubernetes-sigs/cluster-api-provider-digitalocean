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

package networking

import (
	"net/http"

	"github.com/digitalocean/godo"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

// GetLoadBalancer get a LB by LB ID.
func (s *Service) GetLoadBalancer(id string) (*godo.LoadBalancer, error) {
	if id == "" {
		return nil, nil
	}

	lb, res, err := s.scope.LoadBalancers.Get(s.ctx, id)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	return lb, nil
}

// CreateLoadBalancer creates a LB.
func (s *Service) CreateLoadBalancer(spec *infrav1.DOLoadBalancer) (*godo.LoadBalancer, error) {
	clusterName := infrav1.DOSafeName(s.scope.Name())
	name := clusterName + "-" + infrav1.APIServerRoleTagValue + "-" + s.scope.UID()
	request := &godo.LoadBalancerRequest{
		Name:      name,
		Algorithm: spec.Algorithm,
		Region:    s.scope.Region(),
		ForwardingRules: []godo.ForwardingRule{
			{
				EntryProtocol:  "tcp",
				EntryPort:      spec.Port,
				TargetProtocol: "tcp",
				TargetPort:     spec.Port,
			},
		},
		HealthCheck: &godo.HealthCheck{
			Protocol:               "tcp",
			Port:                   spec.Port,
			CheckIntervalSeconds:   spec.HealthCheck.Interval,
			ResponseTimeoutSeconds: spec.HealthCheck.Timeout,
			UnhealthyThreshold:     spec.HealthCheck.UnhealthyThreshold,
			HealthyThreshold:       spec.HealthCheck.HealthyThreshold,
		},
		Tag:     infrav1.ClusterNameUIDRoleTag(clusterName, s.scope.UID(), infrav1.APIServerRoleTagValue),
		VPCUUID: s.scope.VPC().VPCUUID,
	}

	lb, _, err := s.scope.LoadBalancers.Create(s.ctx, request)
	if err != nil {
		return nil, err
	}

	return lb, nil
}

// DeleteLoadBalancer delete a LB by ID.
func (s *Service) DeleteLoadBalancer(id string) error {
	if _, err := s.scope.LoadBalancers.Delete(s.ctx, id); err != nil {
		return err
	}

	return nil
}
