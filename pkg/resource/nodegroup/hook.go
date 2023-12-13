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
	"context"
	"fmt"
	"reflect"
	"time"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"
)

// Taken from the list of nodegroup statuses on the boto3 documentation
// https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/eks.html#EKS.Client.describe_nodegroup
const (
	StatusCreating     = "CREATING"
	StatusActive       = "ACTIVE"
	StatusDeleting     = "DELETING"
	StatusUpdating     = "UPDATING"
	StatusDegraded     = "DEGRADED"
	StatusCreateFailed = "CREATE_FAILED"
	StatusDeleteFailed = "DELETE_FAILED"
)

var (
	// TerminalStatuses are the status strings that are terminal states for a
	// cluster.
	TerminalStatuses = []string{
		StatusDeleting,
		StatusCreateFailed,
		StatusDeleteFailed,
	}
)

var (
	requeueWaitWhileDeleting = ackrequeue.NeededAfter(
		fmt.Errorf("nodegroup in '%s' state, cannot be modified or deleted", StatusDeleting),
		ackrequeue.DefaultRequeueAfterDuration,
	)
	RequeueAfterUpdateDuration = 15 * time.Second
)

// customPreCompare ensures that default values of nil-able types are
// appropriately replaced with empty maps or structs depending on the default
// output of the SDK.
func customPreCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	if a.ko.Spec.Labels == nil && b.ko.Spec.Labels != nil {
		a.ko.Spec.Labels = map[string]*string{}
	} else if a.ko.Spec.Labels != nil && b.ko.Spec.Labels == nil {
		b.ko.Spec.Labels = map[string]*string{}
	}
	if a.ko.Spec.Taints == nil && b.ko.Spec.Taints != nil {
		a.ko.Spec.Taints = make([]*svcapitypes.Taint, 0)
	} else if a.ko.Spec.Taints != nil && b.ko.Spec.Taints == nil {
		b.ko.Spec.Taints = make([]*svcapitypes.Taint, 0)
	}
	if a.ko.Spec.Tags == nil && b.ko.Spec.Tags != nil {
		a.ko.Spec.Tags = map[string]*string{}
	}
	if a.ko.Spec.Taints != nil && a.ko.Spec.Taints == nil || a.ko.Spec.Taints == nil && a.ko.Spec.Taints != nil {
		delta.Add("Spec.Taints", a.ko.Spec.Taints, b.ko.Spec.Taints)
	} else if a.ko.Spec.Taints != nil {
		if len(a.ko.Spec.Taints) != len(b.ko.Spec.Taints) {
			delta.Add("Spec.Taints", a.ko.Spec.Taints, b.ko.Spec.Taints)
		} else {
			for _, taintA := range a.ko.Spec.Taints {
				var matched = false
				for _, taintB := range b.ko.Spec.Taints {
					if reflect.DeepEqual(taintA, taintB) {
						matched = true
						break
					}
				}
				if !matched {
					delta.Add("Spec.Taints", a.ko.Spec.Taints, b.ko.Spec.Taints)
					break
				}
			}
		}
	}
}

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the nodegroup cannot be modified until it reaches an active
// status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.Status == nil {
		return nil
	}
	status := *r.ko.Status.Status
	return ackrequeue.NeededAfter(
		fmt.Errorf("nodegroup in '%s' state, cannot be modified until '%s'",
			status, StatusActive),
		ackrequeue.DefaultRequeueAfterDuration,
	)
}

// requeueAfterAsyncUpdate returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the nodegroup cannot be modified until after the asynchronous
// update has (first, started and then) completed and the nodegroup reaches an
// active status.
func requeueAfterAsyncUpdate() *ackrequeue.RequeueNeededAfter {

	return ackrequeue.NeededAfter(
		fmt.Errorf("nodegroup has started asynchronously updating, cannot be "+
			"modified until '%s'",
			StatusActive),
		RequeueAfterUpdateDuration,
	)
}

// nodegroupHasTerminalStatus returns whether the supplied cluster is in a
// terminal state
func nodegroupHasTerminalStatus(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	for _, s := range TerminalStatuses {
		if cs == s {
			return true
		}
	}
	return false
}

// nodegroupActive returns true if the supplied cluster is in an active status
func nodegroupActive(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusActive
}

// nodegroupCreating returns true if the supplied cluster is in the process of
// being created
func nodegroupCreating(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusCreating
}

// nodegroupDeleting returns true if the supplied cluster is in the process of
// being deleted
func nodegroupDeleting(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusDeleting
}

