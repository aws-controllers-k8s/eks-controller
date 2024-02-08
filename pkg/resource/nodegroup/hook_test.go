// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package nodegroup

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

func TestTaints(t *testing.T) {

	noSchedule := "NO_SCHEDULE"
	owner := "owner"
	project := "project"

	teamOne := "teamone"
	projectOne := "projectone"

	a := &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectOne,
					},
				},
			},
		},
	}

	b := &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectOne,
					},
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
				},
			},
		},
	}

	delta := &ackcompare.Delta{}
	customPreCompare(delta, a, b)
	assert.False(t, delta.DifferentAt("Spec.Taints"), "Taints are equals")

	projectTwo := "projecttwo"

	b = &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectTwo,
					},
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
				},
			},
		},
	}

	delta = &ackcompare.Delta{}
	customPreCompare(delta, a, b)
	assert.True(t, delta.DifferentAt("Spec.Taints"), "Taints are different")

	other := "other"

	b = &resource{
		ko: &v1alpha1.Nodegroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "test",
			},
			Spec: v1alpha1.NodegroupSpec{
				Taints: []*v1alpha1.Taint{
					{
						Effect: &noSchedule,
						Key:    &project,
						Value:  &projectOne,
					},
					{
						Effect: &noSchedule,
						Key:    &owner,
						Value:  &teamOne,
					},
					{
						Effect: &noSchedule,
						Key:    &other,
						Value:  &other,
					},
				},
			},
		},
	}

	delta = &ackcompare.Delta{}
	customPreCompare(delta, a, b)
	assert.True(t, delta.DifferentAt("Spec.Taints"), "Taints have different length")
}

func newScalingConfig(desired, max, min int64) *v1alpha1.NodegroupScalingConfig {
	return &v1alpha1.NodegroupScalingConfig{
		DesiredSize: aws.Int64(desired),
		MaxSize:     aws.Int64(max),
		MinSize:     aws.Int64(min),
	}
}

func newNodegroupWithScalingConfig(desired, max, min int64) *v1alpha1.Nodegroup {
	return &v1alpha1.Nodegroup{
		Spec: v1alpha1.NodegroupSpec{
			ScalingConfig: newScalingConfig(desired, max, min),
		},
	}
}

func newNodegroupScalingConfigManagedByExternalAutoscaler(desired, max, min int64) *v1alpha1.Nodegroup {
	nodegroup := newNodegroupWithScalingConfig(desired, max, min)
	nodegroup.ObjectMeta.SetAnnotations(map[string]string{
		"eks.services.k8s.aws/desired-size-managed-by": "external-autoscaler",
	})
	return nodegroup
}

func Test_compareNodegroupScalingConfigs_ManagedByDefault(t *testing.T) {
	type args struct {
		a *v1alpha1.Nodegroup
		b *v1alpha1.Nodegroup
	}
	tests := []struct {
		name                       string
		args                       args
		expectDelta                bool
		expectDeltaAtScalingConfig bool
	}{
		{
			name: "non nil empty scaling config",
			args: args{
				a: newNodegroupWithScalingConfig(0, 0, 0),
				b: newNodegroupWithScalingConfig(0, 0, 0),
			},
			expectDelta:                false,
			expectDeltaAtScalingConfig: false,
		},
		{
			name: "different scaling configurations",
			args: args{
				a: newNodegroupWithScalingConfig(2, 2, 2),
				b: newNodegroupWithScalingConfig(1, 1, 1),
			},
			expectDelta:                true,
			expectDeltaAtScalingConfig: true,
		},
		{
			name: "different desired size",
			args: args{
				a: newNodegroupWithScalingConfig(2, 2, 2),
				b: newNodegroupWithScalingConfig(1, 2, 2),
			},
			expectDelta:                true,
			expectDeltaAtScalingConfig: true,
		},
		{
			name: "different max/min size",
			args: args{
				a: newNodegroupWithScalingConfig(2, 2, 2),
				b: newNodegroupWithScalingConfig(2, 1, 5),
			},
			expectDelta:                true,
			expectDeltaAtScalingConfig: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(&resource{tt.args.a}, &resource{tt.args.b})
			if tt.expectDelta != (len(delta.Differences) > 0) {
				t.Errorf("customPostCompare() has delta = %v, want %v", delta, tt.expectDelta)
			}
			if tt.expectDeltaAtScalingConfig != delta.DifferentAt("Spec.ScalingConfig.DesiredSize") {
				for _, diff := range delta.Differences {
					fmt.Println("**diff:", diff.Path)
				}
				t.Errorf("customPostCompare() has delta at ScalingConfig = %v, want %v", delta, tt.expectDeltaAtScalingConfig)
			}
		})
	}
}

