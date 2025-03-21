/*
Copyright 2023 The Kubernetes Authors.

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

package networking

import (
	"context"
	"errors"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/digitalocean/godo"
	"go.uber.org/mock/gomock"

	infrav1 "sigs.k8s.io/cluster-api-provider-digitalocean/api/v1beta1"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-digitalocean/cloud/services/networking/mock_networking"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(infrav1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
}

func TestService_GetLoadBalancer(t *testing.T) {
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
		expect  func(mlb *mock_networking.MockLoadBalancersServiceMockRecorder)
		want    *godo.LoadBalancer
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				id: "123456",
			},
			expect: func(mlb *mock_networking.MockLoadBalancersServiceMockRecorder) {
				mlb.Get(gomock.Any(), "123456").Return(&godo.LoadBalancer{
					ID:   "123456",
					Name: "test-lb",
				}, nil, nil)
			},
			want: &godo.LoadBalancer{
				ID:   "123456",
				Name: "test-lb",
			},
		},
		{
			name: "id is empty string (should not return an error)",
			args: args{
				id: "",
			},
			expect:  func(_ *mock_networking.MockLoadBalancersServiceMockRecorder) {},
			wantErr: false,
		},
		{
			name: "loadbalancer not found (should not return an error)",
			args: args{
				id: "123456",
			},
			expect: func(mlb *mock_networking.MockLoadBalancersServiceMockRecorder) {
				mlb.Get(gomock.Any(), "123456").Return(&godo.LoadBalancer{}, &godo.Response{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
					},
				}, errors.New("droplet not found"))
			},
			wantErr: false,
		},
		{
			name: "godo return unknown error",
			args: args{
				id: "123456",
			},
			expect: func(mlb *mock_networking.MockLoadBalancersServiceMockRecorder) {
				mlb.Get(gomock.Any(), "123456").Return(&godo.LoadBalancer{}, &godo.Response{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
					},
				}, errors.New("unexpected error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			mlbalancer := mock_networking.NewMockLoadBalancersService(mctrl)
			cscope, err := scope.NewClusterScope(scope.ClusterScopeParams{
				Client:    fake.NewClientBuilder().WithScheme(scheme).Build(),
				Cluster:   &clusterv1.Cluster{},
				DOCluster: &infrav1.DOCluster{},
				DOClients: scope.DOClients{
					LoadBalancers: mlbalancer,
				},
			})
			if err != nil {
				t.Fatalf("did not expect err: %v", err)
			}

			tt.expect(mlbalancer.EXPECT())
			s := NewService(ctx, cscope)
			got, err := s.GetLoadBalancer(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Service.GetLoadBalancer() = %v, want %v", got, tt.want)
			}
		})
	}
}
