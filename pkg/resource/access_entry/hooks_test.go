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

package access_entry

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
)

func Test_computeAccessPoliciesDelta(t *testing.T) {
	type args struct {
		desired []*v1alpha1.AssociateAccessPolicyInput
		latest  []*v1alpha1.AssociateAccessPolicyInput
	}
	tests := []struct {
		name         string
		args         args
		wantToAdd    []*v1alpha1.AssociateAccessPolicyInput
		wantToDelete []*string
	}{
		{
			name: "nil arrays",
			args: args{
				desired: nil,
				latest:  nil,
			},
			wantToAdd:    nil,
			wantToDelete: nil,
		},
		{
			name: "empty/nil mixed arrays",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{},
				latest:  nil,
			},
			wantToAdd:    nil,
			wantToDelete: nil,
		},
		{
			name: "nil/empty mixed arrays",
			args: args{
				desired: nil,
				latest:  []*v1alpha1.AssociateAccessPolicyInput{},
			},
			wantToAdd:    nil,
			wantToDelete: nil,
		},
		{
			name: "empty arrays",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{},
				latest:  []*v1alpha1.AssociateAccessPolicyInput{},
			},
			wantToAdd:    nil,
			wantToDelete: nil,
		},
		{
			name: "different sizes - to add",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1")}},
				latest:  nil,
			},
			wantToAdd: []*v1alpha1.AssociateAccessPolicyInput{
				{PolicyARN: aws.String("policy-arn-1")},
			},
			wantToDelete: nil,
		},
		{
			name: "different sizes - to delete",
			args: args{
				desired: nil,
				latest:  []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1")}},
			},
			wantToDelete: []*string{aws.String("policy-arn-1")},
		},
		{
			name: "equal sizes - one element - not equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1")}},
				latest:  []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-2")}},
			},
			wantToAdd: []*v1alpha1.AssociateAccessPolicyInput{
				{PolicyARN: aws.String("policy-arn-1")},
			},
			wantToDelete: []*string{aws.String("policy-arn-2")},
		},
		{
			name: "equal sizes - one simple element - equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1")}},
				latest:  []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1")}},
			},
		},
		{
			name: "equal sizes - one full element - equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}}},
				latest:  []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}}},
			},
		},
		{
			name: "equal sizes - nil/empty elements - equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String(""), Namespaces: []*string{aws.String("ns-1")}}}},
				latest:  []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: nil, Namespaces: []*string{aws.String("ns-1")}}}},
			},
		},
		{
			name: "equal sizes - nil/empty elements - equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String(""), Namespaces: nil}}},
				latest:  []*v1alpha1.AssociateAccessPolicyInput{{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: nil, Namespaces: []*string{}}}},
			},
		},
		{
			name: "equal sizes - multiple elements - equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{
					{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-2"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: []*string{}}},
					{PolicyARN: aws.String("policy-arn-3"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-3"), Namespaces: nil}},
				},
				latest: []*v1alpha1.AssociateAccessPolicyInput{
					{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-2"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: nil}},
					{PolicyARN: aws.String("policy-arn-3"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-3"), Namespaces: []*string{}}},
				},
			},
		},
		{
			name: "complex case - multiple elements - not equal",
			args: args{
				desired: []*v1alpha1.AssociateAccessPolicyInput{
					// No-Op Policies
					{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-2"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: []*string{}}},
					{PolicyARN: aws.String("policy-arn-3"), AccessScope: &v1alpha1.AccessScope{Type: nil, Namespaces: nil}},
					{PolicyARN: aws.String("policy-arn-4"), AccessScope: &v1alpha1.AccessScope{Type: aws.String(""), Namespaces: nil}},
					// Policies to update (add/remove)
					{PolicyARN: aws.String("policy-arn-5"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-6"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: []*string{aws.String("ns-2")}}},
					{PolicyARN: aws.String("policy-arn-7"), AccessScope: &v1alpha1.AccessScope{Type: nil, Namespaces: nil}},
					{PolicyARN: aws.String("policy-arn-8"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-10"), Namespaces: nil}},
					// Policies to add (add only)
					{PolicyARN: aws.String("policy-arn-9"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
				},
				latest: []*v1alpha1.AssociateAccessPolicyInput{
					{PolicyARN: aws.String("policy-arn-1"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-2"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: nil}},
					{PolicyARN: aws.String("policy-arn-3"), AccessScope: &v1alpha1.AccessScope{Type: aws.String(""), Namespaces: []*string{}}},
					{PolicyARN: aws.String("policy-arn-4"), AccessScope: &v1alpha1.AccessScope{Type: aws.String(""), Namespaces: nil}},
					// Policies to update (add/remove)
					{PolicyARN: aws.String("policy-arn-5"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1"), aws.String("ns-2"), aws.String("ns-3")}}},
					{PolicyARN: aws.String("policy-arn-6"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-7"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-3"), Namespaces: []*string{aws.String("ns-1")}}},
					{PolicyARN: aws.String("policy-arn-8"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-11"), Namespaces: nil}},
					// Policies to be removed
					{PolicyARN: aws.String("policy-arn-10"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
				},
			},
			wantToAdd: []*v1alpha1.AssociateAccessPolicyInput{
				{PolicyARN: aws.String("policy-arn-5"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
				{PolicyARN: aws.String("policy-arn-6"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-2"), Namespaces: []*string{aws.String("ns-2")}}},
				{PolicyARN: aws.String("policy-arn-7"), AccessScope: &v1alpha1.AccessScope{Type: nil, Namespaces: nil}},
				{PolicyARN: aws.String("policy-arn-8"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-10"), Namespaces: nil}},
				{PolicyARN: aws.String("policy-arn-9"), AccessScope: &v1alpha1.AccessScope{Type: aws.String("type-1"), Namespaces: []*string{aws.String("ns-1")}}},
			},
			wantToDelete: []*string{
				aws.String("policy-arn-5"),
				aws.String("policy-arn-6"),
				aws.String("policy-arn-7"),
				aws.String("policy-arn-8"),
				aws.String("policy-arn-10"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToAdd, gotToDelete := computeAccessPoliciesDelta(tt.args.desired, tt.args.latest)
			if len(gotToAdd) != len(tt.wantToAdd) || (len(gotToAdd) > 0 && !reflect.DeepEqual(gotToAdd, tt.wantToAdd)) {
				t.Errorf("computeAccessPoliciesDelta() gotToAdd = %v, want %v", gotToAdd, tt.wantToAdd)
			}
			if len(gotToDelete) != len(tt.wantToDelete) || (len(gotToDelete) > 0 && !reflect.DeepEqual(gotToDelete, tt.wantToDelete)) {
				b, _ := json.MarshalIndent(gotToDelete, "", "    ")
				t.Log(string(b))
				b, _ = json.MarshalIndent(tt.wantToDelete, "", "    ")
				t.Log(string(b))
				t.Errorf("computeAccessPoliciesDelta() gotToDelete = %v, want %v", gotToDelete, tt.wantToDelete)
			}
		})
	}
}
