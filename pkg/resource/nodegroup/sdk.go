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

// Code generated by ack-generate. DO NOT EDIT.

package nodegroup

import (
	"context"
	"reflect"
	"strings"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &aws.JSONValue{}
	_ = &svcsdk.EKS{}
	_ = &svcapitypes.Nodegroup{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
	_ = &ackcondition.NotManagedMessage
	_ = &reflect.Value{}
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkFind")
	defer exit(err)
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromReadOneInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.DescribeNodegroupOutput
	resp, err = rm.sdkapi.DescribeNodegroupWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "DescribeNodegroup", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "ResourceNotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	if resp.Nodegroup.AmiType != nil {
		ko.Spec.AMIType = resp.Nodegroup.AmiType
	} else {
		ko.Spec.AMIType = nil
	}
	if resp.Nodegroup.CapacityType != nil {
		ko.Spec.CapacityType = resp.Nodegroup.CapacityType
	} else {
		ko.Spec.CapacityType = nil
	}
	if resp.Nodegroup.ClusterName != nil {
		ko.Spec.ClusterName = resp.Nodegroup.ClusterName
	} else {
		ko.Spec.ClusterName = nil
	}
	if resp.Nodegroup.CreatedAt != nil {
		ko.Status.CreatedAt = &metav1.Time{*resp.Nodegroup.CreatedAt}
	} else {
		ko.Status.CreatedAt = nil
	}
	if resp.Nodegroup.DiskSize != nil {
		ko.Spec.DiskSize = resp.Nodegroup.DiskSize
	} else {
		ko.Spec.DiskSize = nil
	}
	if resp.Nodegroup.Health != nil {
		f5 := &svcapitypes.NodegroupHealth{}
		if resp.Nodegroup.Health.Issues != nil {
			f5f0 := []*svcapitypes.Issue{}
			for _, f5f0iter := range resp.Nodegroup.Health.Issues {
				f5f0elem := &svcapitypes.Issue{}
				if f5f0iter.Code != nil {
					f5f0elem.Code = f5f0iter.Code
				}
				if f5f0iter.Message != nil {
					f5f0elem.Message = f5f0iter.Message
				}
				if f5f0iter.ResourceIds != nil {
					f5f0elemf2 := []*string{}
					for _, f5f0elemf2iter := range f5f0iter.ResourceIds {
						var f5f0elemf2elem string
						f5f0elemf2elem = *f5f0elemf2iter
						f5f0elemf2 = append(f5f0elemf2, &f5f0elemf2elem)
					}
					f5f0elem.ResourceIDs = f5f0elemf2
				}
				f5f0 = append(f5f0, f5f0elem)
			}
			f5.Issues = f5f0
		}
		ko.Status.Health = f5
	} else {
		ko.Status.Health = nil
	}
	if resp.Nodegroup.InstanceTypes != nil {
		f6 := []*string{}
		for _, f6iter := range resp.Nodegroup.InstanceTypes {
			var f6elem string
			f6elem = *f6iter
			f6 = append(f6, &f6elem)
		}
		ko.Spec.InstanceTypes = f6
	} else {
		ko.Spec.InstanceTypes = nil
	}
	if resp.Nodegroup.Labels != nil {
		f7 := map[string]*string{}
		for f7key, f7valiter := range resp.Nodegroup.Labels {
			var f7val string
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		ko.Spec.Labels = f7
	} else {
		ko.Spec.Labels = nil
	}
	if resp.Nodegroup.LaunchTemplate != nil {
		f8 := &svcapitypes.LaunchTemplateSpecification{}
		if resp.Nodegroup.LaunchTemplate.Id != nil {
			f8.ID = resp.Nodegroup.LaunchTemplate.Id
		}
		if resp.Nodegroup.LaunchTemplate.Name != nil {
			f8.Name = resp.Nodegroup.LaunchTemplate.Name
		}
		if resp.Nodegroup.LaunchTemplate.Version != nil {
			f8.Version = resp.Nodegroup.LaunchTemplate.Version
		}
		ko.Spec.LaunchTemplate = f8
	} else {
		ko.Spec.LaunchTemplate = nil
	}
	if resp.Nodegroup.ModifiedAt != nil {
		ko.Status.ModifiedAt = &metav1.Time{*resp.Nodegroup.ModifiedAt}
	} else {
		ko.Status.ModifiedAt = nil
	}
	if resp.Nodegroup.NodeRole != nil {
		ko.Spec.NodeRole = resp.Nodegroup.NodeRole
	} else {
		ko.Spec.NodeRole = nil
	}
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.Nodegroup.NodegroupArn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.Nodegroup.NodegroupArn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Nodegroup.NodegroupName != nil {
		ko.Spec.Name = resp.Nodegroup.NodegroupName
	} else {
		ko.Spec.Name = nil
	}
	if resp.Nodegroup.ReleaseVersion != nil {
		ko.Spec.ReleaseVersion = resp.Nodegroup.ReleaseVersion
	} else {
		ko.Spec.ReleaseVersion = nil
	}
	if resp.Nodegroup.RemoteAccess != nil {
		f14 := &svcapitypes.RemoteAccessConfig{}
		if resp.Nodegroup.RemoteAccess.Ec2SshKey != nil {
			f14.EC2SshKey = resp.Nodegroup.RemoteAccess.Ec2SshKey
		}
		if resp.Nodegroup.RemoteAccess.SourceSecurityGroups != nil {
			f14f1 := []*string{}
			for _, f14f1iter := range resp.Nodegroup.RemoteAccess.SourceSecurityGroups {
				var f14f1elem string
				f14f1elem = *f14f1iter
				f14f1 = append(f14f1, &f14f1elem)
			}
			f14.SourceSecurityGroups = f14f1
		}
		ko.Spec.RemoteAccess = f14
	} else {
		ko.Spec.RemoteAccess = nil
	}
	if resp.Nodegroup.Resources != nil {
		f15 := &svcapitypes.NodegroupResources{}
		if resp.Nodegroup.Resources.AutoScalingGroups != nil {
			f15f0 := []*svcapitypes.AutoScalingGroup{}
			for _, f15f0iter := range resp.Nodegroup.Resources.AutoScalingGroups {
				f15f0elem := &svcapitypes.AutoScalingGroup{}
				if f15f0iter.Name != nil {
					f15f0elem.Name = f15f0iter.Name
				}
				f15f0 = append(f15f0, f15f0elem)
			}
			f15.AutoScalingGroups = f15f0
		}
		if resp.Nodegroup.Resources.RemoteAccessSecurityGroup != nil {
			f15.RemoteAccessSecurityGroup = resp.Nodegroup.Resources.RemoteAccessSecurityGroup
		}
		ko.Status.Resources = f15
	} else {
		ko.Status.Resources = nil
	}
	if resp.Nodegroup.ScalingConfig != nil {
		f16 := &svcapitypes.NodegroupScalingConfig{}
		if resp.Nodegroup.ScalingConfig.DesiredSize != nil {
			f16.DesiredSize = resp.Nodegroup.ScalingConfig.DesiredSize
		}
		if resp.Nodegroup.ScalingConfig.MaxSize != nil {
			f16.MaxSize = resp.Nodegroup.ScalingConfig.MaxSize
		}
		if resp.Nodegroup.ScalingConfig.MinSize != nil {
			f16.MinSize = resp.Nodegroup.ScalingConfig.MinSize
		}
		ko.Spec.ScalingConfig = f16
	} else {
		ko.Spec.ScalingConfig = nil
	}
	if resp.Nodegroup.Status != nil {
		ko.Status.Status = resp.Nodegroup.Status
	} else {
		ko.Status.Status = nil
	}
	if resp.Nodegroup.Subnets != nil {
		f18 := []*string{}
		for _, f18iter := range resp.Nodegroup.Subnets {
			var f18elem string
			f18elem = *f18iter
			f18 = append(f18, &f18elem)
		}
		ko.Spec.Subnets = f18
	} else {
		ko.Spec.Subnets = nil
	}
	if resp.Nodegroup.Tags != nil {
		f19 := map[string]*string{}
		for f19key, f19valiter := range resp.Nodegroup.Tags {
			var f19val string
			f19val = *f19valiter
			f19[f19key] = &f19val
		}
		ko.Spec.Tags = f19
	} else {
		ko.Spec.Tags = nil
	}
	if resp.Nodegroup.Taints != nil {
		f20 := []*svcapitypes.Taint{}
		for _, f20iter := range resp.Nodegroup.Taints {
			f20elem := &svcapitypes.Taint{}
			if f20iter.Effect != nil {
				f20elem.Effect = f20iter.Effect
			}
			if f20iter.Key != nil {
				f20elem.Key = f20iter.Key
			}
			if f20iter.Value != nil {
				f20elem.Value = f20iter.Value
			}
			f20 = append(f20, f20elem)
		}
		ko.Spec.Taints = f20
	} else {
		ko.Spec.Taints = nil
	}
	if resp.Nodegroup.UpdateConfig != nil {
		f21 := &svcapitypes.NodegroupUpdateConfig{}
		if resp.Nodegroup.UpdateConfig.MaxUnavailable != nil {
			f21.MaxUnavailable = resp.Nodegroup.UpdateConfig.MaxUnavailable
		}
		if resp.Nodegroup.UpdateConfig.MaxUnavailablePercentage != nil {
			f21.MaxUnavailablePercentage = resp.Nodegroup.UpdateConfig.MaxUnavailablePercentage
		}
		ko.Spec.UpdateConfig = f21
	} else {
		ko.Spec.UpdateConfig = nil
	}
	if resp.Nodegroup.Version != nil {
		ko.Spec.Version = resp.Nodegroup.Version
	} else {
		ko.Spec.Version = nil
	}

	rm.setStatusDefaults(ko)
	if !nodegroupActive(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
	} else {
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionTrue, nil, nil)
	}

	return &resource{ko}, nil
}

