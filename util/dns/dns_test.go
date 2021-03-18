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

package dns

import (
	"fmt"
	"net"
	"testing"

	"github.com/miekg/dns"
	"sigs.k8s.io/cluster-api-provider-digitalocean/util/dns/resolver"
)

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

func TestCheckDNSPropagated(t *testing.T) {
	host := "foo"
	domain := "test.go"
	fqdn := fmt.Sprintf("%s.%s.", host, domain)
	hostIP := net.IPv4(9, 9, 9, 9)
	fakeAuthoritativeNSName := "ns1.somewhere.com."

	type args struct {
		fqdn string
		ip   string
	}
	tests := []struct {
		name           string
		args           args
		fakeResolver   *resolver.FakeDNSResolver
		wantPropagated bool
		wantErr        bool
	}{
		{
			name: "already propagated DNS record",
			args: args{
				fqdn: ToFQDN(host, domain),
				ip:   "9.9.9.9",
			},
			fakeResolver: resolver.NewFakeDNSResolver([]*dns.Msg{
				newDNSTypeSOAMsg(host, fakeAuthoritativeNSName),
				newDNSTypeAMsg(fqdn, hostIP),
			}),
			wantPropagated: true,
		},
		{
			name: "non-propagated DNS record",
			args: args{
				fqdn: ToFQDN(host, domain),
				ip:   "9.9.9.9",
			},
			fakeResolver: resolver.NewFakeDNSResolver([]*dns.Msg{
				newDNSTypeSOAMsg(host, fakeAuthoritativeNSName),
				{},
			}),
			wantPropagated: false,
		},
		{
			name: "wrong IP propagated",
			args: args{
				fqdn: ToFQDN(host, domain),
				ip:   "9.9.9.9",
			},
			fakeResolver: resolver.NewFakeDNSResolver([]*dns.Msg{
				newDNSTypeSOAMsg(host, fakeAuthoritativeNSName),
				newDNSTypeAMsg(fqdn, net.IPv4(192, 168, 1, 1)),
			}),
			wantPropagated: false,
		},
		{
			name: "missing authority section",
			args: args{
				fqdn: ToFQDN(host, domain),
				ip:   "9.9.9.9",
			},
			fakeResolver: resolver.NewFakeDNSResolver([]*dns.Msg{
				{},
			}),
			wantPropagated: false,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultResolver = tt.fakeResolver
			propagated, err := CheckDNSPropagated(tt.args.fqdn, tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDNSPropagated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if propagated != tt.wantPropagated {
				t.Errorf("CheckDNSPropagated() propagated = %v, wantPropagated %v", propagated, tt.wantPropagated)
			}
		})
	}
}
