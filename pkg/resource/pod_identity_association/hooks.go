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
	"context"

	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eks"
)

var syncTags = tags.SyncTags

func (rm *resourceManager) getAssociationID(ctx context.Context, r *resource) (id *string, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getSecretID")
	defer func() {
		exit(err)
	}()

	// ClusterName is a required field for ListPodIdentityAssociations operation
	// we treat an undefined ClusterName as not found.
	if r.ko.Spec.ClusterName == nil {
		return nil, nil
	}

	resp, err := rm.sdkapi.ListPodIdentityAssociations(ctx, &svcsdk.ListPodIdentityAssociationsInput{
		ClusterName:    r.ko.Spec.ClusterName,
		Namespace:      r.ko.Spec.Namespace,
		ServiceAccount: r.ko.Spec.ServiceAccount,
	})
	if err != nil {
		return nil, err
	}

	// if more than one are returned, we don't want to manage them
	// and treat it as not found
	if len(resp.Associations) != 1 {
		return nil, nil
	}

	return resp.Associations[0].AssociationId, nil

}