// returnNodegroupUpdating will set synced to false on the resource and
// return an async requeue error to signify that the resource should be
// forcefully requeued in order to pick up the 'UPDATING' status.
func returnNodegroupUpdating(r *resource) (*resource, error) {
	msg := "Nodegroup is currently being updated"
	ackcondition.SetSynced(r, corev1.ConditionFalse, &msg, nil)
	return r, requeueAfterAsyncUpdate()
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

	// For asynchronous updates, latest(from ReadOne) contains the
	// outdate values for Spec fields. However the status(Cluster status)
	// is correct inside latest.
	// So we construct the updatedRes object from the desired resource to
	// obtain correct spec fields and then copy the status from latest.
	updatedRes := rm.concreteResource(desired.DeepCopy())
	updatedRes.SetStatus(latest)
	if nodegroupDeleting(latest) {
		msg := "Nodegroup is currently being deleted"
		ackcondition.SetSynced(updatedRes, corev1.ConditionFalse, &msg, nil)
		return updatedRes, requeueWaitWhileDeleting
	}
	if !nodegroupActive(latest) {
		msg := "Nodegroup is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(updatedRes, corev1.ConditionFalse, &msg, nil)
		if nodegroupHasTerminalStatus(latest) {
			ackcondition.SetTerminal(updatedRes, corev1.ConditionTrue, &msg, nil)
			return updatedRes, nil
		}
		return updatedRes, requeueWaitUntilCanModify(latest)
	}

	if delta.DifferentAt("Spec.Labels") || delta.DifferentAt("Spec.Taints") ||
		delta.DifferentAt("Spec.ScalingConfig") || delta.DifferentAt("Spec.UpdateConfig") {
		if err := rm.updateConfig(ctx, desired, latest); err != nil {
			return nil, err
		}
		return returnNodegroupUpdating(updatedRes)
	}
	if delta.DifferentAt("Spec.Version") {
		if err := rm.updateVersion(ctx, desired); err != nil {
			return nil, err
		}
		return returnNodegroupUpdating(updatedRes)
	}

	rm.setStatusDefaults(updatedRes.ko)
	return updatedRes, nil
}

// newUpdateLabelsPayload determines which of the labels should be added or
// updated, and which labels should be removed, based on the desired vs the
// latest
func newUpdateLabelsPayload(
	desired *resource,
	latest *resource,
) *svcsdk.UpdateLabelsPayload {
	payload := svcsdk.UpdateLabelsPayload{
		AddOrUpdateLabels: desired.ko.Spec.Labels,
		RemoveLabels:      make([]*string, 0),
	}

	for latestKey := range latest.ko.Spec.Labels {
		if _, isDesired := desired.ko.Spec.Labels[latestKey]; !isDesired {
			toRemove := latestKey
			payload.RemoveLabels = append(payload.RemoveLabels, &toRemove)
		}
	}

	// Payload must have at least one update
	if len(payload.AddOrUpdateLabels) == 0 && len(payload.RemoveLabels) == 0 {
		return nil
	}

	return &payload
}

// newTaint creates a new AWS SDK Taint from a resource Taint value
func newTaint(
	t *svcapitypes.Taint,
) *svcsdk.Taint {
	r := &svcsdk.Taint{}
	if t.Effect != nil {
		r.Effect = t.Effect
	}
	if t.Key != nil {
		r.Key = t.Key
	}
	if t.Value != nil {
		r.Value = t.Value
	}
	return r
}

// newUpdateTaintsPayload determines which of the taints should be added or
// updated, and which taints should be removed, based on the desired vs the
// latest
func newUpdateTaintsPayload(
	desired *resource,
	latest *resource,
) *svcsdk.UpdateTaintsPayload {
	payload := svcsdk.UpdateTaintsPayload{
		AddOrUpdateTaints: make([]*svcsdk.Taint, len(desired.ko.Spec.Taints)),
		RemoveTaints:      make([]*svcsdk.Taint, 0),
	}

	// Add all the desired and existing taints
	for i, t := range desired.ko.Spec.Taints {
		payload.AddOrUpdateTaints[i] = newTaint(t)
	}

	// Check for existing taints that are not desired
	for _, inLatest := range latest.ko.Spec.Taints {
		exists := false
		for _, inDesired := range desired.ko.Spec.Taints {
			if *inDesired.Key != *inLatest.Key {
				continue
			}

			exists = true
			break
		}

		if !exists {
			payload.RemoveTaints = append(payload.RemoveTaints, newTaint(inLatest))
		}
	}

	// Payload must have at least one update
	if len(payload.AddOrUpdateTaints) == 0 && len(payload.RemoveTaints) == 0 {
		return nil
	}

	return &payload
}

func (rm *resourceManager) updateVersion(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateVersion")
	defer exit(err)
	input := &svcsdk.UpdateNodegroupVersionInput{
		NodegroupName: r.ko.Spec.Name,
		ClusterName:   r.ko.Spec.ClusterName,
		Version:       r.ko.Spec.Version,
	}

	_, err = rm.sdkapi.UpdateNodegroupVersionWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateNodegroupVersion", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfig(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfig")
	defer exit(err)

	input := &svcsdk.UpdateNodegroupConfigInput{
		NodegroupName: desired.ko.Spec.Name,
		ClusterName:   desired.ko.Spec.ClusterName,
		Labels:        newUpdateLabelsPayload(desired, latest),
		Taints:        newUpdateTaintsPayload(desired, latest),
	}

	if desired.ko.Spec.ScalingConfig != nil {
		input.SetScalingConfig(rm.newNodegroupScalingConfig(desired))
	}

	if desired.ko.Spec.UpdateConfig != nil {
		input.SetUpdateConfig(rm.newNodegroupUpdateConfig(desired))
	}

	_, err = rm.sdkapi.UpdateNodegroupConfigWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateNodegroupConfig", err)
	if err != nil {
		return err
	}

	return nil
}
