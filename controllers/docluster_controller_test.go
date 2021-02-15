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

package controllers

import (
	"fmt"
	"net"
	"testing"

	"github.com/miekg/dns"
)

type fakeResolver struct {
	// TODO: maybe we'll also need to check requests
	responses []*dns.Msg
	c         int
}

func (f *fakeResolver) Query(servers []string, msg *dns.Msg) (*dns.Msg, error) {
	r := f.responses[f.c]
	f.c++
	return r, nil
}

func (f *fakeResolver) LocalQuery(msg *dns.Msg) (*dns.Msg, error) {
	return f.Query(nil, msg)
}

func newFakeResolver(responses []*dns.Msg) *fakeResolver {
	return &fakeResolver{
		responses: responses,
	}
}

func newDNSTypeSOAMsg(name, ns string) *dns.Msg {
	return &dns.Msg{
		Ns: []dns.RR{&dns.SOA{
			Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeSOA, Class: dns.ClassINET},
			Ns:  ns,
		}},
	}
}

func newDNSTypeAMsg(name string, ip net.IP) *dns.Msg {
	return &dns.Msg{
		Answer: []dns.RR{&dns.A{
			Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET},
			A:   ip,
		}},
	}
}

func TestDNSPropagationCheck(t *testing.T) {
	host := "foo"
	domain := "test.go"
	fqdn := fmt.Sprintf("%s.%s.", host, domain)
	hostIP := net.IPv4(9, 9, 9, 9)
	fakeAuthoritativeNSName := "ns1.somewhere.com."

	tt := []struct {
		desc           string
		fakeDNS        *fakeResolver
		wantPropagated bool
		shouldSucc     bool
	}{
		{
			desc: "already propagated DNS record",
			fakeDNS: newFakeResolver([]*dns.Msg{
				newDNSTypeSOAMsg(host, fakeAuthoritativeNSName),
				newDNSTypeAMsg(fqdn, hostIP),
			}),
			wantPropagated: true,
			shouldSucc:     true,
		},
		{
			desc: "non-propagated DNS record",
			fakeDNS: newFakeResolver([]*dns.Msg{
				newDNSTypeSOAMsg(host, fakeAuthoritativeNSName),
				&dns.Msg{},
			}),
			wantPropagated: false,
			shouldSucc:     true,
		},
		{
			desc: "wrong IP propagated",
			fakeDNS: newFakeResolver([]*dns.Msg{
				newDNSTypeSOAMsg(host, fakeAuthoritativeNSName),
				newDNSTypeAMsg(fqdn, net.IPv4(192, 168, 1, 1)),
			}),
			wantPropagated: false,
			shouldSucc:     true,
		},
		{
			desc: "missing authority section",
			fakeDNS: newFakeResolver([]*dns.Msg{
				&dns.Msg{},
			}),
			wantPropagated: false,
			shouldSucc:     false,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			r := DOClusterReconciler{
				DNS: tc.fakeDNS,
			}
			propagated, err := r.dnsIsPropagated(host, domain, hostIP.String())
			if tc.shouldSucc != (err == nil) {
				t.Errorf("got err '%v', want success %t", err, tc.shouldSucc)
			}
			if !tc.shouldSucc {
				return
			}
			if propagated != tc.wantPropagated {
				t.Errorf("got propagated %t, want %t", propagated, tc.wantPropagated)
			}
		})
	}

}
