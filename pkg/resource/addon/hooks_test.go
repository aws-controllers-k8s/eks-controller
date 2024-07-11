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
	"testing"

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
					{ptr("rolearn"), ptr("serviceaccount")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn"), ptr("serviceaccount")},
				},
			},
			true,
		},
		{
			"3 elements",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn"), ptr("serviceaccount")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn3"), ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn"), ptr("serviceaccount")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn3"), ptr("serviceaccount3")},
				},
			},
			true,
		},
		{
			"3 elements, different order",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn"), ptr("serviceaccount")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn3"), ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn3"), ptr("serviceaccount3")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn"), ptr("serviceaccount")},
				},
			},
			true,
		},
		{
			"different sizes",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn"), ptr("serviceaccount")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn3"), ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn3"), ptr("serviceaccount3")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
				},
			},
			false,
		},
		{
			"same size, different elements",
			args{
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn"), ptr("serviceaccount")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn3"), ptr("serviceaccount3")},
				},
				[]*v1alpha1.AddonPodIdentityAssociations{
					{ptr("rolearn3"), ptr("serviceaccount3")},
					{ptr("rolearn2"), ptr("serviceaccount2")},
					{ptr("rolearn4"), ptr("serviceaccount4")},
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
			if got := formatPodIdentityAssociation(&v1alpha1.AddonPodIdentityAssociations{tt.args.roleARN, tt.args.serviceAccount}); got != tt.want {
				t.Errorf("FormatPodIdentityAssociation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
