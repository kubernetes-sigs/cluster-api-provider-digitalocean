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

// Package dns implement the dns operations.
package dns

import (
	"fmt"
	"strings"
	"sync"

	"github.com/miekg/dns"
	"sigs.k8s.io/cluster-api-provider-digitalocean/util/dns/resolver"
)

var (
	syncOnce        sync.Once
	defaultResolver resolver.DNSResolver
)

func init() {
	defaultResolver = resolver.NewFakeDNSResolver([]*dns.Msg{})
}

// InitFromDNSResolver ...
func InitFromDNSResolver(resolver resolver.DNSResolver) {
	syncOnce.Do(func() {
		defaultResolver = resolver
	})
}

// ToFQDN ...
func ToFQDN(name, domain string) string {
	fqdn := fmt.Sprintf("%s.%s", name, domain)
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	return fqdn
}

// CheckDNSPropagated checks if the DNS is propagated.
func CheckDNSPropagated(fqdn, ip string) (bool, error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeA)

	authNS, err := LookupAuthoritativeServer(fqdn)
	if err != nil {
		return false, err
	}

	resp, err := defaultResolver.Query([]string{authNS}, m)
	if err != nil {
		return false, err
	}

	for _, ans := range resp.Answer {
		switch a := ans.(type) {
		case *dns.A:
			if a.A.String() == ip {
				return true, nil
			}
		default:
			continue
		}
	}

	return false, nil
}

// LookupAuthoritativeServer ...
func LookupAuthoritativeServer(fqdn string) (string, error) {
	m := new(dns.Msg)
	m.SetQuestion(fqdn, dns.TypeSOA)
	m.MsgHdr.RecursionDesired = true
	resp, err := defaultResolver.LocalQuery(m)
	if err != nil {
		return "", err
	}

	if len(resp.Ns) < 1 {
		return "", fmt.Errorf("didn't get DNS authority section")
	}

	var authNS string
	for _, rr := range resp.Ns {
		soa, ok := rr.(*dns.SOA)
		if !ok {
			continue
		}
		authNS = soa.Ns
		break
	}

	if authNS == "" {
		return "", fmt.Errorf("didn't find authority NS")
	}

	return authNS, nil
}
