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
	"strconv"
	"strings"
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	"github.com/aws-controllers-k8s/eks-controller/pkg/util"
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
	// only compare releaseVersion if it's provided by the user. Note that there will always be a
	// ReleaseVersion in the observed state (after a successful creation).
	if a.ko.Spec.ReleaseVersion != nil && *a.ko.Spec.ReleaseVersion != "" {
		if *a.ko.Spec.ReleaseVersion != *b.ko.Spec.ReleaseVersion {
			delta.Add("Spec.ReleaseVersion", a.ko.Spec.ReleaseVersion, b.ko.Spec.ReleaseVersion)
		}
	}
	// only compare version if it's provided by the user. Note that there will always be a
	// Version in the observed state (after a successful creation).
	if a.ko.Spec.Version != nil && *a.ko.Spec.Version != "" {
		if *a.ko.Spec.Version != *b.ko.Spec.Version {
			delta.Add("Spec.Version", a.ko.Spec.Version, b.ko.Spec.Version)
		}
	}
}

func getDesiredSizeManagedByAnnotation(nodegroup *svcapitypes.Nodegroup) (string, bool) {
	if len(nodegroup.Annotations) == 0 {
		return "", false
	}
	managedBy, ok := nodegroup.Annotations[svcapitypes.DesiredSizeManagedByAnnotation]
	return managedBy, ok
}

func isManagedByExternalAutoscaler(nodegroup *svcapitypes.Nodegroup) bool {
	managedBy, ok := getDesiredSizeManagedByAnnotation(nodegroup)
	if !ok {
		return false
	}
	return managedBy == svcapitypes.DesiredSizeManagedByExternalAutoscaler
}

