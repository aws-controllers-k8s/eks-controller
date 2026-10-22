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

package capability

import (
	"testing"

	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

func capabilityWithConfiguration(cfg *v1alpha1.CapabilityConfigurationRequest) *resource {
	return &resource{ko: &v1alpha1.Capability{
		Spec: v1alpha1.CapabilitySpec{
			Name:          aws.String("test-capability"),
			ClusterName:   aws.String("test-cluster"),
			Configuration: cfg,
		},
	}}
}

func ackConfig(enableCrossNamespace *bool, disabledServices ...string) *v1alpha1.AckConfigRequest {
	c := &v1alpha1.AckConfigRequest{EnableCrossNamespace: enableCrossNamespace}
	for _, svc := range disabledServices {
		c.DisabledServices = append(c.DisabledServices, aws.String(svc))
	}
	return c
}

func argoCDConfig(vpceIDs ...string) *v1alpha1.ArgoCDConfigRequest {
	c := &v1alpha1.ArgoCDConfigRequest{
		NetworkAccess: &v1alpha1.ArgoCDNetworkAccessConfigRequest{},
	}
	for _, id := range vpceIDs {
		c.NetworkAccess.VPCEIDs = append(c.NetworkAccess.VPCEIDs, aws.String(id))
	}
	return c
}

// TestSetConfigurationSendsDeclaredAckConfig is the regression test for an ACK
// capability losing its whole configuration on update. Before the ack branch
// existed, any capability without an argoCD block had input.Configuration set
// to nil, so the update silently dropped every declared ACK setting.
func TestSetConfigurationSendsDeclaredAckConfig(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(true), "s3", "ec2"),
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(false)),
	})

	input := &svcsdk.UpdateCapabilityInput{}
	setConfiguration(input, desired, latest)

	require.NotNil(t, input.Configuration)
	require.NotNil(t, input.Configuration.Ack)
	assert.Nil(t, input.Configuration.ArgoCd)
	assert.Equal(t, aws.Bool(true), input.Configuration.Ack.EnableCrossNamespace)
	assert.Equal(t, []string{"s3", "ec2"}, input.Configuration.Ack.DisabledServices)
}

// TestSetConfigurationAckOmittedDisabledServicesIsLeftAlone covers adoption.
// Under the adopt-or-create policy the spec is taken as authored, so a block
// declaring only enableCrossNamespace must not wipe the services an adopted
// capability already had disabled. UpdateAckConfig reads an omitted field as
// "leave unchanged", and late initialization fills the field in instead.
func TestSetConfigurationAckOmittedDisabledServicesIsLeftAlone(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(true)),
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(true), "s3"),
	})

	input := &svcsdk.UpdateCapabilityInput{}
	setConfiguration(input, desired, latest)

	require.NotNil(t, input.Configuration)
	require.NotNil(t, input.Configuration.Ack)
	assert.Nil(t, input.Configuration.Ack.DisabledServices)
}

// TestSetConfigurationAckEmptyDisabledServicesReEnablesEverything keeps the
// explicit empty list distinguishable from an omitted one. An empty list is the
// documented way to turn every service back on, so it has to reach the API as
// an empty, non-nil slice.
func TestSetConfigurationAckEmptyDisabledServicesReEnablesEverything(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: &v1alpha1.AckConfigRequest{DisabledServices: []*string{}},
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(nil, "s3"),
	})

	input := &svcsdk.UpdateCapabilityInput{}
	setConfiguration(input, desired, latest)

	require.NotNil(t, input.Configuration.Ack)
	require.NotNil(t, input.Configuration.Ack.DisabledServices)
	assert.Empty(t, input.Configuration.Ack.DisabledServices)
}

// TestSetConfigurationAckOmitsUndeclaredCrossNamespace keeps the field out of
// the request when the spec never declared it, because UpdateAckConfig reads an
// omitted field as "leave unchanged".
func TestSetConfigurationAckOmitsUndeclaredCrossNamespace(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(nil, "iam"),
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(true)),
	})

	input := &svcsdk.UpdateCapabilityInput{}
	setConfiguration(input, desired, latest)

	require.NotNil(t, input.Configuration.Ack)
	assert.Nil(t, input.Configuration.Ack.EnableCrossNamespace)
	assert.Equal(t, []string{"iam"}, input.Configuration.Ack.DisabledServices)
}

