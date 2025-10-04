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
	"errors"
	"fmt"
	"strconv"
	"time"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrequeue "github.com/aws-controllers-k8s/runtime/pkg/requeue"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/eks"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/smithy-go"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
	"github.com/aws-controllers-k8s/eks-controller/pkg/tags"
	"github.com/aws-controllers-k8s/eks-controller/pkg/util"
)

const (
	LoggingNoChangesError = "No changes needed for the logging config provided"
)

// Taken from the list of cluster statuses on the boto3 documentation
// https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/eks.html#EKS.Client.describe_cluster
const (
	StatusCreating = "CREATING"
	StatusActive   = "ACTIVE"
	StatusDeleting = "DELETING"
	StatusFailed   = "FAILED"
	StatusUpdating = "UPDATING"
	StatusPending  = "PENDING"
)

// GetForceUpgrade returns whether the cluster version upgrade should be forced
// as determined by the annotation on the object, or the default value otherwise.
func GetForceUpgrade(
	m *metav1.ObjectMeta,
) bool {
	resAnnotations := m.GetAnnotations()
	forceUpgrade, ok := resAnnotations[v1alpha1.ForceClusterUpgradeAnnotation]
	if !ok {
		return v1alpha1.DefaultForceClusterUpgrade
	}

	forceUpgradeBool, err := strconv.ParseBool(forceUpgrade)
	if err != nil {
		return v1alpha1.DefaultForceClusterUpgrade
	}

	return forceUpgradeBool
}

var (
	// TerminalStatuses are the status strings that are terminal states for a
	// cluster.
	TerminalStatuses = []string{
		StatusDeleting,
		StatusFailed,
	}
)

var (
	requeueWaitWhileDeleting = ackrequeue.NeededAfter(
		fmt.Errorf("cluster in '%s' state, cannot be modified or deleted", StatusDeleting),
		ackrequeue.DefaultRequeueAfterDuration,
	)
	requeueWaitWhileInUse = ackrequeue.NeededAfter(
		errors.New("cluster is still in use, cannot be deleted"),
		ackrequeue.DefaultRequeueAfterDuration,
	)
	RequeueAfterUpdateDuration = 15 * time.Second
)

// requeueWaitUntilCanModify returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the cluster cannot be modified until it reaches an active status.
func requeueWaitUntilCanModify(r *resource) *ackrequeue.RequeueNeededAfter {
	if r.ko.Status.Status == nil {
		return nil
	}
	status := *r.ko.Status.Status
	return ackrequeue.NeededAfter(
		fmt.Errorf("cluster in '%s' state, cannot be modified until '%s'",
			status, StatusActive),
		ackrequeue.DefaultRequeueAfterDuration,
	)
}

// requeueAfterAsyncUpdate returns a `ackrequeue.RequeueNeededAfter` struct
// explaining the cluster cannot be modified until after the asynchronous update
// has (first, started and then) completed and the cluster reaches an active
// status.
func requeueAfterAsyncUpdate() *ackrequeue.RequeueNeededAfter {
	return ackrequeue.NeededAfter(
		fmt.Errorf("cluster has started asynchronously updating, cannot be modified until '%s'",
			StatusActive),
		RequeueAfterUpdateDuration,
	)
}

// clusterHasTerminalStatus returns whether the supplied cluster is in a
// terminal state
func clusterHasTerminalStatus(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	for _, s := range TerminalStatuses {
		if cs == s {
			return true
		}
	}
	return false
}

// clusterActive returns true if the supplied cluster is in an active status
func clusterActive(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusActive
}

// clusterCreating returns true if the supplied cluster is in the process of
// being created
func clusterCreating(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusCreating
}

// clusterDeleting returns true if the supplied cluster is in the process of
// being deleted
func clusterDeleting(r *resource) bool {
	if r.ko.Status.Status == nil {
		return false
	}
	cs := *r.ko.Status.Status
	return cs == StatusDeleting
}

