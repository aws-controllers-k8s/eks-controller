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
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

func compute(enabled *bool, nodePools ...string) *v1alpha1.ComputeConfigRequest {
	c := &v1alpha1.ComputeConfigRequest{Enabled: enabled}
	for _, np := range nodePools {
		c.NodePools = append(c.NodePools, aws.String(np))
	}
	return c
}

func storage(enabled *bool) *v1alpha1.StorageConfigRequest {
	return &v1alpha1.StorageConfigRequest{BlockStorage: &v1alpha1.BlockStorage{Enabled: enabled}}
}

func clusterWithAutoMode(c *v1alpha1.ComputeConfigRequest, s *v1alpha1.StorageConfigRequest, loadBalancing *bool) *resource {
	spec := v1alpha1.ClusterSpec{
		Name:          aws.String("test-cluster"),
		ComputeConfig: c,
		StorageConfig: s,
		KubernetesNetworkConfig: &v1alpha1.KubernetesNetworkConfigRequest{
			IPFamily:        aws.String("ipv4"),
			ServiceIPv4CIDR: aws.String("172.20.0.0/16"),
		},
	}
	if loadBalancing != nil {
		spec.KubernetesNetworkConfig.ElasticLoadBalancing = &v1alpha1.ElasticLoadBalancing{Enabled: loadBalancing}
	}
	return &resource{ko: &v1alpha1.Cluster{Spec: spec}}
}

func TestNewAutoModeUpdateInputSendsUniformTuple(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		desired := clusterWithAutoMode(compute(aws.Bool(enabled)), storage(aws.Bool(enabled)), aws.Bool(enabled))

		input := newAutoModeUpdateInput(desired)

		require.NotNil(t, input.ComputeConfig)
		require.NotNil(t, input.StorageConfig)
		require.NotNil(t, input.StorageConfig.BlockStorage)
		require.NotNil(t, input.KubernetesNetworkConfig)
		require.NotNil(t, input.KubernetesNetworkConfig.ElasticLoadBalancing)

		assert.Equal(t, aws.Bool(enabled), input.ComputeConfig.Enabled)
		assert.Equal(t, aws.Bool(enabled), input.StorageConfig.BlockStorage.Enabled)
		assert.Equal(t, aws.Bool(enabled), input.KubernetesNetworkConfig.ElasticLoadBalancing.Enabled)
	}
}

func TestNewAutoModeUpdateInputOmitsCreateOnlyNetworkConfig(t *testing.T) {
	desired := clusterWithAutoMode(compute(aws.Bool(true)), storage(aws.Bool(true)), aws.Bool(true))

	input := newAutoModeUpdateInput(desired)

	assert.Empty(t, input.KubernetesNetworkConfig.IpFamily)
	assert.Nil(t, input.KubernetesNetworkConfig.ServiceIpv4Cidr)
}

func TestNewAutoModeUpdateInputWithoutNetworkConfig(t *testing.T) {
	desired := clusterWithAutoMode(compute(aws.Bool(true)), storage(aws.Bool(true)), aws.Bool(true))
	desired.ko.Spec.KubernetesNetworkConfig = nil

	input := newAutoModeUpdateInput(desired)

	require.NotNil(t, input.KubernetesNetworkConfig.ElasticLoadBalancing)
	assert.Nil(t, input.KubernetesNetworkConfig.ElasticLoadBalancing.Enabled)
	assert.Empty(t, input.KubernetesNetworkConfig.IpFamily)
	assert.Nil(t, input.KubernetesNetworkConfig.ServiceIpv4Cidr)
}

