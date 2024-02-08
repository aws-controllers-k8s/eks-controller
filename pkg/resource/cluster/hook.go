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
	"errors"
	"fmt"
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	"github.com/aws-controllers-k8s/eks-controller/pkg/util"
)

const (
	LoggingNoChangesError = "No changes needed for the logging config provided"
)

// Taken from the list of cluster statuses on the boto3 documentation
// https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/eks.html#EKS.Client.describe_cluster
const (
	StatusCreating = "CREATING"
	StatusActive   = "ACTIVE"
	StatusDeleting = "DELETING"
	StatusFailed   = "FAILED"
	StatusUpdating = "UPDATING"
	StatusPending  = "PENDING"
)

var (
	// TerminalStatuses are the status strings that are terminal states for a
	// cluster.
	TerminalStatuses = []string{
		StatusDeleting,
		StatusFailed,
	}
)

var (
	requeueWaitWhileDeleting = ackrequeue.NeededAfter(
		fmt.Errorf("cluster in '%s' state, cannot be modified or deleted", StatusDeleting),
		ackrequeue.DefaultRequeueAfterDuration,
	)
	requeueWaitWhileInUse = ackrequeue.NeededAfter(
		errors.New("cluster is still in use, cannot be deleted"),
		ackrequeue.DefaultRequeueAfterDuration,
	)
	RequeueAfterUpdateDuration = 15 * time.Second
)

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the cluster cannot be modified until it reaches an active status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.Status == nil {
		return nil
	}
	status := *r.ko.Status.Status
	return ackrequeue.NeededAfter(
		fmt.Errorf("cluster in '%s' state, cannot be modified until '%s'",
			status, StatusActive),
		ackrequeue.DefaultRequeueAfterDuration,
	)
}

// requeueAfterAsyncUpdate returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the cluster cannot be modified until after the asynchronous update
// has (first, started and then) completed and the cluster reaches an active
// status.
func requeueAfterAsyncUpdate() *ackrequeue.RequeueNeededAfter {
	return ackrequeue.NeededAfter(
		fmt.Errorf("cluster has started asynchronously updating, cannot be modified until '%s'",
			StatusActive),
		RequeueAfterUpdateDuration,
	)
}

