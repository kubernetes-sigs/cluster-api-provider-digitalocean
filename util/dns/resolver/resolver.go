/*
Copyright 2021 The Kubernetes Authors.

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

// Package resolver implement the dns resolver and apis.
package resolver

import (
	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

// DNSResolver have the interfaces to be implemented.
type DNSResolver interface {
	Query(servers []string, msg *dns.Msg) (*dns.Msg, error)
	LocalQuery(msg *dns.Msg) (*dns.Msg, error)
}

type resolver struct {
	config *dns.ClientConfig
	client *dns.Client
}

// NewDNSResolver creates a new client resolver.
func NewDNSResolver() (DNSResolver, error) {
	dnsConfig, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get DNS config")
	}

	sq := &resolver{
		config: dnsConfig,
		client: new(dns.Client),
	}

	return sq, nil
}

func (r *resolver) Query(servers []string, msg *dns.Msg) (*dns.Msg, error) {
	for _, server := range servers {
		r, _, err := r.client.Exchange(msg, server+":"+r.config.Port)
		if err != nil {
			return nil, err
		}

		if r == nil || r.Rcode == dns.RcodeNameError || r.Rcode == dns.RcodeSuccess {
			return r, err
		}
	}

	return nil, errors.New("No name server to answer the question")
}

func (r *resolver) LocalQuery(msg *dns.Msg) (*dns.Msg, error) {
	return r.Query(r.config.Servers, msg)
}
