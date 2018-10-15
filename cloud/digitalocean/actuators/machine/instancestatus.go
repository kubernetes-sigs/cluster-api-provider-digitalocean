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

package machine

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/cluster-api/pkg/util"
)

// InstanceStatusAnnotationKey is the key of annotation which will hold the status of a machine.
// In long term, we should find a better way to handle this.
const InstanceStatusAnnotationKey = "instance-status"

// instanceStatus is used for conversion between status in annotation and real machine object.
type instanceStatus *clusterv1.Machine

// instanceStatus returns machine object based on instance status from annotation.
func (do *DOClient) instanceStatus(machine *clusterv1.Machine) (instanceStatus, error) {
	if do.v1Alpha1Client == nil {
		return nil, nil
	}
	currentMachine, err := util.GetMachineIfExists(do.v1Alpha1Client.Machines(machine.Namespace), machine.Name)
	if err != nil {
		return nil, err
	}
	if currentMachine == nil {
		return nil, nil
	}
	return do.machineInstanceStatus(machine)
}

// machineInstanceStatus converts string from annotation to the machine object.
func (do *DOClient) machineInstanceStatus(machine *clusterv1.Machine) (instanceStatus, error) {
	if machine.Annotations == nil {
		return nil, nil
	}

	a := machine.Annotations[InstanceStatusAnnotationKey]
	if a == "" {
		return nil, nil
	}

	var status clusterv1.Machine
	serializer := json.NewSerializer(json.DefaultMetaFactory, do.scheme, do.scheme, false)
	_, _, err := serializer.Decode([]byte(a), &schema.GroupVersionKind{
		Group:   "",
		Version: "cluster.k8s.io/v1alpha1",
		Kind:    "Machine",
	}, &status)
	if err != nil {
		return nil, fmt.Errorf("error decoding machine status: %v", err)
	}

	return instanceStatus(&status), nil
}

// setMachineInstanceStatus writes the instance status to the annotation.
func (do *DOClient) setMachineInstanceStatus(machine *clusterv1.Machine, status instanceStatus) (instanceStatus, error) {
	// Avoid status within status.
	status.Annotations[InstanceStatusAnnotationKey] = ""

	b := []byte{}
	buff := bytes.NewBuffer(b)
	serializer := json.NewSerializer(json.DefaultMetaFactory, do.scheme, do.scheme, false)

	err := serializer.Encode((*clusterv1.Machine)(status), buff)
	if err != nil {
		return nil, err
	}

	if machine.Annotations == nil {
		machine.Annotations = make(map[string]string)
	}
	machine.Annotations[InstanceStatusAnnotationKey] = buff.String()

	return machine, nil
}

// updateInstanceStatus updates the machine object with the new annotation.
func (do *DOClient) updateInstanceStatus(machine *clusterv1.Machine) error {
	if do.v1Alpha1Client == nil {
		return nil
	}

	status := instanceStatus(machine)
	currentMachine, err := util.GetMachineIfExists(do.v1Alpha1Client.Machines(machine.Namespace), machine.Name)
	if err != nil {
		return nil
	}
	if currentMachine == nil {
		return fmt.Errorf("status cannot be updated because machine %q does not exist anymore", machine.Name)
	}

	m, err := do.setMachineInstanceStatus(currentMachine, status)
	if err != nil {
		return err
	}

	_, err = do.v1Alpha1Client.Machines(machine.Namespace).Update(m)
	return err
}
