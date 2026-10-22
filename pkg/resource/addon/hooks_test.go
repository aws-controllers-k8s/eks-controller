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

func addonResource(version *string, configurationValues *string, tags map[string]*string) *resource {
	return &resource{ko: &v1alpha1.Addon{Spec: v1alpha1.AddonSpec{
		AddonVersion:        version,
		ConfigurationValues: configurationValues,
		Tags:                tags,
	}}}
}

// The generated delta still flags an unset spec version against the observed
// one, so the update path runs and can adopt the version instead of sending it.
func TestNewResourceDeltaFlagsUnsetAddonVersion(t *testing.T) {
	delta := newResourceDelta(addonResource(nil, nil, nil), addonResource(ptr("v1.13.2-eksbuild.3"), nil, nil))
	if !delta.DifferentAt("Spec.AddonVersion") {
		t.Fatalf("expected a Spec.AddonVersion difference, got %v", delta.Differences)
	}
}

func TestAdoptObservedAddonVersion(t *testing.T) {
	observed := ptr("v1.13.2-eksbuild.3")
	tests := []struct {
		name    string
		desired *resource
		latest  *resource
		adopted bool
	}{
		{"unset version is adopted", addonResource(nil, nil, nil), addonResource(observed, nil, nil), true},
		{"unset version with a tag change is adopted", addonResource(nil, nil, map[string]*string{"k": ptr("v")}), addonResource(observed, nil, nil), true},
		{"unset version with another change goes to the API", addonResource(nil, ptr("{}"), nil), addonResource(observed, ptr("{\"a\":1}"), nil), false},
		{"a set version that differs goes to the API", addonResource(ptr("v1.13.1-eksbuild.1"), nil, nil), addonResource(observed, nil, nil), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, tt.latest)
			got := adoptObservedAddonVersion(tt.desired, tt.latest, delta)
			if (got != nil) != tt.adopted {
				t.Fatalf("adoptObservedAddonVersion() = %v, want adopted=%v", got, tt.adopted)
			}
			if tt.desired.ko.Spec.AddonVersion != nil && tt.desired.ko.Spec.AddonVersion != tt.desired.ko.Spec.AddonVersion {
				t.Errorf("desired must not be rewritten")
			}
			if got != nil {
				if got.ko.Spec.AddonVersion == nil || *got.ko.Spec.AddonVersion != *observed {
					t.Errorf("adopted version = %v, want %s", got.ko.Spec.AddonVersion, *observed)
				}
				if tt.desired.ko.Spec.AddonVersion != nil {
					t.Errorf("desired was rewritten in place")
				}
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