// clusterHasTerminalStatus returns whether the supplied cluster is in a
// terminal state
func clusterHasTerminalStatus(r *resource) bool {
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

// clusterActive returns true if the supplied cluster is in an active status
func clusterActive(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusActive
}

// clusterCreating returns true if the supplied cluster is in the process of
// being created
func clusterCreating(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusCreating
}

// clusterDeleting returns true if the supplied cluster is in the process of
// being deleted
func clusterDeleting(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusDeleting
}

// returnClusterUpdating will set synced to false on the resource and
// return an async requeue error to signify that the resource should be
// forcefully requeued in order to pick up the 'UPDATING' status.
func returnClusterUpdating(r *resource) (*resource, error) {
	msg := "Cluster is currently being updated"
	ackcondition.SetSynced(r, corev1.ConditionFalse, &msg, nil)
	return r, requeueAfterAsyncUpdate()
}

// clusterInUse returns true if the supplied cluster is still being used. It
// determines this by checking if there are any nodegroups still attached to
// the cluster.
func (rm *resourceManager) clusterInUse(ctx context.Context, r *resource) (bool, error) {
	nodes, err := rm.listNodegroups(ctx, r)
	if err != nil {
		return false, err
	}

	// Cluster is in use if # of nodegroups != 0
	return (nodes != nil && len(nodes.Nodegroups) > 0), nil
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
	if clusterDeleting(latest) {
		msg := "Cluster is currently being deleted"
		ackcondition.SetSynced(updatedRes, corev1.ConditionFalse, &msg, nil)
		return updatedRes, requeueWaitWhileDeleting
	}
	if !clusterActive(latest) {
		msg := "Cluster is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(updatedRes, corev1.ConditionFalse, &msg, nil)
		if clusterHasTerminalStatus(latest) {
			ackcondition.SetTerminal(updatedRes, corev1.ConditionTrue, &msg, nil)
			return updatedRes, nil
		}
		return updatedRes, requeueWaitUntilCanModify(latest)
	}

	if delta.DifferentAt("Spec.Tags") {
		if err := tags.SyncTags(
			ctx, rm.sdkapi, rm.metrics,
			string(*desired.ko.Status.ACKResourceMetadata.ARN),
			desired.ko.Spec.Tags, latest.ko.Spec.Tags,
		); err != nil {
			return nil, err
		}
		// Tag updates are not reflected in the status, so we don't need to
		// requeue here.
	}
	if delta.DifferentAt("Spec.Logging") {
		if err := rm.updateConfigLogging(ctx, desired); err != nil {
			awserr, ok := ackerr.AWSError(err)

			// The API responds with an error if there were no changes applied
			if !ok || awserr.Message() != LoggingNoChangesError {
				return nil, err
			}

			// Check to see if we've raced an async update call and need to
			// requeue
			if ok && awserr.Code() == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}
		}
		return returnClusterUpdating(updatedRes)
	}
	if delta.DifferentAt("Spec.ResourcesVPCConfig") {
		if err := rm.updateConfigResourcesVPCConfig(ctx, desired); err != nil {
			awserr, ok := ackerr.AWSError(err)

			// Check to see if we've raced an async update call and need to
			// requeue
			if ok && awserr.Code() == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}
	if delta.DifferentAt("Spec.AccessConfig") {
		if err := rm.updateAccessConfig(ctx, desired); err != nil {
			awserr, ok := ackerr.AWSError(err)

			// Check to see if we've raced an async update call and need to
			// requeue
			if ok && awserr.Code() == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}
			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	if delta.DifferentAt("Spec.Version") {
		if err := rm.updateVersion(ctx, desired, latest); err != nil {
			awserr, ok := ackerr.AWSError(err)

			// Check to see if we've raced an async update call and need to
			// requeue
			if ok && awserr.Code() == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	rm.setStatusDefaults(updatedRes.ko)
	return updatedRes, nil
}

// updateVersion updates the cluster version to the next possible version.
//
// This function isn't supposed to blindly update the cluster version to the
// desired version. It should increment the minor version of the observed
// version and update the cluster to that version. The reconciliation mechanism
// should ensure that the desired version is eventually reached.
func (rm *resourceManager) updateVersion(
	ctx context.Context,
	desired, latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateVersion")
	defer exit(err)

	// If the desired version is less than the observed version, we can't update
	// the cluster to an older version.
	// Note that the desired and observed versions are guaranteed to be never be
	// equal at this stage, as the delta comparison would have caught that.
	compareResult, err := util.CompareEKSKubernetesVersions(*desired.ko.Spec.Version, *latest.ko.Spec.Version)
	if err != nil {
		return ackerr.NewTerminalError(fmt.Errorf("failed to compare the desired and observed versions: %v", err))
	}
	if compareResult != 1 {
		return ackerr.NewTerminalError(
			fmt.Errorf("desired cluster version is less than the observed version: %s < %s",
				*desired.ko.Spec.Version, *latest.ko.Spec.Version,
			),
		)
	}

	// Compure the next minor version of the desired version
	nextVersion, err := util.IncrementEKSMinorVersion(*latest.ko.Spec.Version)
	if err != nil {
		return ackerr.NewTerminalError(fmt.Errorf("failed to compute the next minor version: %v", err))
	}

	input := &svcsdk.UpdateClusterVersionInput{
		Name:    desired.ko.Spec.Name,
		Version: &nextVersion,
	}

	_, err = rm.sdkapi.UpdateClusterVersionWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterVersion", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfigLogging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfigLogging")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:    r.ko.Spec.Name,
		Logging: rm.newLogging(r),
	}

	_, err = rm.sdkapi.UpdateClusterConfigWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func newAccessConfig(r *resource) *svcsdk.UpdateAccessConfigRequest {
	cfg := &svcsdk.UpdateAccessConfigRequest{}
	if r.ko.Spec.AccessConfig != nil {
		cfg.AuthenticationMode = r.ko.Spec.AccessConfig.AuthenticationMode
	}
	return cfg
}

func (rm *resourceManager) updateAccessConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateAccessConfig")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:         r.ko.Spec.Name,
		AccessConfig: newAccessConfig(r),
	}

	_, err = rm.sdkapi.UpdateClusterConfigWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfigResourcesVPCConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfigResourcesVPCConfig")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:               r.ko.Spec.Name,
		ResourcesVpcConfig: rm.newVpcConfigRequest(r),
	}

	// From the EKS documentation:
	// "You can't update the subnets or security group IDs for an existing
	// cluster."
	input.ResourcesVpcConfig.SetSubnetIds(nil)
	input.ResourcesVpcConfig.SetSecurityGroupIds(nil)

	_, err = rm.sdkapi.UpdateClusterConfigWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) listNodegroups(
	ctx context.Context,
	r *resource,
) (nodes *svcsdk.ListNodegroupsOutput, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.listNodegroups")
	defer exit(err)
	input := &svcsdk.ListNodegroupsInput{
		ClusterName: r.ko.Spec.Name,
	}

	nodes, err = rm.sdkapi.ListNodegroupsWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "ListNodegroups", err)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