// returnClusterUpdating will set synced to false on the resource and
// return an async requeue error to signify that the resource should be
// forcefully requeued in order to pick up the 'UPDATING' status.
func returnClusterUpdating(r *resource) (*resource, error) {
	msg := "Cluster is currently being updated"
	ackcondition.SetSynced(r, corev1.ConditionFalse, &msg, nil)
	return r, requeueAfterAsyncUpdate()
}

// clusterInUse returns true if the supplied cluster is still being used. It
// determines this by checking if there are any nodegroups still attached to
// the cluster.
func (rm *resourceManager) clusterInUse(ctx context.Context, r *resource) (bool, error) {
	nodes, err := rm.listNodegroups(ctx, r)
	if err != nil {
		return false, err
	}

	// Cluster is in use if # of nodegroups != 0
	return (nodes != nil && len(nodes.Nodegroups) > 0), nil
}

func customPreCompare(
	a *resource,
	b *resource,
) {
	if a.ko.Spec.UpgradePolicy == nil && b.ko.Spec.UpgradePolicy != nil {
		a.ko.Spec.UpgradePolicy = b.ko.Spec.UpgradePolicy.DeepCopy()
	}
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

	// For asynchronous updates, latest (from ReadOne) contains outdated values
	// for Spec fields. However, the status (Cluster status) is correct inside latest.
	// So we construct the updatedRes object from the desired resource to
	// obtain correct spec fields and then copy the status from latest.
	updatedRes := rm.concreteResource(desired.DeepCopy())
	updatedRes.SetStatus(latest)
	if clusterDeleting(latest) {
		msg := "Cluster is currently being deleted"
		ackcondition.SetSynced(updatedRes, corev1.ConditionFalse, &msg, nil)
		return updatedRes, requeueWaitWhileDeleting
	}
	if !clusterActive(latest) {
		msg := "Cluster is in '" + *latest.ko.Status.Status + "' status"
		ackcondition.SetSynced(updatedRes, corev1.ConditionFalse, &msg, nil)
		if clusterHasTerminalStatus(latest) {
			ackcondition.SetTerminal(updatedRes, corev1.ConditionTrue, &msg, nil)
			return updatedRes, nil
		}
		return updatedRes, requeueWaitUntilCanModify(latest)
	}

	// Sync tags if they have changed
	if delta.DifferentAt("Spec.Tags") {
		err := tags.SyncTags(
			ctx,
			rm.sdkapi,
			rm.metrics,
			string(*latest.ko.Status.ACKResourceMetadata.ARN),
			aws.ToStringMap(desired.ko.Spec.Tags),
			aws.ToStringMap(latest.ko.Spec.Tags),
		)
		if err != nil {
			return nil, err
		}
	}

	// If no changes except tags, return the desired state
	if !delta.DifferentExcept("Spec.Tags") {
		return desired, nil
	}

	// Handle logging configuration updates
	if delta.DifferentAt("Spec.Logging") {
		if err := rm.updateConfigLogging(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// The API responds with an error if there were no changes applied
			if !ok || awsErr.Message != LoggingNoChangesError {
				return nil, err
			}

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}
		}
		return returnClusterUpdating(updatedRes)
	}

	// Handle VPC configuration updates for public and private access
	if delta.DifferentAt("Spec.ResourcesVPCConfig.EndpointPrivateAccess") ||
		delta.DifferentAt("Spec.ResourcesVPCConfig.EndpointPublicAccess") ||
		delta.DifferentAt("Spec.ResourcesVPCConfig.PublicAccessCIDRs") {
		if err := rm.updateConfigResourcesVPCConfigPublicAndPrivateAccess(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	// Handle VPC configuration updates for subnets and security groups
	if delta.DifferentAt("Spec.ResourcesVPCConfig.SecurityGroupIDs") ||
		delta.DifferentAt("Spec.ResourcesVPCConfig.SecurityGroupRefs") ||
		delta.DifferentAt("Spec.ResourcesVPCConfig.SubnetIDs") ||
		delta.DifferentAt("Spec.ResourcesVPCConfig.SubnetRefs") {
		if err := rm.updateConfigResourcesVPCConfigSubnetsAndSecurityGroups(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	// Handle access configuration updates
	if delta.DifferentAt("Spec.AccessConfig") {
		if err := rm.updateAccessConfig(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}
			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	// Handle upgrade policy updates
	if delta.DifferentAt("Spec.UpgradePolicy") {
		if err := rm.updateClusterUpgradePolicy(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}
			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	// Handle encryption configuration updates
	if delta.DifferentAt("Spec.EncryptionConfig") {
		// Set a terminal condition if the observed cluster has encryption
		// config and the desired cluster does not.
		if len(latest.ko.Spec.EncryptionConfig) > 0 && len(desired.ko.Spec.EncryptionConfig) == 0 {
			msg := "Encryption configuration cannot be removed from an existing cluster"
			ackcondition.SetTerminal(updatedRes, corev1.ConditionTrue, &msg, nil)
			return updatedRes, nil
		}
		// Set a terminal condition if the user tries to patch the encryption
		// config of an existing cluster.
		if len(latest.ko.Spec.EncryptionConfig) == 1 && len(desired.ko.Spec.EncryptionConfig) == 1 {
			msg := "Encryption configuration cannot be updated"
			ackcondition.SetTerminal(updatedRes, corev1.ConditionTrue, &msg, nil)
			return updatedRes, nil
		}
		// Set a terminal condition if the user tries to add a second encryption
		// config to an existing cluster.
		if len(latest.ko.Spec.EncryptionConfig) == 0 && len(desired.ko.Spec.EncryptionConfig) > 1 {
			msg := "Only one encryption configuration is allowed"
			ackcondition.SetTerminal(updatedRes, corev1.ConditionTrue, &msg, nil)
			return updatedRes, nil
		}

		if err := rm.associateEncryptionConfig(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		// This doesn't reflect the actual status of the cluster, so we have to explicitly
		// requeue and set the status to updating.
		updatedRes.ko.Status.Status = aws.String(string(svcsdktypes.ClusterStatusUpdating))
		return returnClusterUpdating(updatedRes)
	}

	// Handle version updates
	if delta.DifferentAt("Spec.Version") {
		if err := rm.updateVersion(ctx, desired, latest); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	// Handle computeConfig updates
	if delta.DifferentAt("Spec.ComputeConfig") || delta.DifferentAt("Spec.StorageConfig") || delta.DifferentAt("Spec.KubernetesNetworkConfig") {
		if err := rm.updateComputeConfig(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)
			rlog.Info("attempting to update AutoMode config",
				"error", err,
				"isAWSError", ok,
				"awsErrorCode", awsErr.Code)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				rlog.Info("resource in use, requeueing after async update")
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, fmt.Errorf("failed to update AutoMode config: %w", err)
		}

		return returnClusterUpdating(updatedRes)
	}

	// Handle zonalShiftConfig updates
	if delta.DifferentAt("Spec.ZonalShiftConfig") {
		if err := rm.updateZonalShiftConfig(ctx, desired); err != nil {
			awsErr, ok := extractAWSError(err)

			// Check to see if we've raced an async update call and need to requeue
			if ok && awsErr.Code == "ResourceInUseException" {
				return nil, requeueAfterAsyncUpdate()
			}

			return nil, err
		}
		return returnClusterUpdating(updatedRes)
	}

	// Set default status values and return the updated resource
	rm.setStatusDefaults(updatedRes.ko)
	return updatedRes, nil
}

// updateVersion updates the cluster version to the next possible version.
//
// This function isn't supposed to blindly update the cluster version to the
// desired version. It should increment the minor version of the observed
// version and update the cluster to that version. The reconciliation mechanism
// should ensure that the desired version is eventually reached.
func (rm *resourceManager) updateVersion(
	ctx context.Context,
	desired, latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateVersion")
	defer exit(err)

	// If the desired version is less than the observed version, we can't update
	// the cluster to an older version.
	// Note that the desired and observed versions are guaranteed to be never be
	// equal at this stage, as the delta comparison would have caught that.
	compareResult, err := util.CompareEKSKubernetesVersions(*desired.ko.Spec.Version, *latest.ko.Spec.Version)
	if err != nil {
		return ackerr.NewTerminalError(fmt.Errorf("failed to compare the desired and observed versions: %v", err))
	}
	if compareResult != 1 {
		return ackerr.NewTerminalError(
			fmt.Errorf("desired cluster version is less than the observed version: %s < %s",
				*desired.ko.Spec.Version, *latest.ko.Spec.Version,
			),
		)
	}

	// Compure the next minor version of the desired version
	nextVersion, err := util.IncrementEKSMinorVersion(*latest.ko.Spec.Version)
	if err != nil {
		return ackerr.NewTerminalError(fmt.Errorf("failed to compute the next minor version: %v", err))
	}

	input := &svcsdk.UpdateClusterVersionInput{
		Name:    desired.ko.Spec.Name,
		Version: &nextVersion,
		Force:   GetForceUpgrade(&desired.ko.ObjectMeta),
	}

	_, err = rm.sdkapi.UpdateClusterVersion(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterVersion", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfigLogging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfigLogging")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:    r.ko.Spec.Name,
		Logging: rm.newLogging(r),
	}

	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func newAccessConfig(r *resource) *svcsdktypes.UpdateAccessConfigRequest {
	cfg := &svcsdktypes.UpdateAccessConfigRequest{}
	if r.ko.Spec.AccessConfig != nil {
		cfg.AuthenticationMode = svcsdktypes.AuthenticationMode(*r.ko.Spec.AccessConfig.AuthenticationMode)
	}
	return cfg
}

func (rm *resourceManager) updateAccessConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateAccessConfig")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:         r.ko.Spec.Name,
		AccessConfig: newAccessConfig(r),
	}
	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateClusterUpgradePolicy(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateClusterUpgradePolicy")
	defer func() { exit(err) }()
	input := &svcsdk.UpdateClusterConfigInput{
		Name: r.ko.Spec.Name,
		UpgradePolicy: &svcsdktypes.UpgradePolicyRequest{
			SupportType: svcsdktypes.SupportType(*r.ko.Spec.UpgradePolicy.SupportType),
		},
	}
	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfigResourcesVPCConfigPublicAndPrivateAccess(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfigResourcesVPCConfigPublicAndPrivateAccess")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:               r.ko.Spec.Name,
		ResourcesVpcConfig: rm.newVpcConfigRequest(r),
	}

	// We only want to update endpointPrivateAccess, endpointPublicAccess and
	// publicAccessCidrs
	input.ResourcesVpcConfig.SubnetIds = nil
	input.ResourcesVpcConfig.SecurityGroupIds = nil

	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateConfigResourcesVPCConfigSubnetsAndSecurityGroups(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateConfigResourcesVPCConfigSubnetsAndSecurityGroups")
	defer exit(err)
	input := &svcsdk.UpdateClusterConfigInput{
		Name:               r.ko.Spec.Name,
		ResourcesVpcConfig: rm.newVpcConfigRequest(r),
	}

	// We only want to update securityGroupIds and subnetIds
	input.ResourcesVpcConfig.EndpointPublicAccess = nil
	input.ResourcesVpcConfig.EndpointPrivateAccess = nil
	input.ResourcesVpcConfig.PublicAccessCidrs = nil

	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) listNodegroups(
	ctx context.Context,
	r *resource,
) (nodes *svcsdk.ListNodegroupsOutput, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.listNodegroups")
	defer exit(err)
	input := &svcsdk.ListNodegroupsInput{
		ClusterName: r.ko.Spec.Name,
	}

	nodes, err = rm.sdkapi.ListNodegroups(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "ListNodegroups", err)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// associateEncryptionConfig associates the encryption configuration with the
// cluster.
func (rm *resourceManager) associateEncryptionConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateEncryptionConfiguration")
	defer func() { exit(err) }()

	// Convert []*string to []string
	resources := make([]string, 0, len(r.ko.Spec.EncryptionConfig[0].Resources))
	for _, res := range r.ko.Spec.EncryptionConfig[0].Resources {
		if res != nil {
			resources = append(resources, *res)
		}
	}

	input := &svcsdk.AssociateEncryptionConfigInput{
		ClusterName: r.ko.Spec.Name,
		EncryptionConfig: []svcsdktypes.EncryptionConfig{
			{
				// Being it means that we already have a single encryption config
				// in the spec. So we can safely assume that the first element is
				// the only one.
				Resources: resources,
				Provider: &svcsdktypes.Provider{
					KeyArn: r.ko.Spec.EncryptionConfig[0].Provider.KeyARN,
				},
			},
		},
	}

	_, err = rm.sdkapi.AssociateEncryptionConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "AssociateEncryptionConfig", err)
	if err != nil {
		return err
	}

	return nil
}

// updateComputeConfig updates the compute config of the cluster.
func (rm *resourceManager) updateComputeConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateComputeConfig")
	defer exit(err)

	// All three fields must be initialized for EKS Auto Mode to be properly toggled.
	input := &svcsdk.UpdateClusterConfigInput{
		Name:          r.ko.Spec.Name,
		ComputeConfig: &svcsdktypes.ComputeConfigRequest{},
		StorageConfig: &svcsdktypes.StorageConfigRequest{
			BlockStorage: &svcsdktypes.BlockStorage{},
		},
		KubernetesNetworkConfig: &svcsdktypes.KubernetesNetworkConfigRequest{
			ElasticLoadBalancing: &svcsdktypes.ElasticLoadBalancing{},
		},
	}

	// Process compute configs
	if r.ko.Spec.ComputeConfig != nil {
		// Convert []*string to []string for NodePools
		nodePools := make([]string, 0, len(r.ko.Spec.ComputeConfig.NodePools))
		for _, nodePool := range r.ko.Spec.ComputeConfig.NodePools {
			if nodePool != nil {
				nodePools = append(nodePools, *nodePool)
			}
		}

		input.ComputeConfig.Enabled = r.ko.Spec.ComputeConfig.Enabled
		input.ComputeConfig.NodePools = nodePools // Use the converted []string slice
		input.ComputeConfig.NodeRoleArn = r.ko.Spec.ComputeConfig.NodeRoleARN
	}

	// Process storage configs
	if r.ko.Spec.StorageConfig != nil && r.ko.Spec.StorageConfig.BlockStorage != nil {
		input.StorageConfig.BlockStorage.Enabled = r.ko.Spec.StorageConfig.BlockStorage.Enabled
	}

	// Process network configs
	if r.ko.Spec.KubernetesNetworkConfig != nil {
		if r.ko.Spec.KubernetesNetworkConfig.ElasticLoadBalancing != nil {
			input.KubernetesNetworkConfig.ElasticLoadBalancing.Enabled = r.ko.Spec.KubernetesNetworkConfig.ElasticLoadBalancing.Enabled
		}

		if r.ko.Spec.KubernetesNetworkConfig.IPFamily != nil {
			input.KubernetesNetworkConfig.IpFamily = svcsdktypes.IpFamily(*r.ko.Spec.KubernetesNetworkConfig.IPFamily)
		}

		input.KubernetesNetworkConfig.ServiceIpv4Cidr = r.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR
	}

	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) updateZonalShiftConfig(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateZonalShiftConfig")
	defer exit(err)

	input := &svcsdk.UpdateClusterConfigInput{
		Name: r.ko.Spec.Name,
		ZonalShiftConfig: &svcsdktypes.ZonalShiftConfigRequest{
			Enabled: r.ko.Spec.ZonalShiftConfig.Enabled,
		},
	}

	_, err = rm.sdkapi.UpdateClusterConfig(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateClusterConfig", err)
	if err != nil {
		return err
	}

	return nil
}

// extractAWSError extracts the underlying AWS error from a smithy.GenericAPIError.
func extractAWSError(err error) (awsErr *smithy.GenericAPIError, ok bool) {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return &smithy.GenericAPIError{
			Code:    apiErr.ErrorCode(),
			Message: apiErr.ErrorMessage(),
			Fault:   apiErr.ErrorFault(),
		}, true
	}
	return nil, false
}
