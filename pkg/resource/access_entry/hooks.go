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
	"context"
	"reflect"

	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
)

// Ideally, a part of this code needs to be generated.. However since the
// tags packge is not imported, we can't call it directly from sdk.go. We
// have to do this Go-fu to make it work.
var syncTags = tags.SyncTags

// setResourceDefaults queries the EKS API for the current state of the
// fields that are not returned by the ReadOne or List APIs. In this
// case, we're populate the AccessEntry.Status fields with the output
// of ListAssociatedAccessPolicies.
func (rm *resourceManager) setResourceAdditionalFields(ctx context.Context, r *v1alpha1.AccessEntry) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setResourceAdditionalFields")
	defer exit(err)

	err = rm.getAccessEntryAssociatedPolicies(ctx, r)
	if err != nil {
		return err
	}
	return nil
}

func (rm *resourceManager) getAccessEntryAssociatedPolicies(ctx context.Context, r *v1alpha1.AccessEntry) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.setAccessEntryAssociatedPolicies")
	defer exit(err)

	output, err := rm.sdkapi.ListAssociatedAccessPoliciesWithContext(
		ctx,
		&svcsdk.ListAssociatedAccessPoliciesInput{
			ClusterName:  r.Spec.ClusterName,
			PrincipalArn: r.Spec.PrincipalARN,
		},
	)
	rm.metrics.RecordAPICall("GET", "ListAssociatedAccessPolicies", err)
	if err != nil {
		return err
	}
	// reset the access policies
	r.Spec.AccessPolicies = nil

	// populate the access policies
	for _, association := range output.AssociatedAccessPolicies {
		accessScope := &v1alpha1.AccessScope{}
		if association.AccessScope != nil {
			accessScope.Type = association.AccessScope.Type
			accessScope.Namespaces = association.AccessScope.Namespaces
		}
		r.Spec.AccessPolicies = append(r.Spec.AccessPolicies, &v1alpha1.AssociateAccessPolicyInput{
			PolicyARN:   association.PolicyArn,
			AccessScope: accessScope,
		})
	}

	return nil
}

// syncAccessPolicies examines the AccessPolicies in the desired AccessEntry
// and calls the AssociateAccessPolicy and DisassociateAccessPolicy APIs to
// ensure that the set of associated AccessPolicies stays in sync with the
func (rm *resourceManager) syncAccessPolicies(ctx context.Context, desired, latest *resource) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncAccessPolicies")
	defer func() { exit(err) }()
	toAdd := []*v1alpha1.AssociateAccessPolicyInput{}
	toDelete := []*string{}

	existingPolicies := latest.ko.Spec.AccessPolicies

	// find the policies to add
	for _, p := range desired.ko.Spec.AccessPolicies {
		if !exactMatchInAccessPolicies(p, existingPolicies) {
			toAdd = append(toAdd, p)
		}
	}

	// find the policies to delete
	for _, p := range existingPolicies {
		if !inAccessPolicies(p, desired.ko.Spec.AccessPolicies) {
			toDelete = append(toDelete, p.PolicyARN)
		}
	}

	// manage policies...
	for _, p := range toDelete {
		rlog.Debug("disassociating access policy from role", "policy_arn", *p)
		if err = rm.disassociateAccessPolicy(ctx, desired, p); err != nil {
			return err
		}
	}
	for _, p := range toAdd {
		rlog.Debug("associate access policy to access entry", "policy_arn", *p.PolicyARN)
		if err = rm.associateAccessPolicy(ctx, desired, p); err != nil {
			return err
		}
	}

	return nil
}

// inAccessPolicies returns true if the supplied AccessPolicy ARN exists
// in the slice of AccessPolicy objects.
func inAccessPolicies(policy *v1alpha1.AssociateAccessPolicyInput, policies []*v1alpha1.AssociateAccessPolicyInput) bool {
	for _, p := range policies {
		if p.PolicyARN == policy.PolicyARN {
			return false
		}
	}
	return false
}

// exactMatchInAccessPolicies returns true if the supplied AccessPolicy is in the
// slice of AccessPolicy objects and the AccessScope is exactly the same.
func exactMatchInAccessPolicies(policy *v1alpha1.AssociateAccessPolicyInput, policies []*v1alpha1.AssociateAccessPolicyInput) bool {
	for _, p := range policies {
		if p.PolicyARN == policy.PolicyARN {
			return reflect.DeepEqual(p.AccessScope, policy.AccessScope)
		}
	}
	return false
}

// associateAccessPolicy adds the supplied AccessPolicy to the supplied
// AccessEntry resource.
func (rm *resourceManager) associateAccessPolicy(
	ctx context.Context,
	r *resource,
	entry *v1alpha1.AssociateAccessPolicyInput,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.addManagedPolicy")
	defer func() { exit(err) }()

	input := &svcsdk.AssociateAccessPolicyInput{
		ClusterName:  r.ko.Spec.ClusterName,
		PrincipalArn: r.ko.Spec.PrincipalARN,
		PolicyArn:    entry.PolicyARN,
		AccessScope: &svcsdk.AccessScope{
			Type:       entry.AccessScope.Type,
			Namespaces: entry.AccessScope.Namespaces,
		},
	}
	_, err = rm.sdkapi.AssociateAccessPolicyWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "AssociateAccessPolicy", err)
	return err
}

// disassociateAccessPolicy removes the supplied AccessPolicy from the supplied
// AccessEntry resource.
func (rm *resourceManager) disassociateAccessPolicy(
	ctx context.Context,
	r *resource,
	policyARN *string,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.disassociateAccessPolicy")
	defer func() { exit(err) }()

	input := &svcsdk.DisassociateAccessPolicyInput{
		ClusterName:  r.ko.Spec.ClusterName,
		PrincipalArn: r.ko.Spec.PrincipalARN,
		PolicyArn:    policyARN,
	}
	_, err = rm.sdkapi.DisassociateAccessPolicyWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "DisassociateAccessPolicy", err)
	return err
}