// requiredFieldsMissingFromReadOneInput returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadOneInput(
	r *resource,
) bool {
	return r.ko.Spec.ClusterName == nil || r.ko.Spec.Name == nil

}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.DescribeNodegroupInput, error) {
	res := &svcsdk.DescribeNodegroupInput{}

	if r.ko.Spec.ClusterName != nil {
		res.SetClusterName(*r.ko.Spec.ClusterName)
	}
	if r.ko.Spec.Name != nil {
		res.SetNodegroupName(*r.ko.Spec.Name)
	}

	return res, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a copy of the resource with resource fields (in both Spec and
// Status) filled in with values from the CREATE API operation's Output shape.
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	desired *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkCreate")
	defer exit(err)
	input, err := rm.newCreateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.CreateNodegroupOutput
	_ = resp
	resp, err = rm.sdkapi.CreateNodegroupWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateNodegroup", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.Nodegroup.AmiType != nil {
		ko.Spec.AMIType = resp.Nodegroup.AmiType
	} else {
		ko.Spec.AMIType = nil
	}
	if resp.Nodegroup.CapacityType != nil {
		ko.Spec.CapacityType = resp.Nodegroup.CapacityType
	} else {
		ko.Spec.CapacityType = nil
	}
	if resp.Nodegroup.ClusterName != nil {
		ko.Spec.ClusterName = resp.Nodegroup.ClusterName
	} else {
		ko.Spec.ClusterName = nil
	}
	if resp.Nodegroup.CreatedAt != nil {
		ko.Status.CreatedAt = &metav1.Time{*resp.Nodegroup.CreatedAt}
	} else {
		ko.Status.CreatedAt = nil
	}
	if resp.Nodegroup.DiskSize != nil {
		ko.Spec.DiskSize = resp.Nodegroup.DiskSize
	} else {
		ko.Spec.DiskSize = nil
	}
	if resp.Nodegroup.Health != nil {
		f5 := &svcapitypes.NodegroupHealth{}
		if resp.Nodegroup.Health.Issues != nil {
			f5f0 := []*svcapitypes.Issue{}
			for _, f5f0iter := range resp.Nodegroup.Health.Issues {
				f5f0elem := &svcapitypes.Issue{}
				if f5f0iter.Code != nil {
					f5f0elem.Code = f5f0iter.Code
				}
				if f5f0iter.Message != nil {
					f5f0elem.Message = f5f0iter.Message
				}
				if f5f0iter.ResourceIds != nil {
					f5f0elemf2 := []*string{}
					for _, f5f0elemf2iter := range f5f0iter.ResourceIds {
						var f5f0elemf2elem string
						f5f0elemf2elem = *f5f0elemf2iter
						f5f0elemf2 = append(f5f0elemf2, &f5f0elemf2elem)
					}
					f5f0elem.ResourceIDs = f5f0elemf2
				}
				f5f0 = append(f5f0, f5f0elem)
			}
			f5.Issues = f5f0
		}
		ko.Status.Health = f5
	} else {
		ko.Status.Health = nil
	}
	if resp.Nodegroup.InstanceTypes != nil {
		f6 := []*string{}
		for _, f6iter := range resp.Nodegroup.InstanceTypes {
			var f6elem string
			f6elem = *f6iter
			f6 = append(f6, &f6elem)
		}
		ko.Spec.InstanceTypes = f6
	} else {
		ko.Spec.InstanceTypes = nil
	}
	if resp.Nodegroup.Labels != nil {
		f7 := map[string]*string{}
		for f7key, f7valiter := range resp.Nodegroup.Labels {
			var f7val string
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		ko.Spec.Labels = f7
	} else {
		ko.Spec.Labels = nil
	}
	if resp.Nodegroup.LaunchTemplate != nil {
		f8 := &svcapitypes.LaunchTemplateSpecification{}
		if resp.Nodegroup.LaunchTemplate.Id != nil {
			f8.ID = resp.Nodegroup.LaunchTemplate.Id
		}
		if resp.Nodegroup.LaunchTemplate.Name != nil {
			f8.Name = resp.Nodegroup.LaunchTemplate.Name
		}
		if resp.Nodegroup.LaunchTemplate.Version != nil {
			f8.Version = resp.Nodegroup.LaunchTemplate.Version
		}
		ko.Spec.LaunchTemplate = f8
	} else {
		ko.Spec.LaunchTemplate = nil
	}
	if resp.Nodegroup.ModifiedAt != nil {
		ko.Status.ModifiedAt = &metav1.Time{*resp.Nodegroup.ModifiedAt}
	} else {
		ko.Status.ModifiedAt = nil
	}
	if resp.Nodegroup.NodeRole != nil {
		ko.Spec.NodeRole = resp.Nodegroup.NodeRole
	} else {
		ko.Spec.NodeRole = nil
	}
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.Nodegroup.NodegroupArn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.Nodegroup.NodegroupArn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Nodegroup.NodegroupName != nil {
		ko.Spec.Name = resp.Nodegroup.NodegroupName
	} else {
		ko.Spec.Name = nil
	}
	if resp.Nodegroup.ReleaseVersion != nil {
		ko.Spec.ReleaseVersion = resp.Nodegroup.ReleaseVersion
	} else {
		ko.Spec.ReleaseVersion = nil
	}
	if resp.Nodegroup.RemoteAccess != nil {
		f14 := &svcapitypes.RemoteAccessConfig{}
		if resp.Nodegroup.RemoteAccess.Ec2SshKey != nil {
			f14.EC2SshKey = resp.Nodegroup.RemoteAccess.Ec2SshKey
		}
		if resp.Nodegroup.RemoteAccess.SourceSecurityGroups != nil {
			f14f1 := []*string{}
			for _, f14f1iter := range resp.Nodegroup.RemoteAccess.SourceSecurityGroups {
				var f14f1elem string
				f14f1elem = *f14f1iter
				f14f1 = append(f14f1, &f14f1elem)
			}
			f14.SourceSecurityGroups = f14f1
		}
		ko.Spec.RemoteAccess = f14
	} else {
		ko.Spec.RemoteAccess = nil
	}
	if resp.Nodegroup.Resources != nil {
		f15 := &svcapitypes.NodegroupResources{}
		if resp.Nodegroup.Resources.AutoScalingGroups != nil {
			f15f0 := []*svcapitypes.AutoScalingGroup{}
			for _, f15f0iter := range resp.Nodegroup.Resources.AutoScalingGroups {
				f15f0elem := &svcapitypes.AutoScalingGroup{}
				if f15f0iter.Name != nil {
					f15f0elem.Name = f15f0iter.Name
				}
				f15f0 = append(f15f0, f15f0elem)
			}
			f15.AutoScalingGroups = f15f0
		}
		if resp.Nodegroup.Resources.RemoteAccessSecurityGroup != nil {
			f15.RemoteAccessSecurityGroup = resp.Nodegroup.Resources.RemoteAccessSecurityGroup
		}
		ko.Status.Resources = f15
	} else {
		ko.Status.Resources = nil
	}
	if resp.Nodegroup.ScalingConfig != nil {
		f16 := &svcapitypes.NodegroupScalingConfig{}
		if resp.Nodegroup.ScalingConfig.DesiredSize != nil {
			f16.DesiredSize = resp.Nodegroup.ScalingConfig.DesiredSize
		}
		if resp.Nodegroup.ScalingConfig.MaxSize != nil {
			f16.MaxSize = resp.Nodegroup.ScalingConfig.MaxSize
		}
		if resp.Nodegroup.ScalingConfig.MinSize != nil {
			f16.MinSize = resp.Nodegroup.ScalingConfig.MinSize
		}
		ko.Spec.ScalingConfig = f16
	} else {
		ko.Spec.ScalingConfig = nil
	}
	if resp.Nodegroup.Status != nil {
		ko.Status.Status = resp.Nodegroup.Status
	} else {
		ko.Status.Status = nil
	}
	if resp.Nodegroup.Subnets != nil {
		f18 := []*string{}
		for _, f18iter := range resp.Nodegroup.Subnets {
			var f18elem string
			f18elem = *f18iter
			f18 = append(f18, &f18elem)
		}
		ko.Spec.Subnets = f18
	} else {
		ko.Spec.Subnets = nil
	}
	if resp.Nodegroup.Tags != nil {
		f19 := map[string]*string{}
		for f19key, f19valiter := range resp.Nodegroup.Tags {
			var f19val string
			f19val = *f19valiter
			f19[f19key] = &f19val
		}
		ko.Spec.Tags = f19
	} else {
		ko.Spec.Tags = nil
	}
	if resp.Nodegroup.Taints != nil {
		f20 := []*svcapitypes.Taint{}
		for _, f20iter := range resp.Nodegroup.Taints {
			f20elem := &svcapitypes.Taint{}
			if f20iter.Effect != nil {
				f20elem.Effect = f20iter.Effect
			}
			if f20iter.Key != nil {
				f20elem.Key = f20iter.Key
			}
			if f20iter.Value != nil {
				f20elem.Value = f20iter.Value
			}
			f20 = append(f20, f20elem)
		}
		ko.Spec.Taints = f20
	} else {
		ko.Spec.Taints = nil
	}
	if resp.Nodegroup.UpdateConfig != nil {
		f21 := &svcapitypes.NodegroupUpdateConfig{}
		if resp.Nodegroup.UpdateConfig.MaxUnavailable != nil {
			f21.MaxUnavailable = resp.Nodegroup.UpdateConfig.MaxUnavailable
		}
		if resp.Nodegroup.UpdateConfig.MaxUnavailablePercentage != nil {
			f21.MaxUnavailablePercentage = resp.Nodegroup.UpdateConfig.MaxUnavailablePercentage
		}
		ko.Spec.UpdateConfig = f21
	} else {
		ko.Spec.UpdateConfig = nil
	}
	if resp.Nodegroup.Version != nil {
		ko.Spec.Version = resp.Nodegroup.Version
	} else {
		ko.Spec.Version = nil
	}

	rm.setStatusDefaults(ko)
	// We expect the nodegorup to be in 'CREATING' status since we just issued
	// the call to create it, but I suppose it doesn't hurt to check here.
	if nodegroupCreating(&resource{ko}) {
		// Setting resource synced condition to false will trigger a requeue of
		// the resource. No need to return a requeue error here.
		ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, nil, nil)
		return &resource{ko}, nil
	}

	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateNodegroupInput, error) {
	res := &svcsdk.CreateNodegroupInput{}

	if r.ko.Spec.AMIType != nil {
		res.SetAmiType(*r.ko.Spec.AMIType)
	}
	if r.ko.Spec.CapacityType != nil {
		res.SetCapacityType(*r.ko.Spec.CapacityType)
	}
	if r.ko.Spec.ClientRequestToken != nil {
		res.SetClientRequestToken(*r.ko.Spec.ClientRequestToken)
	}
	if r.ko.Spec.ClusterName != nil {
		res.SetClusterName(*r.ko.Spec.ClusterName)
	}
	if r.ko.Spec.DiskSize != nil {
		res.SetDiskSize(*r.ko.Spec.DiskSize)
	}
	if r.ko.Spec.InstanceTypes != nil {
		f5 := []*string{}
		for _, f5iter := range r.ko.Spec.InstanceTypes {
			var f5elem string
			f5elem = *f5iter
			f5 = append(f5, &f5elem)
		}
		res.SetInstanceTypes(f5)
	}
	if r.ko.Spec.Labels != nil {
		f6 := map[string]*string{}
		for f6key, f6valiter := range r.ko.Spec.Labels {
			var f6val string
			f6val = *f6valiter
			f6[f6key] = &f6val
		}
		res.SetLabels(f6)
	}
	if r.ko.Spec.LaunchTemplate != nil {
		f7 := &svcsdk.LaunchTemplateSpecification{}
		if r.ko.Spec.LaunchTemplate.ID != nil {
			f7.SetId(*r.ko.Spec.LaunchTemplate.ID)
		}
		if r.ko.Spec.LaunchTemplate.Name != nil {
			f7.SetName(*r.ko.Spec.LaunchTemplate.Name)
		}
		if r.ko.Spec.LaunchTemplate.Version != nil {
			f7.SetVersion(*r.ko.Spec.LaunchTemplate.Version)
		}
		res.SetLaunchTemplate(f7)
	}
	if r.ko.Spec.NodeRole != nil {
		res.SetNodeRole(*r.ko.Spec.NodeRole)
	}
	if r.ko.Spec.Name != nil {
		res.SetNodegroupName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.ReleaseVersion != nil {
		res.SetReleaseVersion(*r.ko.Spec.ReleaseVersion)
	}
	if r.ko.Spec.RemoteAccess != nil {
		f11 := &svcsdk.RemoteAccessConfig{}
		if r.ko.Spec.RemoteAccess.EC2SshKey != nil {
			f11.SetEc2SshKey(*r.ko.Spec.RemoteAccess.EC2SshKey)
		}
		if r.ko.Spec.RemoteAccess.SourceSecurityGroups != nil {
			f11f1 := []*string{}
			for _, f11f1iter := range r.ko.Spec.RemoteAccess.SourceSecurityGroups {
				var f11f1elem string
				f11f1elem = *f11f1iter
				f11f1 = append(f11f1, &f11f1elem)
			}
			f11.SetSourceSecurityGroups(f11f1)
		}
		res.SetRemoteAccess(f11)
	}
	if r.ko.Spec.ScalingConfig != nil {
		f12 := &svcsdk.NodegroupScalingConfig{}
		if r.ko.Spec.ScalingConfig.DesiredSize != nil {
			f12.SetDesiredSize(*r.ko.Spec.ScalingConfig.DesiredSize)
		}
		if r.ko.Spec.ScalingConfig.MaxSize != nil {
			f12.SetMaxSize(*r.ko.Spec.ScalingConfig.MaxSize)
		}
		if r.ko.Spec.ScalingConfig.MinSize != nil {
			f12.SetMinSize(*r.ko.Spec.ScalingConfig.MinSize)
		}
		res.SetScalingConfig(f12)
	}
	if r.ko.Spec.Subnets != nil {
		f13 := []*string{}
		for _, f13iter := range r.ko.Spec.Subnets {
			var f13elem string
			f13elem = *f13iter
			f13 = append(f13, &f13elem)
		}
		res.SetSubnets(f13)
	}
	if r.ko.Spec.Tags != nil {
		f14 := map[string]*string{}
		for f14key, f14valiter := range r.ko.Spec.Tags {
			var f14val string
			f14val = *f14valiter
			f14[f14key] = &f14val
		}
		res.SetTags(f14)
	}
	if r.ko.Spec.Taints != nil {
		f15 := []*svcsdk.Taint{}
		for _, f15iter := range r.ko.Spec.Taints {
			f15elem := &svcsdk.Taint{}
			if f15iter.Effect != nil {
				f15elem.SetEffect(*f15iter.Effect)
			}
			if f15iter.Key != nil {
				f15elem.SetKey(*f15iter.Key)
			}
			if f15iter.Value != nil {
				f15elem.SetValue(*f15iter.Value)
			}
			f15 = append(f15, f15elem)
		}
		res.SetTaints(f15)
	}
	if r.ko.Spec.UpdateConfig != nil {
		f16 := &svcsdk.NodegroupUpdateConfig{}
		if r.ko.Spec.UpdateConfig.MaxUnavailable != nil {
			f16.SetMaxUnavailable(*r.ko.Spec.UpdateConfig.MaxUnavailable)
		}
		if r.ko.Spec.UpdateConfig.MaxUnavailablePercentage != nil {
			f16.SetMaxUnavailablePercentage(*r.ko.Spec.UpdateConfig.MaxUnavailablePercentage)
		}
		res.SetUpdateConfig(f16)
	}
	if r.ko.Spec.Version != nil {
		res.SetVersion(*r.ko.Spec.Version)
	}

	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	return rm.customUpdate(ctx, desired, latest, delta)
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer exit(err)
	if nodegroupDeleting(r) {
		return r, requeueWaitWhileDeleting
	}

	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteNodegroupOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteNodegroupWithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteNodegroup", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteNodegroupInput, error) {
	res := &svcsdk.DeleteNodegroupInput{}

	if r.ko.Spec.ClusterName != nil {
		res.SetClusterName(*r.ko.Spec.ClusterName)
	}
	if r.ko.Spec.Name != nil {
		res.SetNodegroupName(*r.ko.Spec.Name)
	}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.Nodegroup,
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions(
	r *resource,
	onSuccess bool,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	var recoverableCondition *ackv1alpha1.Condition = nil
	var syncCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeRecoverable {
			recoverableCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			syncCondition = condition
		}
	}

	if rm.terminalAWSError(err) || err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type: ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		var errorMessage = ""
		if err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound {
			errorMessage = err.Error()
		} else {
			awsErr, _ := ackerr.AWSError(err)
			errorMessage = awsErr.Error()
		}
		terminalCondition.Status = corev1.ConditionTrue
		terminalCondition.Message = &errorMessage
	} else {
		// Clear the terminal condition if no longer present
		if terminalCondition != nil {
			terminalCondition.Status = corev1.ConditionFalse
			terminalCondition.Message = nil
		}
		// Handling Recoverable Conditions
		if err != nil {
			if recoverableCondition == nil {
				// Add a new Condition containing a non-terminal error
				recoverableCondition = &ackv1alpha1.Condition{
					Type: ackv1alpha1.ConditionTypeRecoverable,
				}
				ko.Status.Conditions = append(ko.Status.Conditions, recoverableCondition)
			}
			recoverableCondition.Status = corev1.ConditionTrue
			awsErr, _ := ackerr.AWSError(err)
			errorMessage := err.Error()
			if awsErr != nil {
				errorMessage = awsErr.Error()
			}
			recoverableCondition.Message = &errorMessage
		} else if recoverableCondition != nil {
			recoverableCondition.Status = corev1.ConditionFalse
			recoverableCondition.Message = nil
		}
	}
	// Required to avoid the "declared but not used" error in the default case
	_ = syncCondition
	if terminalCondition != nil || recoverableCondition != nil || syncCondition != nil {
		return &resource{ko}, true // updated
	}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
	if err == nil {
		return false
	}
	awsErr, ok := ackerr.AWSError(err)
	if !ok {
		return false
	}
	switch awsErr.Code() {
	case "ResourceLimitExceeded",
		"ResourceNotFound",
		"ResourceInUse",
		"OptInRequired",
		"InvalidParameterCombination",
		"InvalidParameterValue",
		"InvalidParameterException",
		"InvalidQueryParameter",
		"MalformedQueryString",
		"MissingAction",
		"MissingParameter",
		"ValidationError":
		return true
	default:
		return false
	}
}

