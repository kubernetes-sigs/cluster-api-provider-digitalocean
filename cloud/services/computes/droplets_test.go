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

package computes

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1alpha4"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/computes/mock_computes"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(infrav1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
}

func TestService_GetDroplet(t *testing.T) {
	os.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "super-secret-token")
	defer os.Unsetenv("DIGITALOCEAN_ACCESS_TOKEN")

	mctrl := gomock.NewController(t)
	defer mctrl.Finish()

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		expect  func(md *mock_computes.MockDropletsServiceMockRecorder)
		want    *godo.Droplet
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				id: "123456",
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder) {
				md.Get(gomock.Any(), 123456).
					Return(
						&godo.Droplet{
							ID:   123456,
							Name: "test-droplet",
						},
						&godo.Response{},
						nil,
					)
			},
			want: &godo.Droplet{
				ID:   123456,
				Name: "test-droplet",
			},
		},
		{
			name: "id is empty string (should not return an error)",
			args: args{
				id: "",
			},
			expect:  func(md *mock_computes.MockDropletsServiceMockRecorder) {},
			wantErr: false,
		},
		{
			name: "id is not numeric (should return an error)",
			args: args{
				id: "aaa",
			},
			expect:  func(md *mock_computes.MockDropletsServiceMockRecorder) {},
			wantErr: true,
		},
		{
			name: "droplet not found (should not return an error)",
			args: args{
				id: "123456",
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder) {
				md.Get(gomock.Any(), 123456).
					Return(
						&godo.Droplet{},
						&godo.Response{
							Response: &http.Response{
								StatusCode: http.StatusNotFound,
							},
						},
						errors.New("droplet not found"),
					)
			},
			wantErr: false,
		},
		{
			name: "godo return unknown error",
			args: args{
				id: "123456",
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder) {
				md.Get(gomock.Any(), 123456).
					Return(
						&godo.Droplet{},
						&godo.Response{
							Response: &http.Response{
								StatusCode: http.StatusInternalServerError,
							},
						},
						errors.New("unexpected error"),
					)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			mdroplet := mock_computes.NewMockDropletsService(mctrl)
			cscope, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Client:    fake.NewFakeClientWithScheme(scheme),
				Cluster:   &clusterv1.Cluster{},
				DOCluster: &infrav1.DOCluster{},
				DOClients: scope.DOClients{
					Droplets: mdroplet,
				},
			})
			if err != nil {
				t.Fatalf("did not expect err: %v", err)
			}

			tt.expect(mdroplet.EXPECT())
			s := NewService(ctx, cscope)
			got, err := s.GetDroplet(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetDroplet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.GetDroplet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_CreateDroplet(t *testing.T) {
	os.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "super-secret-token")
	defer os.Unsetenv("DIGITALOCEAN_ACCESS_TOKEN")

	mctrl := gomock.NewController(t)
	defer mctrl.Finish()

	type args struct {
		cluster   *clusterv1.Cluster
		docluster *infrav1.DOCluster
		machine   *clusterv1.Machine
		domachine *infrav1.DOMachine
	}
	tests := []struct {
		name    string
		args    args
		expect  func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder)
		want    *godo.Droplet
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				md.Create(gomock.Any(), &godo.DropletCreateRequest{
					Name:    "capdo-test-control-plane-nkkxn",
					Region:  "nyc1",
					Size:    "s-2vcpu-2gb",
					SSHKeys: []godo.DropletCreateSSHKey{},
					Image: godo.DropletCreateImage{
						ID: 63624555,
					},
					UserData:          "data",
					PrivateNetworking: true,
					Volumes:           []godo.DropletCreateVolume{},
					VPCUUID:           "",
					Tags: infrav1.BuildTags(infrav1.BuildTagParams{
						ClusterName: "capdo-test",
						Name:        "capdo-test-control-plane-nkkxn",
						Role:        "apiserver",
					}),
				}).Return(&godo.Droplet{ID: 123456}, nil, nil)
			},
			want: &godo.Droplet{ID: 123456},
		},
		{
			name: "failed creating droplet (should return an error)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				md.Create(gomock.Any(), &godo.DropletCreateRequest{
					Name:    "capdo-test-control-plane-nkkxn",
					Region:  "nyc1",
					Size:    "s-2vcpu-2gb",
					SSHKeys: []godo.DropletCreateSSHKey{},
					Image: godo.DropletCreateImage{
						ID: 63624555,
					},
					UserData:          "data",
					PrivateNetworking: true,
					Volumes:           []godo.DropletCreateVolume{},
					VPCUUID:           "",
					Tags: infrav1.BuildTags(infrav1.BuildTagParams{
						ClusterName: "capdo-test",
						Name:        "capdo-test-control-plane-nkkxn",
						Role:        "apiserver",
					}),
				}).Return(&godo.Droplet{}, nil, errors.New("error creating droplet"))
			},
			wantErr: true,
		},
		{
			name: "failed getting bootstrap data (should return an error)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("no-exist-bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
			},
			wantErr: true,
		},
		{
			name: "with provided ssh key fingerprint",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
						SSHKeys: []intstr.IntOrString{
							intstr.FromString("12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53"),
						},
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				mk.GetByFingerprint(gomock.Any(), "12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53").Return(&godo.Key{ID: 12345, Fingerprint: "12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53"}, nil, nil)
				md.Create(gomock.Any(), &godo.DropletCreateRequest{
					Name:   "capdo-test-control-plane-nkkxn",
					Region: "nyc1",
					Size:   "s-2vcpu-2gb",
					SSHKeys: []godo.DropletCreateSSHKey{
						{
							ID:          12345,
							Fingerprint: "12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53",
						},
					},
					Image: godo.DropletCreateImage{
						ID: 63624555,
					},
					UserData:          "data",
					PrivateNetworking: true,
					Volumes:           []godo.DropletCreateVolume{},
					VPCUUID:           "",
					Tags: infrav1.BuildTags(infrav1.BuildTagParams{
						ClusterName: "capdo-test",
						Name:        "capdo-test-control-plane-nkkxn",
						Role:        "apiserver",
					}),
				}).Return(&godo.Droplet{ID: 123456}, nil, nil)
			},
			want: &godo.Droplet{ID: 123456},
		},
		{
			name: "with image slug (should getting image id by slug)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromString("capi-image"),
						Size:  "s-2vcpu-2gb",
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				mi.GetBySlug(gomock.Any(), "capi-image").Return(&godo.Image{ID: 63624555}, nil, nil)
				md.Create(gomock.Any(), &godo.DropletCreateRequest{
					Name:    "capdo-test-control-plane-nkkxn",
					Region:  "nyc1",
					Size:    "s-2vcpu-2gb",
					SSHKeys: []godo.DropletCreateSSHKey{},
					Image: godo.DropletCreateImage{
						ID: 63624555,
					},
					UserData:          "data",
					PrivateNetworking: true,
					Volumes:           []godo.DropletCreateVolume{},
					VPCUUID:           "",
					Tags: infrav1.BuildTags(infrav1.BuildTagParams{
						ClusterName: "capdo-test",
						Name:        "capdo-test-control-plane-nkkxn",
						Role:        "apiserver",
					}),
				}).Return(&godo.Droplet{ID: 123456}, nil, nil)
			},
			want: &godo.Droplet{ID: 123456},
		},
		{
			name: "failed getting image (should return an error)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromString("capi-image"),
						Size:  "s-2vcpu-2gb",
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				mi.GetBySlug(gomock.Any(), "capi-image").Return(nil, nil, errors.New("error getting image"))
			},
			wantErr: true,
		},
		{
			name: "with provided ssh key fingerprint but failed getting keys (should return an error)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
						SSHKeys: []intstr.IntOrString{
							intstr.FromString("12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53"),
						},
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				mk.GetByFingerprint(gomock.Any(), "12:f8:7e:78:61:b4:bf:e2:de:24:15:96:4e:d4:72:53").Return(&godo.Key{}, nil, errors.New("error getting keys"))
			},
			wantErr: true,
		},
		{
			name: "with provided data disk",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
						DataDisks: []infrav1.DataDisk{
							{
								NameSuffix:      "etcd",
								FilesystemType:  "ext4",
								FilesystemLabel: "etcd_data",
								DiskSizeGB:      256,
							},
						},
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				ms.ListVolumes(gomock.Any(), &godo.ListVolumeParams{Name: "capdo-test-control-plane-nkkxn-etcd", Region: "nyc1"}).Return([]godo.Volume{{ID: "1234"}}, nil, nil)
				md.Create(gomock.Any(), &godo.DropletCreateRequest{
					Name:    "capdo-test-control-plane-nkkxn",
					Region:  "nyc1",
					Size:    "s-2vcpu-2gb",
					SSHKeys: []godo.DropletCreateSSHKey{},
					Image: godo.DropletCreateImage{
						ID: 63624555,
					},
					UserData:          "data",
					PrivateNetworking: true,
					Volumes: []godo.DropletCreateVolume{
						{
							ID: "1234",
						},
					},
					VPCUUID: "",
					Tags: infrav1.BuildTags(infrav1.BuildTagParams{
						ClusterName: "capdo-test",
						Name:        "capdo-test-control-plane-nkkxn",
						Role:        "apiserver",
					}),
				}).Return(&godo.Droplet{ID: 123456}, nil, nil)
			},
			want: &godo.Droplet{ID: 123456},
		},
		{
			name: "with provided data disk but volume doesn't exists (should return an error)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
						DataDisks: []infrav1.DataDisk{
							{
								NameSuffix:      "etcd",
								FilesystemType:  "ext4",
								FilesystemLabel: "etcd_data",
								DiskSizeGB:      256,
							},
						},
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				ms.ListVolumes(gomock.Any(), &godo.ListVolumeParams{Name: "capdo-test-control-plane-nkkxn-etcd", Region: "nyc1"}).Return([]godo.Volume{}, nil, nil)
			},
			wantErr: true,
		},
		{
			name: "with provided data disk but failed getting volume (should return an error)",
			args: args{
				cluster: &clusterv1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
				},
				docluster: &infrav1.DOCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "capdo-test",
						Namespace: "default",
					},
					Spec: infrav1.DOClusterSpec{
						Region: "nyc1",
					},
				},
				machine: &clusterv1.Machine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-h8f6l",
						Labels: map[string]string{
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Spec: clusterv1.MachineSpec{
						Bootstrap: clusterv1.Bootstrap{
							DataSecretName: pointer.StringPtr("bootstrap-data"),
						},
					},
				},
				domachine: &infrav1.DOMachine{
					ObjectMeta: metav1.ObjectMeta{
						Name: "capdo-test-control-plane-nkkxn",
					},
					Spec: infrav1.DOMachineSpec{
						Image: intstr.FromInt(63624555),
						Size:  "s-2vcpu-2gb",
						DataDisks: []infrav1.DataDisk{
							{
								NameSuffix:      "etcd",
								FilesystemType:  "ext4",
								FilesystemLabel: "etcd_data",
								DiskSizeGB:      256,
							},
						},
					},
				},
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder, mk *mock_computes.MockKeysServiceMockRecorder, mi *mock_computes.MockImagesServiceMockRecorder, ms *mock_computes.MockStorageServiceMockRecorder) {
				ms.ListVolumes(gomock.Any(), &godo.ListVolumeParams{Name: "capdo-test-control-plane-nkkxn-etcd", Region: "nyc1"}).Return([]godo.Volume{}, nil, errors.New("error getting volume"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "bootstrap-data",
				},
				Data: map[string][]byte{
					"value": []byte("data"),
				},
			}

			fclient := fake.NewFakeClientWithScheme(scheme, secret)
			mscope, err := scope.NewMachineScope(scope.MachineScopeParams{
				Client:    fclient,
				Cluster:   tt.args.cluster,
				DOCluster: tt.args.docluster,
				Machine:   tt.args.machine,
				DOMachine: tt.args.domachine,
			})
			if err != nil {
				t.Fatalf("did not expect err: %v", err)
			}

			mdroplet := mock_computes.NewMockDropletsService(mctrl)
			mkey := mock_computes.NewMockKeysService(mctrl)
			mimage := mock_computes.NewMockImagesService(mctrl)
			mstorage := mock_computes.NewMockStorageService(mctrl)
			cscope, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Client:    fclient,
				Cluster:   tt.args.cluster,
				DOCluster: tt.args.docluster,
				DOClients: scope.DOClients{
					Droplets: mdroplet,
					Keys:     mkey,
					Images:   mimage,
					Storage:  mstorage,
				},
			})
			if err != nil {
				t.Fatalf("did not expect err: %v", err)
			}

			tt.expect(mdroplet.EXPECT(), mkey.EXPECT(), mimage.EXPECT(), mstorage.EXPECT())
			s := NewService(ctx, cscope)
			got, err := s.CreateDroplet(mscope)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.CreateDroplet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.CreateDroplet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_DeleteDroplet(t *testing.T) {
	os.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "super-secret-token")
	defer os.Unsetenv("DIGITALOCEAN_ACCESS_TOKEN")

	mctrl := gomock.NewController(t)
	defer mctrl.Finish()

	type args struct {
		id string
	}
	tests := []struct {
		name                string
		args                args
		expect              func(md *mock_computes.MockDropletsServiceMockRecorder)
		expectDeleteDroplet *gomock.Call
		wantErr             bool
	}{
		{
			name: "default",
			args: args{
				"12345",
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder) {
				md.Delete(gomock.Any(), 12345).Return(nil, nil)
			},
		},
		{
			name: "id is empty string (should return an error)",
			args: args{
				"",
			},
			expect:  func(md *mock_computes.MockDropletsServiceMockRecorder) {},
			wantErr: true,
		},
		{
			name: "id is not numeric (should return an error)",
			args: args{
				id: "aaa",
			},
			expect:  func(md *mock_computes.MockDropletsServiceMockRecorder) {},
			wantErr: true,
		},
		{
			name: "failed deleting droplet (should return an error)",
			args: args{
				"12345",
			},
			expect: func(md *mock_computes.MockDropletsServiceMockRecorder) {
				md.Delete(gomock.Any(), 12345).Return(nil, errors.New("error deleting droplet"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			mdroplet := mock_computes.NewMockDropletsService(mctrl)
			cscope, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Client:    fake.NewFakeClientWithScheme(scheme),
				Cluster:   &clusterv1.Cluster{},
				DOCluster: &infrav1.DOCluster{},
				DOClients: scope.DOClients{
					Droplets: mdroplet,
				},
			})
			if err != nil {
				t.Fatalf("did not expect err: %v", err)
			}

			tt.expect(mdroplet.EXPECT())
			s := NewService(ctx, cscope)
			if err := s.DeleteDroplet(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteDroplet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetDropletAddress(t *testing.T) {
	os.Setenv("DIGITALOCEAN_ACCESS_TOKEN", "super-secret-token")
	defer os.Unsetenv("DIGITALOCEAN_ACCESS_TOKEN")

	type args struct {
		droplet *godo.Droplet
	}
	tests := []struct {
		name    string
		args    args
		want    []corev1.NodeAddress
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				&godo.Droplet{
					ID:   1234,
					Name: "capdo-test",
					Networks: &godo.Networks{
						V4: []godo.NetworkV4{
							{
								Type:      "private",
								IPAddress: "10.0.0.1",
							},
							{
								Type:      "public",
								IPAddress: "192.168.1.1",
							},
						},
					},
				},
			},
			want: []corev1.NodeAddress{
				{
					Type:    corev1.NodeInternalIP,
					Address: "10.0.0.1",
				},
				{
					Type:    corev1.NodeExternalIP,
					Address: "192.168.1.1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			cscope, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Client:    fake.NewFakeClientWithScheme(scheme),
				Cluster:   &clusterv1.Cluster{},
				DOCluster: &infrav1.DOCluster{},
			})
			if err != nil {
				t.Fatalf("did not expect err: %v", err)
			}

			s := NewService(ctx, cscope)
			got, err := s.GetDropletAddress(tt.args.droplet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetDropletAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.GetDropletAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
