package capability

import (
	"errors"
	"fmt"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eks"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
)

var syncTags = tags.SyncTags

// Capability statuses, from the list in the EKS API reference:
// https://docs.aws.amazon.com/eks/latest/APIReference/API_Capability.html
const (
	StatusActive   = "ACTIVE"
	StatusCreating = "CREATING"
	StatusUpdating = "UPDATING"
	StatusDeleting = "DELETING"
)

// inFlightStatuses are the statuses where EKS has an operation in progress on
// the capability and rejects UpdateCapability with ResourceInUseException.
//
// CREATE_FAILED, DELETE_FAILED and DEGRADED are deliberately absent. They are
// settled states rather than work in progress, and an update is how a user
// recovers a DEGRADED capability, so gating on them would leave the resource
// requeueing with no path forward.
var inFlightStatuses = []string{StatusCreating, StatusUpdating, StatusDeleting}

// capabilityInFlight reports whether EKS is mid-operation on the capability.
func capabilityInFlight(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	for _, status := range inFlightStatuses {
		if *r.ko.Status.Status == status {
			return true
		}
	}
	return false
}

// requeueWaitUntilCanModify explains that the capability cannot be updated
// until EKS finishes the operation already running against it.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	status := ""
	if r.ko.Status.Status != nil {
		status = *r.ko.Status.Status
	}
	return ackrequeue.NeededAfter(
		fmt.Errorf("capability in '%s' state, cannot be modified until '%s'",
			status, StatusActive),
		ackrequeue.DefaultRequeueAfterDuration,
	)
}

// validateConfiguration rejects a spec that declares both configuration
// blocks. A capability has exactly one type, so at most one block can apply to
// it, and no capability makes both meaningful.
//
// The controller answers this rather than leaving it to EKS because the two
// possible API behaviors have very different costs. A rejected request costs a
// round trip. A request where EKS silently drops the block that does not match
// the capability type costs convergence: DescribeCapability never reports the
// dropped block back, so the delta survives every resync and the resource
// updates forever.
func validateConfiguration(r *resource) error {
	cfg := r.ko.Spec.Configuration
	if cfg == nil {
		return nil
	}
	if cfg.ArgoCD != nil && cfg.Ack != nil {
		return ackerr.NewTerminalError(errors.New(
			"configuration must declare at most one of argoCD or ack; " +
				"a capability carries only the configuration for its own type",
		))
	}
	return nil
}

// differsOutsideConfiguration reports whether the delta holds any difference
// outside Spec.Tags, which the update path syncs separately, and
// Spec.Configuration.
//
// DifferentExcept cannot answer this. It short-circuits to true as soon as the
// delta holds more differences than it was given exception paths, even when
// every one of those differences sits under an excepted path.
func differsOutsideConfiguration(delta *ackcompare.Delta) bool {
	for _, diff := range delta.Differences {
		if diff.Path.Contains("Spec.Tags") || diff.Path.Contains("Spec.Configuration") {
			continue
		}
		return true
	}
	return false
}

// setConfiguration sets the configuration field of UpdateCapabilityInput.
// A capability carries the configuration block for its own type, so each
// branch is gated on the spec declaring that block. Without the gate, a
// capability configured for one type has its whole configuration dropped.
func setConfiguration(
	input *svcsdk.UpdateCapabilityInput,
	desired *resource,
	latest *resource,
) {
	cfg := desired.ko.Spec.Configuration
	if cfg == nil || (cfg.ArgoCD == nil && cfg.Ack == nil) {
		input.Configuration = nil
		return
	}

	if input.Configuration == nil {
		input.Configuration = &svcsdktypes.UpdateCapabilityConfiguration{}
	}

	if cfg.ArgoCD != nil {
		setArgoCDConfiguration(input.Configuration, cfg.ArgoCD, latestArgoCDConfiguration(latest))
	}
	if cfg.Ack != nil {
		setAckConfiguration(input.Configuration, cfg.Ack)
	}
}

// latestArgoCDConfiguration returns the observed Argo CD configuration, or nil
// when the capability has none. The RBAC mapping comparison needs the observed
// mappings to work out removals, and every field above them is optional.
func latestArgoCDConfiguration(latest *resource) *v1alpha1.ArgoCDConfigRequest {
	if latest == nil || latest.ko.Spec.Configuration == nil {
		return nil
	}
	return latest.ko.Spec.Configuration.ArgoCD
}

