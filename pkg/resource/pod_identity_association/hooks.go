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

package pod_identity_association

import (
	"encoding/json"
	"reflect"

	awsiampolicy "github.com/micahhausler/aws-iam-policy/policy"

	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
)

var syncTags = tags.SyncTags

func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	comparePolicy(delta, a, b)
}

// use awsiampolicy to unmarshal Policy before we compare it to latest.
// this ensures unecessary diffs (whitespace/order change/etc.) does not trigger
// an update
func comparePolicy(delta *ackcompare.Delta, a *resource, b *resource) {
	if ackcompare.HasNilDifference(a.ko.Spec.Policy, b.ko.Spec.Policy) {
		delta.Add("Spec.Policy", a.ko.Spec.Policy, b.ko.Spec.Policy)
	} else if a.ko.Spec.Policy != nil && b.ko.Spec.Policy != nil {
		var policyDocumentA awsiampolicy.Policy
		var policyDocumentB awsiampolicy.Policy
		errA := json.Unmarshal([]byte(*a.ko.Spec.Policy), &policyDocumentA)
		errB := json.Unmarshal([]byte(*b.ko.Spec.Policy), &policyDocumentB)

		if errA != nil || errB != nil {
			if *a.ko.Spec.Policy != *b.ko.Spec.Policy {
				delta.Add("Spec.Policy", a.ko.Spec.Policy, b.ko.Spec.Policy)
			}
		} else if !reflect.DeepEqual(policyDocumentA, policyDocumentB) {
			delta.Add("Spec.Policy", a.ko.Spec.Policy, b.ko.Spec.Policy)
		}
	}
}
