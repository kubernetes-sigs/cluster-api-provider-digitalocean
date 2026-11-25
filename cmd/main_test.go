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

package main

import (
	"io"
	"testing"

	"github.com/spf13/pflag"
)

func TestResolveHealthProbeBindAddress(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "defaults use the health probe bind address",
			want: ":8081",
		},
		{
			name: "health-probe-bind-address is used",
			args: []string{"--health-probe-bind-address=:7070"},
			want: ":7070",
		},
		{
			name: "deprecated health-addr alone is honored",
			args: []string{"--health-addr=:9440"},
			want: ":9440",
		},
		{
			name: "health-probe-bind-address wins over health-addr",
			args: []string{"--health-addr=:9440", "--health-probe-bind-address=:7070"},
			want: ":7070",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
			fs.SetOutput(io.Discard)
			initFlags(fs)
			if err := fs.Parse(tt.args); err != nil {
				t.Fatalf("parse flags: %v", err)
			}
			if got := resolveHealthProbeBindAddress(fs); got != tt.want {
				t.Errorf("resolveHealthProbeBindAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWebhookPortFlag(t *testing.T) {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	initFlags(fs)

	if err := fs.Parse(nil); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	if webhookPort != 9443 {
		t.Errorf("default webhook port = %d, want 9443", webhookPort)
	}

	if err := fs.Parse([]string{"--webhook-port=10443"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	if webhookPort != 10443 {
		t.Errorf("webhook port = %d, want 10443", webhookPort)
	}
}