// newNodegroupScalingConfig returns a NodegroupScalingConfig object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newNodegroupScalingConfig(
	r *resource,
) *svcsdk.NodegroupScalingConfig {
	res := &svcsdk.NodegroupScalingConfig{}

	if r.ko.Spec.ScalingConfig.DesiredSize != nil {
		res.SetDesiredSize(*r.ko.Spec.ScalingConfig.DesiredSize)
	}
	if r.ko.Spec.ScalingConfig.MaxSize != nil {
		res.SetMaxSize(*r.ko.Spec.ScalingConfig.MaxSize)
	}
	if r.ko.Spec.ScalingConfig.MinSize != nil {
		res.SetMinSize(*r.ko.Spec.ScalingConfig.MinSize)
	}

	return res
}

// newNodegroupUpdateConfig returns a NodegroupUpdateConfig object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newNodegroupUpdateConfig(
	r *resource,
) *svcsdk.NodegroupUpdateConfig {
	res := &svcsdk.NodegroupUpdateConfig{}

	if r.ko.Spec.UpdateConfig.MaxUnavailable != nil {
		res.SetMaxUnavailable(*r.ko.Spec.UpdateConfig.MaxUnavailable)
	}
	if r.ko.Spec.UpdateConfig.MaxUnavailablePercentage != nil {
		res.SetMaxUnavailablePercentage(*r.ko.Spec.UpdateConfig.MaxUnavailablePercentage)
	}

	return res
}
