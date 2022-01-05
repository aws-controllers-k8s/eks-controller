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

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
)

func (rm *resourceManager) customUpdateCluster(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateCluster")
	defer exit(err)

	if delta.DifferentAt("Spec.Logging") || delta.DifferentAt("Spec.ResourcesVPCConfig") {
		if err := rm.updateConfig(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Version") {
		if err := rm.updateVersion(ctx, desired); err != nil {
			return nil, err
		}
	}

	return desired, nil
}

func (rm *resourceManager) updateVersion(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateVersion")
	defer exit(err)
	input := &svcsdk.UpdateClusterVersionInput{
		Name:    r.ko.Spec.Name,
		Version: r.ko.Spec.Version,
	}

	_, err = rm.sdkapi.UpdateClusterVersionWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterVersion", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfig")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:               r.ko.Spec.Name,
		Logging:            rm.newLogging(r),
		ResourcesVpcConfig: rm.newResourcesVpcConfig(r),
	}

	_, err = rm.sdkapi.UpdateClusterConfigWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}
