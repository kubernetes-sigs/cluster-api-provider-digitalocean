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

package v1alpha4

import (
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
)

// ConvertTo converts this DOCluster to the Hub version (v1beta1).
func (src *DOCluster) ConvertTo(dstRaw conversion.Hub) error { // nolint
	dst := dstRaw.(*infrav1.DOCluster)
	if err := Convert_v1alpha4_DOCluster_To_v1beta1_DOCluster(src, dst, nil); err != nil {
		return err
	}

	// Manually restore data from annotations
	restored := &infrav1.DOCluster{}
	if ok, err := utilconversion.UnmarshalData(src, restored); err != nil || !ok {
		return err
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DOCluster) ConvertFrom(srcRaw conversion.Hub) error { // nolint
	src := srcRaw.(*infrav1.DOCluster)
	if err := Convert_v1beta1_DOCluster_To_v1alpha4_DOCluster(src, dst, nil); err != nil {
		return err
	}

	// Preserve Hub data on down-conversion.
	if err := utilconversion.MarshalData(src, dst); err != nil {
		return err
	}

	return nil
}

// ConvertTo converts this DOClusterList to the Hub version (v1beta1).
func (src *DOClusterList) ConvertTo(dstRaw conversion.Hub) error { // nolint
	dst := dstRaw.(*infrav1.DOClusterList)
	return Convert_v1alpha4_DOClusterList_To_v1beta1_DOClusterList(src, dst, nil)
}

// ConvertFrom converts from the Hub version (v1beta1) to this version.
func (dst *DOClusterList) ConvertFrom(srcRaw conversion.Hub) error { // nolint
	src := srcRaw.(*infrav1.DOClusterList)
	return Convert_v1beta1_DOClusterList_To_v1alpha4_DOClusterList(src, dst, nil)
}