// TestSetConfigurationArgoCDUnchanged pins the pre-existing Argo CD behavior so
// the ack branch cannot regress it.
func TestSetConfigurationArgoCDUnchanged(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		ArgoCD: argoCDConfig("vpce-123"),
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		ArgoCD: argoCDConfig(),
	})

	input := &svcsdk.UpdateCapabilityInput{}
	setConfiguration(input, desired, latest)

	require.NotNil(t, input.Configuration)
	require.NotNil(t, input.Configuration.ArgoCd)
	assert.Nil(t, input.Configuration.Ack)
	require.NotNil(t, input.Configuration.ArgoCd.NetworkAccess)
	assert.Equal(t, []string{"vpce-123"}, input.Configuration.ArgoCd.NetworkAccess.VpceIds)
}

// TestSetConfigurationArgoCDWithUnobservedConfiguration guards the RBAC
// comparison against a latest resource that carries no configuration, which
// happens on the first update after an adoption.
func TestSetConfigurationArgoCDWithUnobservedConfiguration(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		ArgoCD: &v1alpha1.ArgoCDConfigRequest{
			RbacRoleMappings: []*v1alpha1.ArgoCDRoleMapping{{Role: aws.String("admin")}},
		},
	})
	latest := capabilityWithConfiguration(nil)

	input := &svcsdk.UpdateCapabilityInput{}
	require.NotPanics(t, func() { setConfiguration(input, desired, latest) })

	require.NotNil(t, input.Configuration.ArgoCd)
	require.NotNil(t, input.Configuration.ArgoCd.RbacRoleMappings)
	assert.Len(t, input.Configuration.ArgoCd.RbacRoleMappings.AddOrUpdateRoleMappings, 1)
	assert.Empty(t, input.Configuration.ArgoCd.RbacRoleMappings.RemoveRoleMappings)
}

// TestSetConfigurationBothBlocks is defensive. The API models the two blocks as
// independent members, so a spec carrying both must not have either dropped.
func TestSetConfigurationBothBlocks(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack:    ackConfig(aws.Bool(true), "s3"),
		ArgoCD: argoCDConfig("vpce-123"),
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{})

	input := &svcsdk.UpdateCapabilityInput{}
	setConfiguration(input, desired, latest)

	require.NotNil(t, input.Configuration.Ack)
	require.NotNil(t, input.Configuration.ArgoCd)
	assert.Equal(t, []string{"s3"}, input.Configuration.Ack.DisabledServices)
	assert.Equal(t, []string{"vpce-123"}, input.Configuration.ArgoCd.NetworkAccess.VpceIds)
}

// TestUndeclaredConfigurationDoesNotDriveUpdate covers the shape the SDK bump
// introduces. DescribeCapability reports the ack block EKS resolved for an ACK
// capability, including enableCrossNamespace=false when it was never set, so a
// spec that declares no configuration reads back as a difference on every
// resync. The update path has nothing to send for it, so it must not call
// UpdateCapability on the strength of that difference alone.
func TestUndeclaredConfigurationDoesNotDriveUpdate(t *testing.T) {
	desired := capabilityWithConfiguration(nil)
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(false)),
	})

	delta := newResourceDelta(desired, latest)
	require.True(t, delta.DifferentAt("Spec.Configuration"),
		"expected the observed ack block to register as a difference")
	assert.False(t, differsOutsideConfiguration(delta))
}

// TestDeclaredConfigurationDrivesUpdate is the other half: a declared
// configuration that drifts still has to reach UpdateCapability, and the gate
// only applies when the spec declares none.
func TestDeclaredConfigurationDrivesUpdate(t *testing.T) {
	desired := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(true), "s3"),
	})
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(false)),
	})

	delta := newResourceDelta(desired, latest)
	require.True(t, delta.DifferentAt("Spec.Configuration.Ack.EnableCrossNamespace"))
	assert.NotNil(t, desired.ko.Spec.Configuration)
}

