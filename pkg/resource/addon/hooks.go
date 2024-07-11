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
	"context"
	"fmt"
	"strings"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
)

// Taken from the list of addon statuses in the EKS API documentation:
// https://docs.aws.amazon.com/eks/latest/APIReference/API_Addon.html#AmazonEKS-Type-Addon-status
const (
	StatusActive       = "ACTIVE"
	StatusCreating     = "CREATING"
	StatusCreateFailed = "CREATE_FAILED"
	StatusUpdating     = "UPDATING"
	StatusUpdateFailed = "UPDATE_FAILED"
	StatusDeleting     = "DELETING"
	StatusDeleteFailed = "DELETE_FAILED"
	StatusDegraded     = "DEGRADED"
)

var (
	requeueWaitWhileDeleting = ackrequeue.NeededAfter(
		fmt.Errorf("addon in '%s' state, cannot be modified or deleted", StatusDeleting),
		ackrequeue.DefaultRequeueAfterDuration,
	)
)

var (
	// TerminalStatuses defines the list of statuses that are terminal for an addon
	TerminalStatuses = []string{
		StatusCreateFailed,
		StatusUpdateFailed,
		StatusDeleteFailed,
		// Still not sure if we should consider DEGRADED as terminal
		// StatusDegraded,
	}
)

// addonActive returns true if the supplied addib is in an active state
func addonActive(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusActive
}

// addonCreating returns true if the supplied addon is in a creating state
func addonCreating(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusCreating
}

// addonDeleting returns true if the supplied addon is in a deleting state
func addonDeleting(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusDeleting
}

// addonHasTerminalStatus returns true if the supplied addon is in a terminal state
func addonHasTerminalStatus(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	for _, ts := range TerminalStatuses {
		if cs == ts {
			return true
		}
	}
	return false
}

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the addon cannot be modified until it reaches an active status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.Status == nil {
		return nil
	}
	status := *r.ko.Status.Status
	return ackrequeue.NeededAfter(
		fmt.Errorf("addon in '%s' state, cannot be modified until '%s'",
			status, StatusActive),
		ackrequeue.DefaultRequeueAfterDuration/15,
	)
}

// returnAddonUpdating will set synced to false on the resource and
// return an async requeue error to signify that the resource should be
// forcefully requeued in order to pick up the 'UPDATING' status.
func returnAddonUpdating(r *resource) (*resource, error) {
	msg := "Addon is currently being updated"
	ackcondition.SetSynced(r, corev1.ConditionFalse, &msg, nil)
	return r, ackrequeue.NeededAfter(
		fmt.Errorf("addon in '%s' state, cannot be modified until '%s'",
			StatusUpdating, StatusActive),
		15,
	)
}

var syncTags = tags.SyncTags

// setResourceDefaults queries the EKS API for the current state of the
// fields that are not returned by the ReadOne or List APIs.
func (rm *resourceManager) setResourceAdditionalFields(ctx context.Context, r *v1alpha1.Addon, associationARNs []*string) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceAdditionalFields")
	defer func() { exit(err) }()

	podIdentityAssociations, err := rm.describeAddonPodIdentityAssociations(ctx, r.Spec.ClusterName, associationARNs)
	if err != nil {
		return err
	}
	r.Spec.PodIdentityAssociations = podIdentityAssociations

	return nil
}

// describeAddonPodIdentityAssociations queries the EKS API for the pod identity associations
// associated with the addon.
func (rm *resourceManager) describeAddonPodIdentityAssociations(
	ctx context.Context,
	clusterName *string,
	associationARNs []*string,
) (podIdentityAssociations []*v1alpha1.AddonPodIdentityAssociations, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.describeAddonPodIdentityAssociations")
	defer func() { exit(err) }()

	for _, associationARN := range associationARNs {
		associationID := getAssociationID(*associationARN)
		resp, err := rm.sdkapi.DescribePodIdentityAssociationWithContext(
			ctx,
			&svcsdk.DescribePodIdentityAssociationInput{
				ClusterName:   clusterName,
				AssociationId: &associationID,
			},
		)
		if err != nil {
			return nil, err
		}
		podIdentityAssociations = append(podIdentityAssociations, &v1alpha1.AddonPodIdentityAssociations{
			RoleARN:        resp.Association.RoleArn,
			ServiceAccount: resp.Association.ServiceAccount,
		})
	}

	return podIdentityAssociations, nil
}

// getAssociationID returns the association ID from the association ARN.
func getAssociationID(associationARN string) string {
	parts := strings.Split(associationARN, "/")
	return parts[len(parts)-1]
}

// equalPodIdentityAssociations returns true if the desired and latest pod identity associations are equal
// regardless of the order of the associations.
func equalPodIdentityAssociations(desired, latest []*v1alpha1.AddonPodIdentityAssociations) bool {
	if len(desired) != len(latest) {
		return false
	}

	// Just avoiding unnecessary checks and allocations
	if len(desired) == 0 {
		return true
	}

	// Create a map of the latest pod identity associations for easy lookup
	latestMap := make(map[string]struct{}, len(latest))
	for _, latestAssociation := range latest {
		latestMap[formatPodIdentityAssociation(latestAssociation)] = struct{}{}
	}

	// Check if all desired pod identity associations are present in the latest pod identity associations
	for _, desiredAssociation := range desired {
		if _, ok := latestMap[formatPodIdentityAssociation(desiredAssociation)]; !ok {
			return false
		}
	}

	return true
}

// customPreCompare is a custom pre-compare function that compares the PodIdentityAssociations field
func customPreCompare(delta *ackcompare.Delta, desired, latest *resource) {
	if !equalPodIdentityAssociations(desired.ko.Spec.PodIdentityAssociations, latest.ko.Spec.PodIdentityAssociations) {
		delta.Add("Spec.PodIdentityAssociations", desired.ko.Spec.PodIdentityAssociations, latest.ko.Spec.PodIdentityAssociations)
	}
}

// formatPodIdentityAssociation returns a string representation of the pod identity association
// in the format "serviceAccount/roleARN". This is used to compare the desired and latest
// pod identity associations.
func formatPodIdentityAssociation(association *v1alpha1.AddonPodIdentityAssociations) string {
	serviceAccount := ""
	if association.ServiceAccount != nil {
		serviceAccount = *association.ServiceAccount
	}
	roleARN := ""
	if association.RoleARN != nil {
		roleARN = *association.RoleARN
	}
	return fmt.Sprintf("%s/%s", serviceAccount, roleARN)
}
