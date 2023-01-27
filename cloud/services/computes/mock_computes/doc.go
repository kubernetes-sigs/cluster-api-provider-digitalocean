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

//go:generate ../../../../hack/tools/bin/mockgen -destination droplets_mock.go -package mock_computes github.com/digitalocean/godo DropletsService
//go:generate ../../../../hack/tools/bin/mockgen -destination images_mock.go -package mock_computes github.com/digitalocean/godo ImagesService
//go:generate ../../../../hack/tools/bin/mockgen -destination sshkeys_mock.go -package mock_computes github.com/digitalocean/godo KeysService
//go:generate ../../../../hack/tools/bin/mockgen -destination volumes_mock.go -package mock_computes github.com/digitalocean/godo StorageService
//go:generate /usr/bin/env bash -c "cat ../../../../hack/boilerplate/boilerplate.generatego.txt droplets_mock.go > _droplets_mock.go && mv _droplets_mock.go droplets_mock.go"
//go:generate /usr/bin/env bash -c "cat ../../../../hack/boilerplate/boilerplate.generatego.txt images_mock.go > _images_mock.go && mv _images_mock.go images_mock.go"
//go:generate /usr/bin/env bash -c "cat ../../../../hack/boilerplate/boilerplate.generatego.txt sshkeys_mock.go > _sshkeys_mock.go && mv _sshkeys_mock.go sshkeys_mock.go"
//go:generate /usr/bin/env bash -c "cat ../../../../hack/boilerplate/boilerplate.generatego.txt volumes_mock.go > _volumes_mock.go && mv _volumes_mock.go volumes_mock.go"
package mock_computes // nolint
