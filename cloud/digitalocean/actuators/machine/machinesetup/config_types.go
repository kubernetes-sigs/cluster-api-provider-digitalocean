/*
Copyright 2018 The Kubernetes Authors.

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

package machinesetup

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/ghodss/yaml"
)

// MachineSetupConfig interface is used to operate with the machine userdata.
// The userdata is provided as a ConfigMap and mounted in the machine-controller pod as a file.
type MachineSetupConfig interface {
	GetYaml() (string, error)
	GetImage(params *ConfigParams) (string, error)
	GetUserdata(params *ConfigParams) (string, error)
}

// config is a single machine setup config that has userdata and parameters such as image name.
type config struct {
	// Params is a list of valid combinations of ConfigParams that will map to the appropriate Userdata.
	Params []*ConfigParams `json:"machineParams"`

	// Userdata is a script used to provision instance.
	Userdata string `json:"userdata"`
}

type ConfigParams struct {
	Image    string                       `json:"image"`
	Versions clusterv1.MachineVersionInfo `json:"versions"`
}

// configList is list of configs.
type configList struct {
	Items []config `json:"items"`
}

// ValidConfigs contains parsed and valid configs.
type ValidConfigs struct {
	configList *configList
}

// ConfigWatch contains the path of the userdata file in the machine-controller pod. We're working with files instead of
// ConfigMaps directly, so we don't depend on the Kubernetes Client and the API server.
type ConfigWatch struct {
	path string
}

// NewConfigWatch returns new ConfigWatch object if the path is valid and file exists.
func NewConfigWatch(path string) (*ConfigWatch, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	return &ConfigWatch{path: path}, nil
}

// GetMachineSetupConfig returns userdata for an instnace.
func (cw *ConfigWatch) GetMachineSetupConfig() (MachineSetupConfig, error) {
	file, err := os.Open(cw.path)
	if err != nil {
		return nil, err
	}
	return parseMachineSetupYaml(file)
}

// GetYaml returns yaml from of the Config object.
func (vc *ValidConfigs) GetYaml() (string, error) {
	bytes, err := yaml.Marshal(vc.configList)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GetImage returns the name of the image from Param object.
func (vc *ValidConfigs) GetImage(params *ConfigParams) (string, error) {
	_, err := vc.matchMachineSetupConfig(params)
	if err != nil {
		return "", err
	}
	return params.Image, nil
}

func (vc *ValidConfigs) GetUserdata(params *ConfigParams) (string, error) {
	machineSetupConfig, err := vc.matchMachineSetupConfig(params)
	if err != nil {
		return "", err
	}
	return machineSetupConfig.Userdata, nil
}

func (vc *ValidConfigs) matchMachineSetupConfig(params *ConfigParams) (*config, error) {
	matchingConfigs := make([]config, 0)
	for _, conf := range vc.configList.Items {
		for _, validParams := range conf.Params {
			if params.Image != validParams.Image {
				continue
			}
			if params.Versions != validParams.Versions {
				continue
			}
			matchingConfigs = append(matchingConfigs, conf)
		}
	}

	if len(matchingConfigs) == 1 {
		return &matchingConfigs[0], nil
	} else {
		return nil, fmt.Errorf("could not find setup configs for params %#v", params)
	}
}

func parseMachineSetupYaml(reader io.Reader) (*ValidConfigs, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	configList := &configList{}
	err = yaml.Unmarshal(bytes, configList)
	if err != nil {
		return nil, err
	}

	return &ValidConfigs{configList}, nil
}
