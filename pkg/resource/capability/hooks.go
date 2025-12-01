package capability

import (
	"reflect"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eks"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
)

var syncTags = tags.SyncTags

// setConfiguration sets the configuration field of UpdateCapabilityInput.
// It especially compares which rbacRoleMappings need to be added and updated
// vs the ones that need to be removed
func setConfiguration(
	input *svcsdk.UpdateCapabilityInput,
	desired *resource,
	latest *resource,
) {
	if desired.ko.Spec.Configuration == nil || desired.ko.Spec.Configuration.ArgoCD == nil {
		input.Configuration = nil
		return
	}

	if input.Configuration == nil || input.Configuration.ArgoCd == nil {
		input.Configuration = &svcsdktypes.UpdateCapabilityConfiguration{
			ArgoCd: &svcsdktypes.UpdateArgoCdConfig{},
		}
	}

	if desired.ko.Spec.Configuration.ArgoCD.NetworkAccess != nil {
		input.Configuration.ArgoCd.NetworkAccess = &svcsdktypes.ArgoCdNetworkAccessConfigRequest{
			VpceIds: aws.ToStringSlice(desired.ko.Spec.Configuration.ArgoCD.NetworkAccess.VPCEIDs),
		}
	}
	if desired.ko.Spec.Configuration.ArgoCD.RbacRoleMappings != nil {
		input.Configuration.ArgoCd.RbacRoleMappings = &svcsdktypes.UpdateRoleMappings{}
		toAddOrUpdate, toRemove := compareRbacRoleMappings(desired.ko.Spec.Configuration.ArgoCD.RbacRoleMappings, latest.ko.Spec.Configuration.ArgoCD.RbacRoleMappings)
		if len(toAddOrUpdate) > 0 {
			input.Configuration.ArgoCd.RbacRoleMappings.AddOrUpdateRoleMappings = toAddOrUpdate
		}
		if len(toRemove) > 0 {
			input.Configuration.ArgoCd.RbacRoleMappings.RemoveRoleMappings = toRemove
		}
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
