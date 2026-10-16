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

package addon

import (
	"errors"
	"testing"
	"time"

	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

func TestEqualPodIdentityAssociation(t *testing.T) {
	type args struct {
		a []*v1alpha1.AddonPodIdentityAssociations
		b []*v1alpha1.AddonPodIdentityAssociations
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"empty slices",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{},
				[]*v1alpha1.AddonPodIdentityAssociations{},
			},
			true,
		},
		{
			"non-empty slices",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
				},
			},
			true,
		},
		{
			"3 elements",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
				},
			},
			true,
		},
		{
			"3 elements, different order",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
				},
			},
			true,
		},
		{
			"different sizes",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
				},
			},
			false,
		},
		{
			"same size, different elements",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn"), ServiceAccount: ptr("serviceaccount")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{RoleARN: ptr("rolearn3"), ServiceAccount: ptr("serviceaccount3")},
					{RoleARN: ptr("rolearn2"), ServiceAccount: ptr("serviceaccount2")},
					{RoleARN: ptr("rolearn4"), ServiceAccount: ptr("serviceaccount4")},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equalPodIdentityAssociations(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("EqualPodIdentityAssociation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatPodIdentityAssociation(t *testing.T) {
	type args struct {
		roleARN        *string
		serviceAccount *string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"non-nil pointers",
			args{
				ptr("rolearn"),
				ptr("serviceaccount"),
			},
			"serviceaccount/rolearn",
		},
		{
			"nil pointers",
			args{
				nil,
				ptr("serviceaccount"),
			},
			"serviceaccount/",
		},
		{
			"nil pointers",
			args{
				ptr("rolearn"),
				nil,
			},
			"/rolearn",
		},
		{
			"nil pointers",
			args{
				nil,
				nil,
			},
			"/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatPodIdentityAssociation(&v1alpha1.AddonPodIdentityAssociations{RoleARN: tt.args.roleARN, ServiceAccount: tt.args.serviceAccount}); got != tt.want {
				t.Errorf("FormatPodIdentityAssociation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestNewResourceDeltaAddonVersion(t *testing.T) {
	tests := []struct {
		name          string
		desired       *string
		latest        *string
		wantDifferent bool
	}{
		{"unset spec keeps the observed version", nil, ptr("v1.13.2-eksbuild.3"), false},
		{"equal versions", ptr("v1.13.2-eksbuild.3"), ptr("v1.13.2-eksbuild.3"), false},
		{"different versions", ptr("v1.13.2-eksbuild.3"), ptr("v1.13.1-eksbuild.1"), true},
		{"nothing observed", ptr("v1.13.2-eksbuild.3"), nil, true},
		{"both unset", nil, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired := &resource{ko: &v1alpha1.Addon{Spec: v1alpha1.AddonSpec{AddonVersion: tt.desired}}}
			latest := &resource{ko: &v1alpha1.Addon{Spec: v1alpha1.AddonSpec{AddonVersion: tt.latest}}}
			delta := newResourceDelta(desired, latest)
			if got := delta.DifferentAt("Spec.AddonVersion"); got != tt.wantDifferent {
				t.Errorf("DifferentAt(Spec.AddonVersion) = %v, want %v", got, tt.wantDifferent)
			}
			if desired.ko.Spec.AddonVersion != tt.desired || latest.ko.Spec.AddonVersion != tt.latest {
				t.Errorf("the compare hook must not rewrite desired or latest")
			}
		})
	}
}

func TestReturnAddonUpdatingRequeuesAfterSeconds(t *testing.T) {
	_, err := returnAddonUpdating(&resource{ko: &v1alpha1.Addon{}})
	var requeueErr *ackrequeue.RequeueNeededAfter
	if !errors.As(err, &requeueErr) {
		t.Fatalf("returnAddonUpdating() error = %T, want *ackrequeue.RequeueNeededAfter", err)
	}
	if got := requeueErr.Duration(); got != 15*time.Second {
		t.Errorf("requeue duration = %v, want %v", got, 15*time.Second)
	}
}
