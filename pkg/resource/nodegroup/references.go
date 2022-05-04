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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ec2apitypes "github.com/aws-controllers-k8s/ec2-controller/apis/v1alpha1"
	iamapitypes "github.com/aws-controllers-k8s/iam-controller/apis/v1alpha1"
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	acktypes "github.com/aws-controllers-k8s/runtime/pkg/types"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

// +kubebuilder:rbac:groups=iam.services.k8s.aws,resources=roles,verbs=get;list
// +kubebuilder:rbac:groups=iam.services.k8s.aws,resources=roles/status,verbs=get;list

// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=securitygroups,verbs=get;list
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=securitygroups/status,verbs=get;list

// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=subnets,verbs=get;list
// +kubebuilder:rbac:groups=ec2.services.k8s.aws,resources=subnets/status,verbs=get;list

// ResolveReferences finds if there are any Reference field(s) present
// inside AWSResource passed in the parameter and attempts to resolve
// those reference field(s) into target field(s).
// It returns an AWSResource with resolved reference(s), and an error if the
// passed AWSResource's reference field(s) cannot be resolved.
// This method also adds/updates the ConditionTypeReferencesResolved for the
// AWSResource.
func (rm *resourceManager) ResolveReferences(
	ctx context.Context,
	apiReader client.Reader,
	res acktypes.AWSResource,
) (acktypes.AWSResource, error) {
	namespace := res.MetaObject().GetNamespace()
	ko := rm.concreteResource(res).ko.DeepCopy()
	err := validateReferenceFields(ko)
	if err == nil {
		err = resolveReferenceForClusterName(ctx, apiReader, namespace, ko)
	}
	if err == nil {
		err = resolveReferenceForNodeRole(ctx, apiReader, namespace, ko)
	}
	if err == nil {
		err = resolveReferenceForRemoteAccess_SourceSecurityGroups(ctx, apiReader, namespace, ko)
	}
	if err == nil {
		err = resolveReferenceForSubnets(ctx, apiReader, namespace, ko)
	}

	if hasNonNilReferences(ko) {
		return ackcondition.WithReferencesResolvedCondition(&resource{ko}, err)
	}
	return &resource{ko}, err
}

// validateReferenceFields validates the reference field and corresponding
// identifier field.
func validateReferenceFields(ko *svcapitypes.Nodegroup) error {
	if ko.Spec.ClusterRef != nil && ko.Spec.ClusterName != nil {
		return ackerr.ResourceReferenceAndIDNotSupportedFor("ClusterName", "ClusterRef")
	}
	if ko.Spec.ClusterRef == nil && ko.Spec.ClusterName == nil {
		return ackerr.ResourceReferenceOrIDRequiredFor("ClusterName", "ClusterRef")
	}
	if ko.Spec.NodeRoleRef != nil && ko.Spec.NodeRole != nil {
		return ackerr.ResourceReferenceAndIDNotSupportedFor("NodeRole", "NodeRoleRef")
	}
	if ko.Spec.NodeRoleRef == nil && ko.Spec.NodeRole == nil {
		return ackerr.ResourceReferenceOrIDRequiredFor("NodeRole", "NodeRoleRef")
	}
	if ko.Spec.RemoteAccess != nil {
		if ko.Spec.RemoteAccess.SourceSecurityGroupRefs != nil && ko.Spec.RemoteAccess.SourceSecurityGroups != nil {
			return ackerr.ResourceReferenceAndIDNotSupportedFor("RemoteAccess.SourceSecurityGroups", "RemoteAccess.SourceSecurityGroupRefs")
		}
	}
	if ko.Spec.SubnetRefs != nil && ko.Spec.Subnets != nil {
		return ackerr.ResourceReferenceAndIDNotSupportedFor("Subnets", "SubnetRefs")
	}
	if ko.Spec.SubnetRefs == nil && ko.Spec.Subnets == nil {
		return ackerr.ResourceReferenceOrIDRequiredFor("Subnets", "SubnetRefs")
	}
	return nil
}