// TestDiffersOutsideConfigurationSeesOtherFields guards the gate itself: a real
// change alongside an undeclared configuration must still reach the API.
func TestDiffersOutsideConfigurationSeesOtherFields(t *testing.T) {
	desired := capabilityWithConfiguration(nil)
	desired.ko.Spec.RoleARN = aws.String("arn:aws:iam::123456789012:role/desired")
	latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack: ackConfig(aws.Bool(false)),
	})
	latest.ko.Spec.RoleARN = aws.String("arn:aws:iam::123456789012:role/latest")

	delta := newResourceDelta(desired, latest)
	assert.True(t, differsOutsideConfiguration(delta))
}

func TestSetConfigurationWithoutDeclaredConfiguration(t *testing.T) {
	for name, cfg := range map[string]*v1alpha1.CapabilityConfigurationRequest{
		"nil configuration":   nil,
		"empty configuration": {},
	} {
		t.Run(name, func(t *testing.T) {
			desired := capabilityWithConfiguration(cfg)
			latest := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
				Ack: ackConfig(aws.Bool(true)),
			})

			input := &svcsdk.UpdateCapabilityInput{}
			setConfiguration(input, desired, latest)

			assert.Nil(t, input.Configuration)
		})
	}
}

func capabilityWithStatus(status *string) *resource {
	return &resource{ko: &v1alpha1.Capability{
		Status: v1alpha1.CapabilityStatus{Status: status},
	}}
}

// TestValidateConfigurationRejectsBothBlocks covers the only spec the EKS API
// cannot act on coherently. A capability has one type, so declaring both
// configuration blocks is a user error the controller should surface as
// terminal rather than retry.
func TestValidateConfigurationRejectsBothBlocks(t *testing.T) {
	r := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
		Ack:    ackConfig(aws.Bool(true), "s3"),
		ArgoCD: argoCDConfig("vpce-1"),
	})

	err := validateConfiguration(r)

	require.Error(t, err)
	var terminal *ackerr.TerminalError
	require.ErrorAs(t, err, &terminal)
}

func TestValidateConfigurationAcceptsOneBlockOrNone(t *testing.T) {
	t.Run("ack only", func(t *testing.T) {
		r := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
			Ack: ackConfig(aws.Bool(true), "s3"),
		})
		assert.NoError(t, validateConfiguration(r))
	})
	t.Run("argoCD only", func(t *testing.T) {
		r := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{
			ArgoCD: argoCDConfig("vpce-1"),
		})
		assert.NoError(t, validateConfiguration(r))
	})
	t.Run("empty configuration", func(t *testing.T) {
		r := capabilityWithConfiguration(&v1alpha1.CapabilityConfigurationRequest{})
		assert.NoError(t, validateConfiguration(r))
	})
	t.Run("no configuration", func(t *testing.T) {
		assert.NoError(t, validateConfiguration(capabilityWithConfiguration(nil)))
	})
}

// TestCapabilityInFlight pins which statuses hold the update back. The failed
// and degraded statuses have to stay out: EKS is not working on the capability
// in those states, so gating on them would requeue with nothing to wait for.
func TestCapabilityInFlight(t *testing.T) {
	for _, tc := range []struct {
		status   *string
		inFlight bool
	}{
		{aws.String("CREATING"), true},
		{aws.String("UPDATING"), true},
		{aws.String("DELETING"), true},
		{aws.String("ACTIVE"), false},
		{aws.String("DEGRADED"), false},
		{aws.String("CREATE_FAILED"), false},
		{aws.String("DELETE_FAILED"), false},
		{nil, false},
	} {
		name := "nil"
		if tc.status != nil {
			name = *tc.status
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.inFlight, capabilityInFlight(capabilityWithStatus(tc.status)))
		})
	}
}

// TestRequeueWaitUntilCanModify checks the message names the observed status,
// including when the status has not been read yet.
func TestRequeueWaitUntilCanModify(t *testing.T) {
	err := requeueWaitUntilCanModify(capabilityWithStatus(aws.String("CREATING")))
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "CREATING")
	assert.Contains(t, err.Error(), "ACTIVE")

	err = requeueWaitUntilCanModify(capabilityWithStatus(nil))
	require.NotNil(t, err)
}