// setArgoCDConfiguration maps the declared Argo CD configuration onto the
// update input. It especially compares which rbacRoleMappings need to be added
// and updated vs the ones that need to be removed.
func setArgoCDConfiguration(
	input *svcsdktypes.UpdateCapabilityConfiguration,
	desired *v1alpha1.ArgoCDConfigRequest,
	latest *v1alpha1.ArgoCDConfigRequest,
) {
	if input.ArgoCd == nil {
		input.ArgoCd = &svcsdktypes.UpdateArgoCdConfig{}
	}

	if desired.NetworkAccess != nil {
		input.ArgoCd.NetworkAccess = &svcsdktypes.ArgoCdNetworkAccessConfigRequest{
			VpceIds: aws.ToStringSlice(desired.NetworkAccess.VPCEIDs),
		}
	}
	if desired.RbacRoleMappings != nil {
		var latestRoleMappings []*v1alpha1.ArgoCDRoleMapping
		if latest != nil {
			latestRoleMappings = latest.RbacRoleMappings
		}
		input.ArgoCd.RbacRoleMappings = &svcsdktypes.UpdateRoleMappings{}
		toAddOrUpdate, toRemove := compareRbacRoleMappings(desired.RbacRoleMappings, latestRoleMappings)
		if len(toAddOrUpdate) > 0 {
			input.ArgoCd.RbacRoleMappings.AddOrUpdateRoleMappings = toAddOrUpdate
		}
		if len(toRemove) > 0 {
			input.ArgoCd.RbacRoleMappings.RemoveRoleMappings = toRemove
		}
	}
}

// setAckConfiguration maps the declared ACK configuration onto the update
// input. Only declared fields are sent, because UpdateAckConfig reads an
// omitted field as "leave unchanged".
//
// An undeclared field is therefore left alone rather than cleared. That
// matters for a capability adopted under the adopt-or-create policy, where the
// spec is taken as authored and a block declaring only one field would
// otherwise wipe the configuration the capability already had. Late
// initialization fills the omitted field in from the observed state instead,
// after which it is declared and reconciled normally.
//
// A declared disabledServices is sent whole, because the API replaces the
// stored list rather than merging with it. An explicitly empty list is the
// documented way to turn every service back on, and it survives this path as
// an empty, non-nil slice.
func setAckConfiguration(
	input *svcsdktypes.UpdateCapabilityConfiguration,
	desired *v1alpha1.AckConfigRequest,
) {
	if input.Ack == nil {
		input.Ack = &svcsdktypes.UpdateAckConfig{}
	}

	if desired.EnableCrossNamespace != nil {
		input.Ack.EnableCrossNamespace = desired.EnableCrossNamespace
	}
	if desired.DisabledServices != nil {
		input.Ack.DisabledServices = aws.ToStringSlice(desired.DisabledServices)
	}
}

func compareRbacRoleMappings(desired []*v1alpha1.ArgoCDRoleMapping, latest []*v1alpha1.ArgoCDRoleMapping) ([]svcsdktypes.ArgoCdRoleMapping, []svcsdktypes.ArgoCdRoleMapping) {
	var toAddOrUpdate []svcsdktypes.ArgoCdRoleMapping
	var toRemove []svcsdktypes.ArgoCdRoleMapping

	for _, desiredMapping := range desired {
		found := false
		for _, latestMapping := range latest {
			if isRoleMappingEqual(desiredMapping, latestMapping) {
				found = true
				break
			}
		}
		if !found {
			toAddOrUpdate = append(toAddOrUpdate, roleMappingToServiceSDK(desiredMapping))
		}
	}

	for _, latestMapping := range latest {
		found := false
		for _, desiredMapping := range desired {
			if isRoleMappingEqual(desiredMapping, latestMapping) {
				found = true
				break
			}
		}
		if !found {
			toRemove = append(toRemove, roleMappingToServiceSDK(latestMapping))
		}
	}

	return toAddOrUpdate, toRemove
}

func isRoleMappingEqual(desired *v1alpha1.ArgoCDRoleMapping, latest *v1alpha1.ArgoCDRoleMapping) bool {
	return aws.ToString(desired.Role) == aws.ToString(latest.Role) &&
		reflect.DeepEqual(desired.Identities, latest.Identities)
}

func roleMappingToServiceSDK(roleMapping *v1alpha1.ArgoCDRoleMapping) svcsdktypes.ArgoCdRoleMapping {
	rm := svcsdktypes.ArgoCdRoleMapping{
		Role: svcsdktypes.ArgoCdRole(aws.ToString(roleMapping.Role)),
	}
	identities := make([]svcsdktypes.SsoIdentity, len(roleMapping.Identities))
	for i, identity := range roleMapping.Identities {
		identities[i] = svcsdktypes.SsoIdentity{
			Id:   identity.ID,
			Type: svcsdktypes.SsoIdentityType(aws.ToString(identity.Type)),
		}
	}
	return rm
}