func TestNewResourceDeltaAutoMode(t *testing.T) {
	tests := []struct {
		name      string
		desired   *resource
		latest    *resource
		different bool
	}{
		{
			"nothing declared, AWS reports the Auto Mode defaults",
			clusterWithAutoMode(nil, nil, nil),
			clusterWithAutoMode(compute(aws.Bool(false)), storage(aws.Bool(false)), aws.Bool(false)),
			false,
		},
		{
			"nothing declared, AWS omits compute and storage",
			clusterWithAutoMode(nil, nil, nil),
			clusterWithAutoMode(nil, nil, aws.Bool(false)),
			false,
		},
		{
			"only storage declared, AWS reports the rest",
			clusterWithAutoMode(nil, storage(aws.Bool(false)), nil),
			clusterWithAutoMode(compute(aws.Bool(false)), storage(aws.Bool(false)), aws.Bool(false)),
			false,
		},
		{
			"load balancing declared disabled, AWS omits it",
			clusterWithAutoMode(nil, nil, aws.Bool(false)),
			clusterWithAutoMode(nil, nil, nil),
			false,
		},
		{
			"compute and storage declared disabled, AWS omits them",
			clusterWithAutoMode(compute(aws.Bool(false)), storage(aws.Bool(false)), aws.Bool(false)),
			clusterWithAutoMode(nil, nil, aws.Bool(false)),
			false,
		},
		{
			"whole tuple declared disabled, AWS omits all of it",
			clusterWithAutoMode(compute(aws.Bool(false)), storage(aws.Bool(false)), aws.Bool(false)),
			clusterWithAutoMode(nil, nil, nil),
			false,
		},
		{
			"nodePools declared on a disabled cluster AWS omits",
			clusterWithAutoMode(compute(aws.Bool(false), "general-purpose"), storage(aws.Bool(false)), aws.Bool(false)),
			clusterWithAutoMode(nil, nil, aws.Bool(false)),
			false,
		},
		{
			"enable requested against a non Auto Mode cluster still differs",
			clusterWithAutoMode(compute(aws.Bool(true)), storage(aws.Bool(true)), aws.Bool(true)),
			clusterWithAutoMode(nil, nil, aws.Bool(false)),
			true,
		},
		{
			"disable requested against an Auto Mode cluster still differs",
			clusterWithAutoMode(compute(aws.Bool(false)), storage(aws.Bool(false)), aws.Bool(false)),
			clusterWithAutoMode(compute(aws.Bool(true)), storage(aws.Bool(true)), aws.Bool(true)),
			true,
		},
		{
			"nodePools change on an Auto Mode cluster still differs",
			clusterWithAutoMode(compute(aws.Bool(true), "general-purpose"), storage(aws.Bool(true)), aws.Bool(true)),
			clusterWithAutoMode(compute(aws.Bool(true), "system", "general-purpose"), storage(aws.Bool(true)), aws.Bool(true)),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, tt.latest)

			autoModeDiff := delta.DifferentAt("Spec.ComputeConfig") ||
				delta.DifferentAt("Spec.StorageConfig") ||
				delta.DifferentAt("Spec.KubernetesNetworkConfig.ElasticLoadBalancing")
			assert.Equal(t, tt.different, autoModeDiff, "differences: %v", delta.Differences)
		})
	}
}

func TestNewResourceDeltaCreateOnlyNetworkConfig(t *testing.T) {
	withNetwork := func(ipFamily, cidr *string) *resource {
		r := clusterWithAutoMode(compute(aws.Bool(false)), storage(aws.Bool(false)), aws.Bool(false))
		r.ko.Spec.KubernetesNetworkConfig.IPFamily = ipFamily
		r.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR = cidr
		return r
	}
	tests := []struct {
		name      string
		desired   *resource
		latest    *resource
		different bool
	}{
		{
			"both omitted, AWS reports what it assigned at creation",
			withNetwork(nil, nil),
			withNetwork(aws.String("ipv4"), aws.String("172.20.0.0/16")),
			false,
		},
		{
			"declared and matching",
			withNetwork(aws.String("ipv4"), aws.String("172.20.0.0/16")),
			withNetwork(aws.String("ipv4"), aws.String("172.20.0.0/16")),
			false,
		},
		{
			"ipFamily changed",
			withNetwork(aws.String("ipv6"), aws.String("172.20.0.0/16")),
			withNetwork(aws.String("ipv4"), aws.String("172.20.0.0/16")),
			true,
		},
		{
			"serviceIPv4CIDR changed",
			withNetwork(aws.String("ipv4"), aws.String("10.100.0.0/16")),
			withNetwork(aws.String("ipv4"), aws.String("172.20.0.0/16")),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, tt.latest)

			createOnlyDiff := delta.DifferentAt("Spec.KubernetesNetworkConfig.IPFamily") ||
				delta.DifferentAt("Spec.KubernetesNetworkConfig.ServiceIPv4CIDR")
			assert.Equal(t, tt.different, createOnlyDiff, "differences: %v", delta.Differences)

			autoModeDiff := delta.DifferentAt("Spec.ComputeConfig") ||
				delta.DifferentAt("Spec.StorageConfig") ||
				delta.DifferentAt("Spec.KubernetesNetworkConfig.ElasticLoadBalancing")
			assert.False(t, autoModeDiff, "create-only drift must not dispatch an Auto Mode update")
		})
	}
}

func TestNewAutoModeUpdateInputWithoutComputeConfig(t *testing.T) {
	desired := clusterWithAutoMode(nil, storage(aws.Bool(false)), aws.Bool(false))

	input := newAutoModeUpdateInput(desired)

	require.NotNil(t, input.ComputeConfig)
	assert.Nil(t, input.ComputeConfig.Enabled)
	assert.Nil(t, input.ComputeConfig.NodePools)
	assert.Nil(t, input.ComputeConfig.NodeRoleArn)
}

