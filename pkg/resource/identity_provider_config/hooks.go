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

package identity_provider_config

import (
	"context"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
)

// Taken from the list of nodegroup statuses on the boto3 documentation
// https://https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/eks/client/describe_identity_provider_config.html#describe-identity-provider-config
const (
	StatusCreating = "CREATING"
	StatusActive   = "ACTIVE"
	StatusDeleting = "DELETING"

	IdentityProviderConfigType = "oidc"
)

// customCheckRequiredFieldsMissing returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) customCheckRequiredFieldsMissing(
	r *resource,
) bool {
	if r.ko.Spec.ClusterName == nil || r.ko.Spec.OIDC.IdentityProviderConfigName == nil {
		return true
	}
	return false
}

// identityProviderActive returns true if the supplied cluster is in an active status
func identityProviderActive(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusActive
}

// identityProviderCreating returns true if the supplied cluster is in the process of
// being created
func identityProviderCreating(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusCreating
}

// identityProviderDeleting returns true if the supplied cluster is in the process of
// being deleted
func identityProviderDeleting(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusDeleting
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

	return nil, nil
}
