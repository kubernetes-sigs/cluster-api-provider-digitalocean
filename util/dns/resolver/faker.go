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

package resolver

import (
	"github.com/miekg/dns"
)

type FakeDNSResolver struct {
	expectedMsg []*dns.Msg
	c           int
}

func NewFakeDNSResolver(expectedMsg []*dns.Msg) *FakeDNSResolver {
	return &FakeDNSResolver{
		expectedMsg: expectedMsg,
	}
}

func (f *FakeDNSResolver) Query(servers []string, msg *dns.Msg) (*dns.Msg, error) {
	if len(f.expectedMsg) <= f.c {
		return msg, nil
	}

	r := f.expectedMsg[f.c]
	f.c++
	return r, nil
}

func (f *FakeDNSResolver) LocalQuery(msg *dns.Msg) (*dns.Msg, error) {
	return f.Query(nil, msg)
}