func customPostCompare(
	delta *ackcompare.Delta,
	a *resource,
	b *resource,
) {
	// We only want to compare the desiredSize field if and only if the
	// desiredSize is managed by the controller, meaning that in the case
	// where the desiredSize is managed by an external entity, we do not
	// want to compare the desiredSize field.
	// When managed by an external entity, an annotation is set on the
	// nodegroup resource to indicate that the desiredSize is managed
	// externally.
	if isManagedByExternalAutoscaler(a.ko) && delta.DifferentAt("Spec.ScalingConfig.DesiredSize") {
		// We need to unset the desiredSize field in the delta so that the
		// controller does not attempt to reconcile the desiredSize of the
		// nodegroup.
		newDiffs := make([]*ackcompare.Difference, 0)
		for _, d := range delta.Differences {
			if !d.Path.Contains("Spec.ScalingConfig.DesiredSize") {
				newDiffs = append(newDiffs, d)
			}
		}
		delta.Differences = newDiffs
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

	if immutableFieldChanges := rm.getImmutableFieldChanges(delta); len(immutableFieldChanges) > 0 {
		msg := fmt.Sprintf("Immutable Spec fields have been modified: %s", strings.Join(immutableFieldChanges, ","))
		return nil, ackerr.NewTerminalError(fmt.Errorf(msg))
	}

	if delta.DifferentAt("Spec.Tags") {
		err := tags.SyncTags(
			ctx, rm.sdkapi, rm.metrics,
			string(*latest.ko.Status.ACKResourceMetadata.ARN),
			desired.ko.Spec.Tags, latest.ko.Spec.Tags,
		)
		if err != nil {
			return nil, err
		}
	}
	if !delta.DifferentExcept("Spec.Tags") {
		return desired, nil
	}

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
		if err := rm.updateConfig(ctx, delta, desired, latest); err != nil {
			return nil, err
		}
		return returnNodegroupUpdating(updatedRes)
	}

	// At the stage we know that at least one of Version, ReleaseVersion or
	// LaunchTemplate has changed. The API does not allow using LaunchTemplate
	// with either Version or ReleaseVersion. So we need to check if the user
	// has provided a valid desired state.
	// There is no need to manually set a terminal condition here as the
	// controller will automatically set a terminal condition if the api
	// returns an InvalidParameterException

	if delta.DifferentAt("Spec.Version") || delta.DifferentAt("Spec.ReleaseVersion") || delta.DifferentAt("Spec.LaunchTemplate") {
		// Before trying to trigger a Nodegroup version update, we need to ensure that the user
		// has provided a valid desired state. For context the EKS UpdateNodegroupVersion API
		// accepts optional parameters Version and ReleaseVersion.
		//
		// The following are the valid combinations of the Version and ReleaseVersion parameters:
		// 1. None of the parameters are provided
		// 2. Only the Version parameter is provided
		// 3. Only the ReleaseVersion parameter is provided
		// 4. Both the Version and ReleaseVersion parameters are provided and they match
		//
		// The first case is not applicable here as it's counterintuitive in a declarative
		// model to not provide a desired state and have the controller trigger a blind update.

		// We need to set a terminal condition if the user provides both a version and release version
		// and they do not match. This is needed because the controller could potentially start alternating
		// between the non-matching version and release version in the spec and the observed state.
		if desired.ko.Spec.Version != nil && desired.ko.Spec.ReleaseVersion != nil &&
			*desired.ko.Spec.Version != "" && *desired.ko.Spec.ReleaseVersion != "" {

			// First parse the user provided release version and desired release
			desiredReleaseVersionTrimmed, err := util.GetEKSVersionFromReleaseVersion(*desired.ko.Spec.ReleaseVersion)
			if err != nil {
				return nil, ackerr.NewTerminalError(err)
			}

			// Set a terminal condition if the release version and version do not match.
			// e.g if the user provides a release version of 1.16.8-20211201 and a version of 1.17
			// They will either need to provide one of the following:
			// 2. A version
			// 1. A release version
			// 3. A version and release version that matches (e.g 1.16 and 1.16.8-20211201)
			if desiredReleaseVersionTrimmed != *desired.ko.Spec.Version {
				return nil, ackerr.NewTerminalError(
					fmt.Errorf("version and release version do not match: %s and %s", *desired.ko.Spec.Version, desiredReleaseVersionTrimmed),
				)
			}
		}

		if err := rm.updateVersion(ctx, delta, desired); err != nil {
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

func newUpdateNodegroupVersionPayload(
	delta *ackcompare.Delta,
	desired *resource,
) *svcsdk.UpdateNodegroupVersionInput {
	input := &svcsdk.UpdateNodegroupVersionInput{
		NodegroupName: desired.ko.Spec.Name,
		ClusterName:   desired.ko.Spec.ClusterName,
	}

	if delta.DifferentAt("Spec.Version") {
		input.Version = desired.ko.Spec.Version
	}

	if delta.DifferentAt("Spec.ReleaseVersion") {
		input.ReleaseVersion = desired.ko.Spec.ReleaseVersion
	}

	if delta.DifferentAt("Spec.LaunchTemplate") {
		// We need to be careful here to not access a nil pointer
		if desired.ko.Spec.LaunchTemplate != nil {
			input.SetLaunchTemplate(&svcsdk.LaunchTemplateSpecification{
				Id:      desired.ko.Spec.LaunchTemplate.ID,
				Name:    desired.ko.Spec.LaunchTemplate.Name,
				Version: desired.ko.Spec.LaunchTemplate.Version,
			})
		}
	}

	// If the force annotation is set, we set the force flag on the input
	// payload.
	if getUpdateNodeGroupForceAnnotation(desired.ko.ObjectMeta) {
		input.SetForce(true)
	}
	return input
}

func (rm *resourceManager) updateVersion(
	ctx context.Context,
	delta *ackcompare.Delta,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateVersion")
	defer exit(err)

	input := newUpdateNodegroupVersionPayload(delta, r)

	_, err = rm.sdkapi.UpdateNodegroupVersionWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateNodegroupVersion", err)
	if err != nil {
		return err
	}

	return nil
}

const (
	defaultUpgradeNodeGroupVersion = false
)

// GetDeleteForce returns whether the nodegroup should be deleted forcefully.
//
// https://docs.aws.amazon.com/eks/latest/APIReference/API_UpdateNodegroupVersion.html
func getUpdateNodeGroupForceAnnotation(
	m metav1.ObjectMeta,
) bool {
	resAnnotations := m.GetAnnotations()
	forceVersionUpdate, ok := resAnnotations[svcapitypes.ForceNodeGroupUpdateVersionAnnotation]
	if !ok {
		return defaultUpgradeNodeGroupVersion
	}

	forceVersionUpdateBool, err := strconv.ParseBool(forceVersionUpdate)
	if err != nil {
		return defaultUpgradeNodeGroupVersion
	}

	return forceVersionUpdateBool
}
func (rm *resourceManager) newUpdateScalingConfigPayload(
	desired, latest *resource,
) *svcsdk.NodegroupScalingConfig {
	sc := rm.newNodegroupScalingConfig(desired)
	// We need to default the desiredSize to the current observed
	// value in the case where the desiredSize is managed externally.
	isManagedExternally := isManagedByExternalAutoscaler(desired.ko)
	if isManagedExternally {
		rm.log.Info(
			"detected that the desiredSize is managed by an external entity.",
			"annotation", fmt.Sprintf("%s: '%s'", svcapitypes.DesiredSizeManagedByAnnotation, svcapitypes.DesiredSizeManagedByExternalAutoscaler),
		)
	}
	if isManagedExternally && latest.ko.Spec.ScalingConfig != nil {
		rm.log.Info(
			"ignoring the difference in desiredSize as it is managed by an external entity.",
			"external_desired_size", latest.ko.Spec.ScalingConfig.DesiredSize,
			"ack_desired_size", desired.ko.Spec.ScalingConfig.DesiredSize,
		)
		sc.DesiredSize = latest.ko.Spec.ScalingConfig.DesiredSize
	}
	return sc
}

func (rm *resourceManager) updateConfig(
	ctx context.Context,
	delta *ackcompare.Delta,
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
		input.SetScalingConfig(rm.newUpdateScalingConfigPayload(desired, latest))
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
