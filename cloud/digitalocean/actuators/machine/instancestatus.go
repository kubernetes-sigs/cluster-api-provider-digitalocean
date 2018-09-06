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
	currentMachine, err := util.GetMachineIfExists(do.v1Alpha1Client.Machines(machine.Namespace), machine.ObjectMeta.Name)
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
	if machine.ObjectMeta.Annotations == nil {
		return nil, nil
	}

	a := machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey]
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
	status.ObjectMeta.Annotations[InstanceStatusAnnotationKey] = ""

	b := []byte{}
	buff := bytes.NewBuffer(b)
	serializer := json.NewSerializer(json.DefaultMetaFactory, do.scheme, do.scheme, false)

	err := serializer.Encode((*clusterv1.Machine)(status), buff)
	if err != nil {
		return nil, err
	}

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = make(map[string]string)
	}
	machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey] = buff.String()

	return machine, nil
}

// updateInstanceStatus updates the machine object with the new annotation.
func (do *DOClient) updateInstanceStatus(machine *clusterv1.Machine) error {
	if do.v1Alpha1Client == nil {
		return nil
	}

	status := instanceStatus(machine)
	currentMachine, err := util.GetMachineIfExists(do.v1Alpha1Client.Machines(machine.Namespace), machine.ObjectMeta.Name)
	if err != nil {
		return nil
	}
	if currentMachine == nil {
		return fmt.Errorf("machine '%s' does not exist anymore, so status can't be updated", machine.ObjectMeta.Name)
	}

	m, err := do.setMachineInstanceStatus(currentMachine, status)
	if err != nil {
		return err
	}

	_, err = do.v1Alpha1Client.Machines(machine.Namespace).Update(m)
	return err
}