func Test_compareNodegroupScalingConfigs_ManagedByExternal(t *testing.T) {
	type args struct {
		a *v1alpha1.Nodegroup
		b *v1alpha1.Nodegroup
	}
	tests := []struct {
		name                       string
		args                       args
		expectDelta                bool
		expectDeltaAtScalingConfig bool
	}{
		{
			name: "non nil empty scaling config",
			args: args{
				a: newNodegroupScalingConfigManagedByExternalAutoscaler(0, 0, 0),
				b: newNodegroupScalingConfigManagedByExternalAutoscaler(0, 0, 0),
			},
			expectDelta:                false,
			expectDeltaAtScalingConfig: false,
		},
		{
			name: "different scaling configurations",
			args: args{
				a: newNodegroupScalingConfigManagedByExternalAutoscaler(2, 2, 2),
				b: newNodegroupScalingConfigManagedByExternalAutoscaler(1, 1, 1),
			},
			expectDelta:                true,
			expectDeltaAtScalingConfig: false,
		},
		{
			name: "different desired size",
			args: args{
				a: newNodegroupScalingConfigManagedByExternalAutoscaler(2, 2, 2),
				b: newNodegroupScalingConfigManagedByExternalAutoscaler(1, 2, 2),
			},
			expectDelta:                false,
			expectDeltaAtScalingConfig: false,
		},
		{
			name: "different max/min size",
			args: args{
				a: newNodegroupWithScalingConfig(2, 2, 2),
				b: newNodegroupWithScalingConfig(2, 1, 5),
			},
			expectDelta:                true,
			expectDeltaAtScalingConfig: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(&resource{tt.args.a}, &resource{tt.args.b})
			if tt.expectDelta != (len(delta.Differences) > 0) {
				t.Errorf("customPostCompare() has delta = %v, want %v", delta, tt.expectDelta)
			}
			if tt.expectDeltaAtScalingConfig != delta.DifferentAt("Spec.ScalingConfig.DesiredSize") {
				t.Errorf("customPostCompare() has delta at ScalingConfig = %v, want %v", delta, tt.expectDeltaAtScalingConfig)
			}
		})
	}
}

func newUpdateScalingConfigPayload(desired, max, min int64) *svcsdk.NodegroupScalingConfig {
	return &svcsdk.NodegroupScalingConfig{
		DesiredSize: aws.Int64(desired),
		MaxSize:     aws.Int64(max),
		MinSize:     aws.Int64(min),
	}
}