func TestAutoModeRejected(t *testing.T) {
	tupleMismatch := &smithy.GenericAPIError{
		Code:    "InvalidParameterException",
		Message: "For EKS Auto Mode, please ensure that all required configs, including computeConfig, kubernetesNetworkConfig, and blockStorage are all either fully enabled or fully disabled.",
	}
	notEnabled := &smithy.GenericAPIError{
		Code:    "InvalidRequestException",
		Message: "Cannot modify EKS Auto Mode configuration. Auto Mode is not enabled on this cluster.",
	}
	tupleIncomplete := &smithy.GenericAPIError{
		Code:    "InvalidParameterException",
		Message: "The type for cluster update was not provided.",
	}
	rolePropagating := &smithy.GenericAPIError{
		Code:    "ClientException",
		Message: "The provided role doesn't have the Amazon EKS Managed Policies associated with it.",
	}
	subnetPropagating := &smithy.GenericAPIError{
		Code:    "InvalidParameterException",
		Message: "Subnet subnet-0abc is not currently available.",
	}

	assert.True(t, autoModeRejected(fmt.Errorf("failed to update AutoMode config: %w", tupleMismatch)))
	assert.False(t, autoModeRejected(fmt.Errorf("wrapped: %w", notEnabled)),
		"the compare hook must stop this call from being made; terminal would strand a resource the user cannot fix")
	assert.False(t, autoModeRejected(fmt.Errorf("wrapped: %w", tupleIncomplete)),
		"an incomplete tuple is a controller bug, not a spec the user can correct")
	assert.False(t, autoModeRejected(fmt.Errorf("wrapped: %w", rolePropagating)),
		"IAM propagation is retriable and must stay recoverable")
	assert.False(t, autoModeRejected(fmt.Errorf("wrapped: %w", subnetPropagating)),
		"InvalidParameterException also covers retriable propagation and must stay recoverable")
	assert.False(t, autoModeRejected(errors.New("boom")))
}

func TestNewResourceDeltaAutoModeAdopted(t *testing.T) {
	enabledCluster := func() *resource {
		return clusterWithAutoMode(
			compute(aws.Bool(true), "general-purpose", "system"),
			storage(aws.Bool(true)),
			aws.Bool(true),
		)
	}
	tests := []struct {
		name      string
		desired   *resource
		different bool
	}{
		{"nothing declared against an enabled cluster", clusterWithAutoMode(nil, nil, nil), true},
		{"only compute declared against an enabled cluster", clusterWithAutoMode(compute(aws.Bool(true)), nil, nil), true},
		{"whole tuple declared matching an enabled cluster", enabledCluster(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delta := newResourceDelta(tt.desired, enabledCluster())

			autoModeDiff := delta.DifferentAt("Spec.ComputeConfig") ||
				delta.DifferentAt("Spec.StorageConfig") ||
				delta.DifferentAt("Spec.KubernetesNetworkConfig.ElasticLoadBalancing")
			assert.Equal(t, tt.different, autoModeDiff, "differences: %v", delta.Differences)
		})
	}
}

func TestNewResourceDeltaDoesNotMutateDesired(t *testing.T) {
	desired := clusterWithAutoMode(nil, nil, nil)
	desired.ko.Spec.KubernetesNetworkConfig.IPFamily = nil
	desired.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR = nil
	before := desired.ko.Spec.DeepCopy()

	newResourceDelta(desired, clusterWithAutoMode(
		compute(aws.Bool(true), "general-purpose"), storage(aws.Bool(true)), aws.Bool(true),
	))

	assert.Equal(t, before, desired.ko.Spec.DeepCopy(),
		"the compare hook must not write observed state into desired; customUpdate builds the AWS request from it")
}

func TestAutoModeTupleDeclared(t *testing.T) {
	full := clusterWithAutoMode(compute(aws.Bool(true)), storage(aws.Bool(true)), aws.Bool(true))
	assert.True(t, autoModeTupleDeclared(full))

	assert.False(t, autoModeTupleDeclared(clusterWithAutoMode(nil, nil, nil)))
	assert.False(t, autoModeTupleDeclared(clusterWithAutoMode(compute(aws.Bool(true)), storage(aws.Bool(true)), nil)))
	assert.False(t, autoModeTupleDeclared(clusterWithAutoMode(compute(nil), storage(aws.Bool(true)), aws.Bool(true))))
	assert.False(t, autoModeTupleDeclared(clusterWithAutoMode(compute(aws.Bool(true)), &v1alpha1.StorageConfigRequest{}, aws.Bool(true))))
}
