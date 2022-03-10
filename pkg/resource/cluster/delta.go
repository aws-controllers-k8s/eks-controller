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

package cluster

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

	if ackcompare.HasNilDifference(a.ko.Spec.ClientRequestToken, b.ko.Spec.ClientRequestToken) {
		delta.Add("Spec.ClientRequestToken", a.ko.Spec.ClientRequestToken, b.ko.Spec.ClientRequestToken)
	} else if a.ko.Spec.ClientRequestToken != nil && b.ko.Spec.ClientRequestToken != nil {
		if *a.ko.Spec.ClientRequestToken != *b.ko.Spec.ClientRequestToken {
			delta.Add("Spec.ClientRequestToken", a.ko.Spec.ClientRequestToken, b.ko.Spec.ClientRequestToken)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.EncryptionConfig, b.ko.Spec.EncryptionConfig) {
		delta.Add("Spec.EncryptionConfig", a.ko.Spec.EncryptionConfig, b.ko.Spec.EncryptionConfig)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.KubernetesNetworkConfig, b.ko.Spec.KubernetesNetworkConfig) {
		delta.Add("Spec.KubernetesNetworkConfig", a.ko.Spec.KubernetesNetworkConfig, b.ko.Spec.KubernetesNetworkConfig)
	} else if a.ko.Spec.KubernetesNetworkConfig != nil && b.ko.Spec.KubernetesNetworkConfig != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.KubernetesNetworkConfig.IPFamily, b.ko.Spec.KubernetesNetworkConfig.IPFamily) {
			delta.Add("Spec.KubernetesNetworkConfig.IPFamily", a.ko.Spec.KubernetesNetworkConfig.IPFamily, b.ko.Spec.KubernetesNetworkConfig.IPFamily)
		} else if a.ko.Spec.KubernetesNetworkConfig.IPFamily != nil && b.ko.Spec.KubernetesNetworkConfig.IPFamily != nil {
			if *a.ko.Spec.KubernetesNetworkConfig.IPFamily != *b.ko.Spec.KubernetesNetworkConfig.IPFamily {
				delta.Add("Spec.KubernetesNetworkConfig.IPFamily", a.ko.Spec.KubernetesNetworkConfig.IPFamily, b.ko.Spec.KubernetesNetworkConfig.IPFamily)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR, b.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR) {
			delta.Add("Spec.KubernetesNetworkConfig.ServiceIPv4CIDR", a.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR, b.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR)
		} else if a.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR != nil && b.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR != nil {
			if *a.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR != *b.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR {
				delta.Add("Spec.KubernetesNetworkConfig.ServiceIPv4CIDR", a.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR, b.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Logging, b.ko.Spec.Logging) {
		delta.Add("Spec.Logging", a.ko.Spec.Logging, b.ko.Spec.Logging)
	} else if a.ko.Spec.Logging != nil && b.ko.Spec.Logging != nil {
		if !reflect.DeepEqual(a.ko.Spec.Logging.ClusterLogging, b.ko.Spec.Logging.ClusterLogging) {
			delta.Add("Spec.Logging.ClusterLogging", a.ko.Spec.Logging.ClusterLogging, b.ko.Spec.Logging.ClusterLogging)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Name, b.ko.Spec.Name) {
		delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
	} else if a.ko.Spec.Name != nil && b.ko.Spec.Name != nil {
		if *a.ko.Spec.Name != *b.ko.Spec.Name {
			delta.Add("Spec.Name", a.ko.Spec.Name, b.ko.Spec.Name)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.ResourcesVPCConfig, b.ko.Spec.ResourcesVPCConfig) {
		delta.Add("Spec.ResourcesVPCConfig", a.ko.Spec.ResourcesVPCConfig, b.ko.Spec.ResourcesVPCConfig)
	} else if a.ko.Spec.ResourcesVPCConfig != nil && b.ko.Spec.ResourcesVPCConfig != nil {
		if ackcompare.HasNilDifference(a.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess, b.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess) {
			delta.Add("Spec.ResourcesVPCConfig.EndpointPrivateAccess", a.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess, b.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess)
		} else if a.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess != nil && b.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess != nil {
			if *a.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess != *b.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess {
				delta.Add("Spec.ResourcesVPCConfig.EndpointPrivateAccess", a.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess, b.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess)
			}
		}
		if ackcompare.HasNilDifference(a.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess, b.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess) {
			delta.Add("Spec.ResourcesVPCConfig.EndpointPublicAccess", a.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess, b.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess)
		} else if a.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess != nil && b.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess != nil {
			if *a.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess != *b.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess {
				delta.Add("Spec.ResourcesVPCConfig.EndpointPublicAccess", a.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess, b.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess)
			}
		}
		if !ackcompare.SliceStringPEqual(a.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs, b.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs) {
			delta.Add("Spec.ResourcesVPCConfig.PublicAccessCIDRs", a.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs, b.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs)
		}
		if !ackcompare.SliceStringPEqual(a.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs, b.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs) {
			delta.Add("Spec.ResourcesVPCConfig.SecurityGroupIDs", a.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs, b.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs)
		}
		if !ackcompare.SliceStringPEqual(a.ko.Spec.ResourcesVPCConfig.SubnetIDs, b.ko.Spec.ResourcesVPCConfig.SubnetIDs) {
			delta.Add("Spec.ResourcesVPCConfig.SubnetIDs", a.ko.Spec.ResourcesVPCConfig.SubnetIDs, b.ko.Spec.ResourcesVPCConfig.SubnetIDs)
		}
	}
	if ackcompare.HasNilDifference(a.ko.Spec.RoleARN, b.ko.Spec.RoleARN) {
		delta.Add("Spec.RoleARN", a.ko.Spec.RoleARN, b.ko.Spec.RoleARN)
	} else if a.ko.Spec.RoleARN != nil && b.ko.Spec.RoleARN != nil {
		if *a.ko.Spec.RoleARN != *b.ko.Spec.RoleARN {
			delta.Add("Spec.RoleARN", a.ko.Spec.RoleARN, b.ko.Spec.RoleARN)
		}
	}
	if !reflect.DeepEqual(a.ko.Spec.RoleRef, b.ko.Spec.RoleRef) {
		delta.Add("Spec.RoleRef", a.ko.Spec.RoleRef, b.ko.Spec.RoleRef)
	}
	if ackcompare.HasNilDifference(a.ko.Spec.Tags, b.ko.Spec.Tags) {
		delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
	} else if a.ko.Spec.Tags != nil && b.ko.Spec.Tags != nil {
		if !ackcompare.MapStringStringPEqual(a.ko.Spec.Tags, b.ko.Spec.Tags) {
			delta.Add("Spec.Tags", a.ko.Spec.Tags, b.ko.Spec.Tags)
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
