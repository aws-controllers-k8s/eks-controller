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

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eks"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eks/types"

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

	output, err := rm.sdkapi.ListAssociatedAccessPolicies(
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
			accessScope.Type = aws.String(string(association.AccessScope.Type))
			accessScope.Namespaces = make([]*string, len(association.AccessScope.Namespaces))

			for i, ns := range association.AccessScope.Namespaces {
				accessScope.Namespaces[i] = aws.String(ns)
			}
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

	existingPolicies := latest.ko.Spec.AccessPolicies
	desiredPolicies := desired.ko.Spec.AccessPolicies

	toAdd, toDelete := computeAccessPoliciesDelta(desiredPolicies, existingPolicies)

	// remove policies first (to avoid conflicts)
	for _, p := range toDelete {
		rlog.Debug("disassociating access policy from access entry", "policy_arn", *p)
		if err = rm.disassociateAccessPolicy(ctx, desired, p); err != nil {
			return err
		}
	}
	for _, p := range toAdd {
		rlog.Debug("associating access policy to access entry", "policy_arn", *p.PolicyARN)
		if err = rm.associateAccessPolicy(ctx, desired, p); err != nil {
			return err
		}
	}

	return nil
}

// associateAccessPolicy adds the supplied AccessPolicy to the supplied
// AccessEntry resource.
func (rm *resourceManager) associateAccessPolicy(
	ctx context.Context,
	r *resource,
	entry *v1alpha1.AssociateAccessPolicyInput,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.associateAccessPolicy")
	defer func() { exit(err) }()

	// Convert []*string to []string
	namespaces := make([]string, 0, len(entry.AccessScope.Namespaces))
	for _, ns := range entry.AccessScope.Namespaces {
		if ns != nil {
			namespaces = append(namespaces, *ns)
		}
	}

	input := &svcsdk.AssociateAccessPolicyInput{
		ClusterName:  r.ko.Spec.ClusterName,
		PrincipalArn: r.ko.Spec.PrincipalARN,
		PolicyArn:    entry.PolicyARN,
		AccessScope: &svcsdktypes.AccessScope{
			Type:       svcsdktypes.AccessScopeType(*entry.AccessScope.Type),
			Namespaces: namespaces,
		},
	}
	_, err = rm.sdkapi.AssociateAccessPolicy(ctx, input)
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
	_, err = rm.sdkapi.DisassociateAccessPolicy(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "DisassociateAccessPolicy", err)
	return err
}

// inAccessPolicies returns true if the supplied AccessPolicy ARN exists
// in the slice of AccessPolicy objects.
func inAccessPolicies(policy *v1alpha1.AssociateAccessPolicyInput, policies []*v1alpha1.AssociateAccessPolicyInput) bool {
	for _, p := range policies {
		if *p.PolicyARN == *policy.PolicyARN {
			return true
		}
	}
	return false
}

// equalAccessScopes returns true if the supplied AccessScope objects are
// exactly the same.
func equalAccessScopes(a, b *v1alpha1.AccessScope) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return equalZeroString(b.Type) && len(b.Namespaces) == 0
	}
	if b == nil {
		return equalZeroString(b.Type) && len(a.Namespaces) == 0
	}
	return equalStrings(a.Type, b.Type) && ackcompare.SliceStringPEqual(a.Namespaces, b.Namespaces)
}

// exactMatchInAccessPolicies returns true if the supplied AccessPolicy is in the
// slice of AccessPolicy objects and the AccessScope is exactly the same.
func exactMatchInAccessPolicies(policy *v1alpha1.AssociateAccessPolicyInput, policies []*v1alpha1.AssociateAccessPolicyInput) bool {
	for _, p := range policies {
		if p.PolicyARN == nil {
			continue
		}
		if *p.PolicyARN == *policy.PolicyARN {
			return equalAccessScopes(p.AccessScope, policy.AccessScope)
		}
	}
	return false
}

// computeAccessPoliciesDelta returns two slices of AccessPolicy objects: one
// slice of AccessPolicy objects that are in the desired slice but not in the
// latest slice, and one slice of AccessPolicy objects that are in the latest
// slice but not in the desired slice.
func computeAccessPoliciesDelta(desired, latest []*v1alpha1.AssociateAccessPolicyInput) (toAdd []*v1alpha1.AssociateAccessPolicyInput, toDelete []*string) {
	// useful for the toDelete elements
	visited := map[string]bool{}

	// First we need to loop through the desired policies and see if they are
	// in the latest policies. If they are, we need to check if the AccessScope
	// is the same. If it is, we don't need to do anything. If it's not, we need
	// to add the policy to the toAdd slice and remove it from the toDelete slice.
	//
	// The delete is necessary because the API dosen't allow us to update the
	// AccessScope of an existing policy. We need to disassociate the policy and
	// then reassociate it.
	for _, p := range desired {
		visited[*p.PolicyARN] = true
		// If it's an exact match, we don't need to do anything.
		if exactMatchInAccessPolicies(p, latest) {
			continue
		}
		// If it's in the latest policies, but the AccessScope is different, we
		// need to remove it from the toDelete slice (update).
		if inAccessPolicies(p, latest) {
			toAdd = append(toAdd, p)
			toDelete = append(toDelete, p.PolicyARN)
		} else {
			// If it's not in the latest policies, we need to add it.
			toAdd = append(toAdd, p)
		}
	}

	// Now that we've handled the desired policies, we need to loop through the
	// latest policies and see if they are in the desired policies. If they are
	// not, we need to add them to the toDelete slice.
	for _, p := range latest {
		// If we've already visited this policy, we don't need to do anything.
		if visited[*p.PolicyARN] {
			continue
		}
		// If it's not in the desired policies, we need to remove it.
		if !inAccessPolicies(p, desired) {
			toDelete = append(toDelete, p.PolicyARN)
		}
	}
	return toAdd, toDelete
}

func customPreCompare(delta *ackcompare.Delta, a, b *resource) {
	if len(a.ko.Spec.AccessPolicies) != len(b.ko.Spec.AccessPolicies) {
		delta.Add("Spec.AccessPolicies", a.ko.Spec.AccessPolicies, b.ko.Spec.AccessPolicies)
	} else if toAdd, toRemove := computeAccessPoliciesDelta(a.ko.Spec.AccessPolicies, b.ko.Spec.AccessPolicies); len(toAdd) > 0 || len(toRemove) > 0 {
		delta.Add("Spec.AccessPolicies", a.ko.Spec.AccessPolicies, b.ko.Spec.AccessPolicies)
	}
}

// EqualStrings returns true if two strings are equal e.g., both are nil, one is
// nil and the other is empty string, or both non-zero strings are equal.
func equalStrings(a, b *string) bool {
	if a == nil {
		return b == nil || *b == ""
	}
	if b == nil {
		return *a == ""
	}
	return *a == *b
}

func equalZeroString(a *string) bool {
	return equalStrings(a, aws.String(""))
}
