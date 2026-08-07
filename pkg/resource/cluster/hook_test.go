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

package cluster

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

func clusterWithDeletionProtection(name string, dp *bool) *resource {
	return &resource{
		ko: &svcapitypes.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "test",
			},
			Spec: svcapitypes.ClusterSpec{
				Name:               aws.String(name),
				DeletionProtection: dp,
			},
		},
	}
}

// TestUpdateDeletionProtectionNilIsNoOp is a regression test for a bug that
// prevented clusters from being deleted.
//
// DeletionProtection is configured with `pre_delete_include: true`, so on
// delete the runtime computes DeltaForPreDelete(desired, observed). When the
// user never set the field, desired is nil while EKS reports false. The
// generated delta flags that as a difference and the merged resource it
// returns carries the nil through to customUpdate, which then called
// UpdateClusterConfig with a nil DeletionProtection -- a request with no
// update type, rejected by EKS with:
//
//	InvalidParameterException: The type for cluster update was not provided.
//
// On the pre-delete path that error is fatal: the cluster is never deleted.
//
// updateDeletionProtection must therefore treat a nil value as a no-op rather
// than relying on the delta alone. The resourceManager here has a nil sdkapi,
// so the test fails loudly (panic) if the AWS call is ever reached.
func TestUpdateDeletionProtectionNilIsNoOp(t *testing.T) {
	rm := &resourceManager{}
	desired := clusterWithDeletionProtection("test-cluster", nil)

	require.NotPanics(t, func() {
		err := rm.updateDeletionProtection(context.Background(), desired)
		assert.NoError(t, err)
	}, "updateDeletionProtection must not call the EKS API for a nil DeletionProtection")
}

// TestDeletionProtectionDeltaFlagsNilDesired documents the generated-delta
// behaviour that makes the guard in updateDeletionProtection necessary. If
// ack-generate ever stops reporting nil-desired vs non-nil-observed as a
// difference, this test will fail and the guard can be revisited.
func TestDeletionProtectionDeltaFlagsNilDesired(t *testing.T) {
	desired := clusterWithDeletionProtection("test-cluster", nil)
	observed := clusterWithDeletionProtection("test-cluster", aws.Bool(false))

	preDeleteDelta, merged := newResourceDeltaForPreDelete(desired, observed)

	assert.True(
		t, preDeleteDelta.DifferentAt("Spec.DeletionProtection"),
		"pre-delete delta is expected to flag nil desired vs false observed",
	)
	assert.Nil(
		t, merged.ko.Spec.DeletionProtection,
		"merged pre-delete resource carries the nil desired value into customUpdate",
	)
}