// hasNonNilReferences returns true if resource contains a reference to another
// resource
func hasNonNilReferences(ko *svcapitypes.Nodegroup) bool {
	return false || (ko.Spec.ClusterRef != nil) || (ko.Spec.NodeRoleRef != nil) || (ko.Spec.RemoteAccess != nil && ko.Spec.RemoteAccess.SourceSecurityGroupRefs != nil) || (ko.Spec.SubnetRefs != nil)
}

// resolveReferenceForClusterName reads the resource referenced
// from ClusterRef field and sets the ClusterName
// from referenced resource
func resolveReferenceForClusterName(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.Nodegroup,
) error {
	if ko.Spec.ClusterRef != nil &&
		ko.Spec.ClusterRef.From != nil {
		arr := ko.Spec.ClusterRef.From
		if arr == nil || arr.Name == nil || *arr.Name == "" {
			return fmt.Errorf("provided resource reference is nil or empty")
		}
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      *arr.Name,
		}
		obj := svcapitypes.Cluster{}
		err := apiReader.Get(ctx, namespacedName, &obj)
		if err != nil {
			return err
		}
		var refResourceSynced, refResourceTerminal bool
		for _, cond := range obj.Status.Conditions {
			if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
				cond.Status == corev1.ConditionTrue {
				refResourceSynced = true
			}
			if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
				cond.Status == corev1.ConditionTrue {
				refResourceTerminal = true
			}
		}
		if refResourceTerminal {
			return ackerr.ResourceReferenceTerminalFor(
				"Cluster",
				namespace, *arr.Name)
		}
		if !refResourceSynced {
			return ackerr.ResourceReferenceNotSyncedFor(
				"Cluster",
				namespace, *arr.Name)
		}
		if obj.Spec.Name == nil {
			return ackerr.ResourceReferenceMissingTargetFieldFor(
				"Cluster",
				namespace, *arr.Name,
				"Spec.Name")
		}
		referencedValue := string(*obj.Spec.Name)
		ko.Spec.ClusterName = &referencedValue
	}
	return nil
}

// resolveReferenceForNodeRole reads the resource referenced
// from NodeRoleRef field and sets the NodeRole
// from referenced resource
func resolveReferenceForNodeRole(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.Nodegroup,
) error {
	if ko.Spec.NodeRoleRef != nil &&
		ko.Spec.NodeRoleRef.From != nil {
		arr := ko.Spec.NodeRoleRef.From
		if arr == nil || arr.Name == nil || *arr.Name == "" {
			return fmt.Errorf("provided resource reference is nil or empty")
		}
		namespacedName := types.NamespacedName{
			Namespace: namespace,
			Name:      *arr.Name,
		}
		obj := iamapitypes.Role{}
		err := apiReader.Get(ctx, namespacedName, &obj)
		if err != nil {
			return err
		}
		var refResourceSynced, refResourceTerminal bool
		for _, cond := range obj.Status.Conditions {
			if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
				cond.Status == corev1.ConditionTrue {
				refResourceSynced = true
			}
			if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
				cond.Status == corev1.ConditionTrue {
				refResourceTerminal = true
			}
		}
		if refResourceTerminal {
			return ackerr.ResourceReferenceTerminalFor(
				"Role",
				namespace, *arr.Name)
		}
		if !refResourceSynced {
			return ackerr.ResourceReferenceNotSyncedFor(
				"Role",
				namespace, *arr.Name)
		}
		if obj.Status.ACKResourceMetadata.ARN == nil {
			return ackerr.ResourceReferenceMissingTargetFieldFor(
				"Role",
				namespace, *arr.Name,
				"Status.ACKResourceMetadata.ARN")
		}
		referencedValue := string(*obj.Status.ACKResourceMetadata.ARN)
		ko.Spec.NodeRole = &referencedValue
	}
	return nil
}