func Test_resourceManager_newUpdateScalingConfigPayload_ManagedByExternalAutoscaler(t *testing.T) {
	type args struct {
		latest  *v1alpha1.Nodegroup
		desired *v1alpha1.Nodegroup
	}
	tests := []struct {
		name string
		args args
		want *svcsdk.NodegroupScalingConfig
	}{
		{
			name: "no changes",
			args: args{
				latest:  newNodegroupScalingConfigManagedByExternalAutoscaler(2, 2, 2),
				desired: newNodegroupScalingConfigManagedByExternalAutoscaler(2, 2, 2),
			},
			// In such a situation the request will never be sent to EKS, however we still wanna test the behaviour
			// of the function.
			want: newUpdateScalingConfigPayload(2, 2, 2),
		},
		{
			name: "only desired size changed",
			args: args{
				desired: newNodegroupScalingConfigManagedByExternalAutoscaler(10, 2, 2),
				latest:  newNodegroupScalingConfigManagedByExternalAutoscaler(2, 2, 2),
			},
			want: newUpdateScalingConfigPayload(2, 2, 2),
		},
		{
			name: "all the fields changed",
			args: args{
				desired: newNodegroupScalingConfigManagedByExternalAutoscaler(20, 15, 20),
				latest:  newNodegroupScalingConfigManagedByExternalAutoscaler(10, 10, 10),
			},
			want: newUpdateScalingConfigPayload(10, 15, 20),
		},
		{
			name: "all the fields changed except desired size",
			args: args{
				desired: newNodegroupScalingConfigManagedByExternalAutoscaler(10, 15, 20),
				latest:  newNodegroupScalingConfigManagedByExternalAutoscaler(10, 10, 10),
			},
			want: newUpdateScalingConfigPayload(10, 15, 20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := resourceManager{
				log: zap.New(zap.UseFlagOptions(&zap.Options{
					DestWriter: io.Discard,
				})),
			}
			if got := rm.newUpdateScalingConfigPayload(&resource{tt.args.desired}, &resource{tt.args.latest}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resourceManager.newUpdateScalingConfigPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resourceManager_newUpdateScalingConfigPayload_ManagedByDefault(t *testing.T) {
	type args struct {
		latest  *v1alpha1.Nodegroup
		desired *v1alpha1.Nodegroup
	}
	tests := []struct {
		name string
		args args
		want *svcsdk.NodegroupScalingConfig
	}{
		{
			name: "no changes",
			args: args{
				latest:  newNodegroupWithScalingConfig(2, 2, 2),
				desired: newNodegroupWithScalingConfig(2, 2, 2),
			},
			// In such a situation the request will never be sent to EKS, however we still wanna test the behaviour
			// of the function.
			want: newUpdateScalingConfigPayload(2, 2, 2),
		},
		{
			name: "only desired size changed",
			args: args{
				desired: newNodegroupWithScalingConfig(10, 2, 2),
				latest:  newNodegroupWithScalingConfig(2, 2, 2),
			},
			want: newUpdateScalingConfigPayload(10, 2, 2),
		},
		{
			name: "all the fields changed",
			args: args{
				desired: newNodegroupWithScalingConfig(20, 15, 20),
				latest:  newNodegroupWithScalingConfig(10, 10, 10),
			},
			want: newUpdateScalingConfigPayload(20, 15, 20),
		},
		{
			name: "all the fields changed except desired size",
			args: args{
				desired: newNodegroupWithScalingConfig(10, 15, 20),
				latest:  newNodegroupWithScalingConfig(10, 10, 10),
			},
			want: newUpdateScalingConfigPayload(10, 15, 20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := resourceManager{}
			if got := rm.newUpdateScalingConfigPayload(&resource{tt.args.desired}, &resource{tt.args.latest}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resourceManager.newUpdateScalingConfigPayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newUpdateNodegroupPayload(t *testing.T) {
	delta := ackcompare.NewDelta()
	delta.Add("Spec.Version", nil, nil)
	delta.Add("Spec.LaunchTemplate", nil, nil)

	type args struct {
		r *resource
	}
	tests := []struct {
		name               string
		args               args
		wantVersion        string
		wantForce          bool
		wantLaunchTemplate bool
	}{
		{
			name: "version only",
			args: args{
				r: &resource{
					ko: &v1alpha1.Nodegroup{
						Spec: v1alpha1.NodegroupSpec{
							Version: aws.String("1.21"),
						},
					},
				},
			},
			wantVersion:        "1.21",
			wantForce:          false,
			wantLaunchTemplate: false,
		},
		{
			name: "all fields",
			args: args{
				r: &resource{
					ko: &v1alpha1.Nodegroup{
						Spec: v1alpha1.NodegroupSpec{
							Version:        aws.String("1.21"),
							ReleaseVersion: aws.String("someversion"),
							LaunchTemplate: &v1alpha1.LaunchTemplateSpecification{
								ID: aws.String("id"),
							},
						},
					},
				},
			},
			wantVersion:        "1.21",
			wantForce:          false,
			wantLaunchTemplate: true,
		},
		{
			name: "force update annotation false",
			args: args{
				r: &resource{
					ko: &v1alpha1.Nodegroup{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"eks.services.k8s.aws/force-update-version": "false",
							},
						},
						Spec: v1alpha1.NodegroupSpec{
							Version: aws.String("1.21"),
						},
					},
				},
			},
			wantVersion:        "1.21",
			wantForce:          false,
			wantLaunchTemplate: false,
		},
		{
			name: "force update annotation true",
			args: args{
				r: &resource{
					ko: &v1alpha1.Nodegroup{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"eks.services.k8s.aws/force-update-version": "true",
							},
						},
						Spec: v1alpha1.NodegroupSpec{
							Version: aws.String("1.21"),
						},
					},
				},
			},
			wantVersion:        "1.21",
			wantForce:          true,
			wantLaunchTemplate: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newUpdateNodegroupVersionPayload(delta, tt.args.r)
			assert.Equal(t, tt.wantVersion, *got.Version)
			if tt.wantForce {
				assert.NotNil(t, got.Force)
				assert.True(t, *got.Force)
			} else {
				assert.Nil(t, got.Force)
			}
			if tt.wantLaunchTemplate {
				assert.NotNil(t, got.LaunchTemplate)
			} else {
				assert.Nil(t, got.LaunchTemplate)
			}
		})
	}
}
