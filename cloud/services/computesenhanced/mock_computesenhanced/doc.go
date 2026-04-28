/*
Copyright The Kubernetes Authors.

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

//go:generate ../../../../hack/tools/bin/mockgen -destination droplets_mock.go -package mock_computesenhanced sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/computesenhanced DropletsService
//go:generate /usr/bin/env bash -c "cat ../../../../hack/boilerplate/boilerplate.generatego.txt droplets_mock.go > _droplets_mock.go && mv _droplets_mock.go droplets_mock.go"
package mock_computesenhanced // nolint
