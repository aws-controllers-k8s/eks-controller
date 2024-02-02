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

package fargate_profile

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
)

var (
	UnableToUpdateError = "Changes to FargateProfile resources are not" +
		" currently possible. To update the resource, delete and re-create it"
)

var (
	requeueWaitWhileDeleting = ackrequeue.NeededAfter(
		fmt.Errorf("profile is in '%s' state, cannot be modified or deleted", svcsdk.FargateProfileStatusDeleting),
		ackrequeue.DefaultRequeueAfterDuration,
	)
)

// profileDeleting returns true if the supplied EKS FargateProfile is in the
// `Deleting` status
func profileDeleting(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	ps := *r.ko.Status.Status
	return ps == svcsdk.FargateProfileStatusDeleting
}

// customPreCompare ensures that default values of nil-able types are
// appropriately replaced with empty maps or structs depending on the default
// output of the SDK.
func customPreCompare(
	a *resource,
	b *resource,
) {
	if a.ko.Spec.Tags == nil && b.ko.Spec.Tags != nil {
		a.ko.Spec.Tags = map[string]*string{}
	} else if a.ko.Spec.Tags != nil && b.ko.Spec.Tags == nil {
		b.ko.Spec.Tags = map[string]*string{}
	}
}

func (rm *resourceManager) customUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdate")
	defer exit(err)

	if delta.DifferentAt("Spec.Tags") {
		if err := tags.SyncTags(
			ctx, rm.sdkapi, rm.metrics,
			string(*desired.ko.Status.ACKResourceMetadata.ARN),
			desired.ko.Spec.Tags, latest.ko.Spec.Tags,
		); err != nil {
			return nil, err
		}
		return desired, nil
	}

	// Never allow any changes
	updated = &resource{ko: desired.ko.DeepCopy()}
	ackcondition.SetSynced(updated, corev1.ConditionFalse, &UnableToUpdateError, nil)
	return updated, nil
}
