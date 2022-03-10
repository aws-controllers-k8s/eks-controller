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
	"bytes"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
)

// Hack to avoid import errors during build...
var (
	_ = &bytes.Buffer{}
	_ = &reflect.Method{}
)

// newResourceDelta returns a new `ackcompare.Delta` used to compare two
// resources
func newResourceDelta(
	a *resource,
	b *resource,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if (a == nil && b != nil) ||
		(a != nil && b == nil) {
		delta.Add("", a, b)
		return delta
	}
	customPreCompare(a, b)

	if ackcompare.HasNilDifference(a.ko.Spec.AMIType, b.ko.Spec.AMIType) {
		delta.Add("Spec.AMIType", a.ko.Spec.AMIType, b.ko.Spec.AMIType)
	} else if a.ko.Spec.AMIType != nil && b.ko.Spec.AMIType != nil {
		if *a.ko.Spec.AMIType != *b.ko.Spec.AMIType {
			delta.Add("Spec.AMIType", a.ko.Spec.AMIType, b.ko.Spec.AMIType)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.CapacityType, b.ko.Spec.CapacityType) {
		delta.Add("Spec.CapacityType", a.ko.Spec.CapacityType, b.ko.Spec.CapacityType)
	} else if a.ko.Spec.CapacityType != nil && b.ko.Spec.CapacityType != nil {
		if *a.ko.Spec.CapacityType != *b.ko.Spec.CapacityType {
			delta.Add("Spec.CapacityType", a.ko.Spec.CapacityType, b.ko.Spec.CapacityType)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ClientRequestToken, b.ko.Spec.ClientRequestToken) {
		delta.Add("Spec.ClientRequestToken", a.ko.Spec.ClientRequestToken, b.ko.Spec.ClientRequestToken)
	} else if a.ko.Spec.ClientRequestToken != nil && b.ko.Spec.ClientRequestToken != nil {
		if *a.ko.Spec.ClientRequestToken != *b.ko.Spec.ClientRequestToken {
			delta.Add("Spec.ClientRequestToken", a.ko.Spec.ClientRequestToken, b.ko.Spec.ClientRequestToken)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ClusterName, b.ko.Spec.ClusterName) {
		delta.Add("Spec.ClusterName", a.ko.Spec.ClusterName, b.ko.Spec.ClusterName)
	} else if a.ko.Spec.ClusterName != nil && b.ko.Spec.ClusterName != nil {
		if *a.ko.Spec.ClusterName != *b.ko.Spec.ClusterName {
			delta.Add("Spec.ClusterName", a.ko.Spec.ClusterName, b.ko.Spec.ClusterName)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.ClusterRef, b.ko.Spec.ClusterRef) {
		delta.Add("Spec.ClusterRef", a.ko.Spec.ClusterRef, b.ko.Spec.ClusterRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.DiskSize, b.ko.Spec.DiskSize) {
		delta.Add("Spec.DiskSize", a.ko.Spec.DiskSize, b.ko.Spec.DiskSize)
	} else if a.ko.Spec.DiskSize != nil && b.ko.Spec.DiskSize != nil {
		if *a.ko.Spec.DiskSize != *b.ko.Spec.DiskSize {
			delta.Add("Spec.DiskSize", a.ko.Spec.DiskSize, b.ko.Spec.DiskSize)
		}
	}
	if !ackcompare.SliceStringPEqual(a.ko.Spec.InstanceTypes, b.ko.Spec.InstanceTypes) {
		delta.Add("Spec.InstanceTypes", a.ko.Spec.InstanceTypes, b.ko.Spec.InstanceTypes)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Labels, b.ko.Spec.Labels) {
		delta.Add("Spec.Labels", a.ko.Spec.Labels, b.ko.Spec.Labels)
	} else if a.ko.Spec.Labels != nil && b.ko.Spec.Labels != nil {
		if !ackcompare.MapStringStringPEqual(a.ko.Spec.Labels, b.ko.Spec.Labels) {
			delta.Add("Spec.Labels", a.ko.Spec.Labels, b.ko.Spec.Labels)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.LaunchTemplate, b.ko.Spec.LaunchTemplate) {
		delta.Add("Spec.LaunchTemplate", a.ko.Spec.LaunchTemplate, b.ko.Spec.LaunchTemplate)
	} else if a.ko.Spec.LaunchTemplate != nil && b.ko.Spec.LaunchTemplate != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.LaunchTemplate.ID, b.ko.Spec.LaunchTemplate.ID) {
			delta.Add("Spec.LaunchTemplate.ID", a.ko.Spec.LaunchTemplate.ID, b.ko.Spec.LaunchTemplate.ID)
		} else if a.ko.Spec.LaunchTemplate.ID != nil && b.ko.Spec.LaunchTemplate.ID != nil {
			if *a.ko.Spec.LaunchTemplate.ID != *b.ko.Spec.LaunchTemplate.ID {
				delta.Add("Spec.LaunchTemplate.ID", a.ko.Spec.LaunchTemplate.ID, b.ko.Spec.LaunchTemplate.ID)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.LaunchTemplate.Name, b.ko.Spec.LaunchTemplate.Name) {
			delta.Add("Spec.LaunchTemplate.Name", a.ko.Spec.LaunchTemplate.Name, b.ko.Spec.LaunchTemplate.Name)
		} else if a.ko.Spec.LaunchTemplate.Name != nil && b.ko.Spec.LaunchTemplate.Name != nil {
			if *a.ko.Spec.LaunchTemplate.Name != *b.ko.Spec.LaunchTemplate.Name {
				delta.Add("Spec.LaunchTemplate.Name", a.ko.Spec.LaunchTemplate.Name, b.ko.Spec.LaunchTemplate.Name)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.LaunchTemplate.Version, b.ko.Spec.LaunchTemplate.Version) {
			delta.Add("Spec.LaunchTemplate.Version", a.ko.Spec.LaunchTemplate.Version, b.ko.Spec.LaunchTemplate.Version)
		} else if a.ko.Spec.LaunchTemplate.Version != nil && b.ko.Spec.LaunchTemplate.Version != nil {
			if *a.ko.Spec.LaunchTemplate.Version != *b.ko.Spec.LaunchTemplate.Version {
				delta.Add("Spec.LaunchTemplate.Version", a.ko.Spec.LaunchTemplate.Version, b.ko.Spec.LaunchTemplate.Version)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Name, b.ko.Spec.Name) {
		delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
	} else if a.ko.Spec.Name != nil && b.ko.Spec.Name != nil {
		if *a.ko.Spec.Name != *b.ko.Spec.Name {
			delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.NodeRole, b.ko.Spec.NodeRole) {
		delta.Add("Spec.NodeRole", a.ko.Spec.NodeRole, b.ko.Spec.NodeRole)
	} else if a.ko.Spec.NodeRole != nil && b.ko.Spec.NodeRole != nil {
		if *a.ko.Spec.NodeRole != *b.ko.Spec.NodeRole {
			delta.Add("Spec.NodeRole", a.ko.Spec.NodeRole, b.ko.Spec.NodeRole)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.NodeRoleRef, b.ko.Spec.NodeRoleRef) {
		delta.Add("Spec.NodeRoleRef", a.ko.Spec.NodeRoleRef, b.ko.Spec.NodeRoleRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ReleaseVersion, b.ko.Spec.ReleaseVersion) {
		delta.Add("Spec.ReleaseVersion", a.ko.Spec.ReleaseVersion, b.ko.Spec.ReleaseVersion)
	} else if a.ko.Spec.ReleaseVersion != nil && b.ko.Spec.ReleaseVersion != nil {
		if *a.ko.Spec.ReleaseVersion != *b.ko.Spec.ReleaseVersion {
			delta.Add("Spec.ReleaseVersion", a.ko.Spec.ReleaseVersion, b.ko.Spec.ReleaseVersion)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.RemoteAccess, b.ko.Spec.RemoteAccess) {
		delta.Add("Spec.RemoteAccess", a.ko.Spec.RemoteAccess, b.ko.Spec.RemoteAccess)
	} else if a.ko.Spec.RemoteAccess != nil && b.ko.Spec.RemoteAccess != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.RemoteAccess.EC2SshKey, b.ko.Spec.RemoteAccess.EC2SshKey) {
			delta.Add("Spec.RemoteAccess.EC2SshKey", a.ko.Spec.RemoteAccess.EC2SshKey, b.ko.Spec.RemoteAccess.EC2SshKey)
		} else if a.ko.Spec.RemoteAccess.EC2SshKey != nil && b.ko.Spec.RemoteAccess.EC2SshKey != nil {
			if *a.ko.Spec.RemoteAccess.EC2SshKey != *b.ko.Spec.RemoteAccess.EC2SshKey {
				delta.Add("Spec.RemoteAccess.EC2SshKey", a.ko.Spec.RemoteAccess.EC2SshKey, b.ko.Spec.RemoteAccess.EC2SshKey)
			}
		}
		if !ackcompare.SliceStringPEqual(a.ko.Spec.RemoteAccess.SourceSecurityGroups, b.ko.Spec.RemoteAccess.SourceSecurityGroups) {
			delta.Add("Spec.RemoteAccess.SourceSecurityGroups", a.ko.Spec.RemoteAccess.SourceSecurityGroups, b.ko.Spec.RemoteAccess.SourceSecurityGroups)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ScalingConfig, b.ko.Spec.ScalingConfig) {
		delta.Add("Spec.ScalingConfig", a.ko.Spec.ScalingConfig, b.ko.Spec.ScalingConfig)
	} else if a.ko.Spec.ScalingConfig != nil && b.ko.Spec.ScalingConfig != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.ScalingConfig.DesiredSize, b.ko.Spec.ScalingConfig.DesiredSize) {
			delta.Add("Spec.ScalingConfig.DesiredSize", a.ko.Spec.ScalingConfig.DesiredSize, b.ko.Spec.ScalingConfig.DesiredSize)
		} else if a.ko.Spec.ScalingConfig.DesiredSize != nil && b.ko.Spec.ScalingConfig.DesiredSize != nil {
			if *a.ko.Spec.ScalingConfig.DesiredSize != *b.ko.Spec.ScalingConfig.DesiredSize {
				delta.Add("Spec.ScalingConfig.DesiredSize", a.ko.Spec.ScalingConfig.DesiredSize, b.ko.Spec.ScalingConfig.DesiredSize)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.ScalingConfig.MaxSize, b.ko.Spec.ScalingConfig.MaxSize) {
			delta.Add("Spec.ScalingConfig.MaxSize", a.ko.Spec.ScalingConfig.MaxSize, b.ko.Spec.ScalingConfig.MaxSize)
		} else if a.ko.Spec.ScalingConfig.MaxSize != nil && b.ko.Spec.ScalingConfig.MaxSize != nil {
			if *a.ko.Spec.ScalingConfig.MaxSize != *b.ko.Spec.ScalingConfig.MaxSize {
				delta.Add("Spec.ScalingConfig.MaxSize", a.ko.Spec.ScalingConfig.MaxSize, b.ko.Spec.ScalingConfig.MaxSize)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.ScalingConfig.MinSize, b.ko.Spec.ScalingConfig.MinSize) {
			delta.Add("Spec.ScalingConfig.MinSize", a.ko.Spec.ScalingConfig.MinSize, b.ko.Spec.ScalingConfig.MinSize)
		} else if a.ko.Spec.ScalingConfig.MinSize != nil && b.ko.Spec.ScalingConfig.MinSize != nil {
			if *a.ko.Spec.ScalingConfig.MinSize != *b.ko.Spec.ScalingConfig.MinSize {
				delta.Add("Spec.ScalingConfig.MinSize", a.ko.Spec.ScalingConfig.MinSize, b.ko.Spec.ScalingConfig.MinSize)
			}
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.SubnetRefs, b.ko.Spec.SubnetRefs) {
		delta.Add("Spec.SubnetRefs", a.ko.Spec.SubnetRefs, b.ko.Spec.SubnetRefs)
	}
	if !ackcompare.SliceStringPEqual(a.ko.Spec.Subnets, b.ko.Spec.Subnets) {
		delta.Add("Spec.Subnets", a.ko.Spec.Subnets, b.ko.Spec.Subnets)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Tags, b.ko.Spec.Tags) {
		delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
	} else if a.ko.Spec.Tags != nil && b.ko.Spec.Tags != nil {
		if !ackcompare.MapStringStringPEqual(a.ko.Spec.Tags, b.ko.Spec.Tags) {
			delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.Taints, b.ko.Spec.Taints) {
		delta.Add("Spec.Taints", a.ko.Spec.Taints, b.ko.Spec.Taints)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.UpdateConfig, b.ko.Spec.UpdateConfig) {
		delta.Add("Spec.UpdateConfig", a.ko.Spec.UpdateConfig, b.ko.Spec.UpdateConfig)
	} else if a.ko.Spec.UpdateConfig != nil && b.ko.Spec.UpdateConfig != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.UpdateConfig.MaxUnavailable, b.ko.Spec.UpdateConfig.MaxUnavailable) {
			delta.Add("Spec.UpdateConfig.MaxUnavailable", a.ko.Spec.UpdateConfig.MaxUnavailable, b.ko.Spec.UpdateConfig.MaxUnavailable)
		} else if a.ko.Spec.UpdateConfig.MaxUnavailable != nil && b.ko.Spec.UpdateConfig.MaxUnavailable != nil {
			if *a.ko.Spec.UpdateConfig.MaxUnavailable != *b.ko.Spec.UpdateConfig.MaxUnavailable {
				delta.Add("Spec.UpdateConfig.MaxUnavailable", a.ko.Spec.UpdateConfig.MaxUnavailable, b.ko.Spec.UpdateConfig.MaxUnavailable)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.UpdateConfig.MaxUnavailablePercentage, b.ko.Spec.UpdateConfig.MaxUnavailablePercentage) {
			delta.Add("Spec.UpdateConfig.MaxUnavailablePercentage", a.ko.Spec.UpdateConfig.MaxUnavailablePercentage, b.ko.Spec.UpdateConfig.MaxUnavailablePercentage)
		} else if a.ko.Spec.UpdateConfig.MaxUnavailablePercentage != nil && b.ko.Spec.UpdateConfig.MaxUnavailablePercentage != nil {
			if *a.ko.Spec.UpdateConfig.MaxUnavailablePercentage != *b.ko.Spec.UpdateConfig.MaxUnavailablePercentage {
				delta.Add("Spec.UpdateConfig.MaxUnavailablePercentage", a.ko.Spec.UpdateConfig.MaxUnavailablePercentage, b.ko.Spec.UpdateConfig.MaxUnavailablePercentage)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Version, b.ko.Spec.Version) {
		delta.Add("Spec.Version", a.ko.Spec.Version, b.ko.Spec.Version)
	} else if a.ko.Spec.Version != nil && b.ko.Spec.Version != nil {
		if *a.ko.Spec.Version != *b.ko.Spec.Version {
			delta.Add("Spec.Version", a.ko.Spec.Version, b.ko.Spec.Version)
		}
	}

	return delta
}