// resolveReferenceForRemoteAccess_SourceSecurityGroups reads the resource referenced
// from RemoteAccess.SourceSecurityGroupRefs field and sets the RemoteAccess.SourceSecurityGroups
// from referenced resource
func resolveReferenceForRemoteAccess_SourceSecurityGroups(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.Nodegroup,
) error {
	if ko.Spec.RemoteAccess == nil {
		return nil
	}
	if ko.Spec.RemoteAccess.SourceSecurityGroupRefs != nil &&
		len(ko.Spec.RemoteAccess.SourceSecurityGroupRefs) > 0 {
		resolvedReferences := []*string{}
		for _, arrw := range ko.Spec.RemoteAccess.SourceSecurityGroupRefs {
			arr := arrw.From
			if arr == nil || arr.Name == nil || *arr.Name == "" {
				return fmt.Errorf("provided resource reference is nil or empty")
			}
			namespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      *arr.Name,
			}
			obj := ec2apitypes.SecurityGroup{}
			err := apiReader.Get(ctx, namespacedName, &obj)
			if err != nil {
				return err
			}
			var refResourceSynced, refResourceTerminal bool
			for _, cond := range obj.Status.Conditions {
				if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
					cond.Status == corev1.ConditionTrue {
					refResourceSynced = true
				}
				if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
					cond.Status == corev1.ConditionTrue {
					refResourceTerminal = true
				}
			}
			if refResourceTerminal {
				return ackerr.ResourceReferenceTerminalFor(
					"SecurityGroup",
					namespace, *arr.Name)
			}
			if !refResourceSynced {
				return ackerr.ResourceReferenceNotSyncedFor(
					"SecurityGroup",
					namespace, *arr.Name)
			}
			if obj.Status.ID == nil {
				return ackerr.ResourceReferenceMissingTargetFieldFor(
					"SecurityGroup",
					namespace, *arr.Name,
					"Status.ID")
			}
			referencedValue := string(*obj.Status.ID)
			resolvedReferences = append(resolvedReferences, &referencedValue)
		}
		ko.Spec.RemoteAccess.SourceSecurityGroups = resolvedReferences
	}
	return nil
}

// resolveReferenceForSubnets reads the resource referenced
// from SubnetRefs field and sets the Subnets
// from referenced resource
func resolveReferenceForSubnets(
	ctx context.Context,
	apiReader client.Reader,
	namespace string,
	ko *svcapitypes.Nodegroup,
) error {
	if ko.Spec.SubnetRefs != nil &&
		len(ko.Spec.SubnetRefs) > 0 {
		resolvedReferences := []*string{}
		for _, arrw := range ko.Spec.SubnetRefs {
			arr := arrw.From
			if arr == nil || arr.Name == nil || *arr.Name == "" {
				return fmt.Errorf("provided resource reference is nil or empty")
			}
			namespacedName := types.NamespacedName{
				Namespace: namespace,
				Name:      *arr.Name,
			}
			obj := ec2apitypes.Subnet{}
			err := apiReader.Get(ctx, namespacedName, &obj)
			if err != nil {
				return err
			}
			var refResourceSynced, refResourceTerminal bool
			for _, cond := range obj.Status.Conditions {
				if cond.Type == ackv1alpha1.ConditionTypeResourceSynced &&
					cond.Status == corev1.ConditionTrue {
					refResourceSynced = true
				}
				if cond.Type == ackv1alpha1.ConditionTypeTerminal &&
					cond.Status == corev1.ConditionTrue {
					refResourceTerminal = true
				}
			}
			if refResourceTerminal {
				return ackerr.ResourceReferenceTerminalFor(
					"Subnet",
					namespace, *arr.Name)
			}
			if !refResourceSynced {
				return ackerr.ResourceReferenceNotSyncedFor(
					"Subnet",
					namespace, *arr.Name)
			}
			if obj.Status.SubnetID == nil {
				return ackerr.ResourceReferenceMissingTargetFieldFor(
					"Subnet",
					namespace, *arr.Name,
					"Status.SubnetID")
			}
			referencedValue := string(*obj.Status.SubnetID)
			resolvedReferences = append(resolvedReferences, &referencedValue)
		}
		ko.Spec.Subnets = resolvedReferences
	}
	return nil
}
