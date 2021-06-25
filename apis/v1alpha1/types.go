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

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = &aws.JSONValue{}
	_ = ackv1alpha1.AWSAccountID("")
)

// The health of the add-on.
type AddonHealth struct {
	Issues []*AddonIssue `json:"issues,omitempty"`
}

// Information about an add-on.
type AddonInfo struct {
	AddonName *string `json:"addonName,omitempty"`
	Type      *string `json:"type_,omitempty"`
}

// An issue related to an add-on.
type AddonIssue struct {
	Code        *string   `json:"code,omitempty"`
	Message     *string   `json:"message,omitempty"`
	ResourceIDs []*string `json:"resourceIDs,omitempty"`
}

// Information about an add-on version.
type AddonVersionInfo struct {
	AddonVersion *string   `json:"addonVersion,omitempty"`
	Architecture []*string `json:"architecture,omitempty"`
}

// An Amazon EKS add-on.
type Addon_SDK struct {
	AddonARN     *string      `json:"addonARN,omitempty"`
	AddonName    *string      `json:"addonName,omitempty"`
	AddonVersion *string      `json:"addonVersion,omitempty"`
	ClusterName  *string      `json:"clusterName,omitempty"`
	CreatedAt    *metav1.Time `json:"createdAt,omitempty"`
	// The health of the add-on.
	Health                *AddonHealth       `json:"health,omitempty"`
	ModifiedAt            *metav1.Time       `json:"modifiedAt,omitempty"`
	ServiceAccountRoleARN *string            `json:"serviceAccountRoleARN,omitempty"`
	Status                *string            `json:"status,omitempty"`
	Tags                  map[string]*string `json:"tags,omitempty"`
}

// An Auto Scaling group that is associated with an Amazon EKS managed node
// group.
type AutoScalingGroup struct {
	Name *string `json:"name,omitempty"`
}

// An object representing the certificate-authority-data for your cluster.
type Certificate struct {
	Data *string `json:"data,omitempty"`
}

// An object representing an Amazon EKS cluster.
type Cluster_SDK struct {
	ARN *string `json:"arn,omitempty"`
	// An object representing the certificate-authority-data for your cluster.
	CertificateAuthority *Certificate        `json:"certificateAuthority,omitempty"`
	ClientRequestToken   *string             `json:"clientRequestToken,omitempty"`
	CreatedAt            *metav1.Time        `json:"createdAt,omitempty"`
	EncryptionConfig     []*EncryptionConfig `json:"encryptionConfig,omitempty"`
	Endpoint             *string             `json:"endpoint,omitempty"`
	// An object representing an identity provider.
	Identity *Identity `json:"identity,omitempty"`
	// The Kubernetes network configuration for the cluster.
	KubernetesNetworkConfig *KubernetesNetworkConfigResponse `json:"kubernetesNetworkConfig,omitempty"`
	// An object representing the logging configuration for resources in your cluster.
	Logging         *Logging `json:"logging,omitempty"`
	Name            *string  `json:"name,omitempty"`
	PlatformVersion *string  `json:"platformVersion,omitempty"`
	// An object representing an Amazon EKS cluster VPC configuration response.
	ResourcesVPCConfig *VPCConfigResponse `json:"resourcesVPCConfig,omitempty"`
	RoleARN            *string            `json:"roleARN,omitempty"`
	Status             *string            `json:"status,omitempty"`
	Tags               map[string]*string `json:"tags,omitempty"`
	Version            *string            `json:"version,omitempty"`
}

// Compatibility information.
type Compatibility struct {
	ClusterVersion   *string   `json:"clusterVersion,omitempty"`
	DefaultVersion   *bool     `json:"defaultVersion,omitempty"`
	PlatformVersions []*string `json:"platformVersions,omitempty"`
}

// The encryption configuration for the cluster.
type EncryptionConfig struct {
	// Identifies the AWS Key Management Service (AWS KMS) key used to encrypt the
	// secrets.
	Provider  *Provider `json:"provider,omitempty"`
	Resources []*string `json:"resources,omitempty"`
}

// An object representing an error when an asynchronous operation fails.
type ErrorDetail struct {
	ErrorCode    *string   `json:"errorCode,omitempty"`
	ErrorMessage *string   `json:"errorMessage,omitempty"`
	ResourceIDs  []*string `json:"resourceIDs,omitempty"`
}

// An object representing an AWS Fargate profile selector.
type FargateProfileSelector struct {
	Labels    map[string]*string `json:"labels,omitempty"`
	Namespace *string            `json:"namespace,omitempty"`
}

// An object representing an AWS Fargate profile.
type FargateProfile_SDK struct {
	ClusterName         *string                   `json:"clusterName,omitempty"`
	CreatedAt           *metav1.Time              `json:"createdAt,omitempty"`
	FargateProfileARN   *string                   `json:"fargateProfileARN,omitempty"`
	FargateProfileName  *string                   `json:"fargateProfileName,omitempty"`
	PodExecutionRoleARN *string                   `json:"podExecutionRoleARN,omitempty"`
	Selectors           []*FargateProfileSelector `json:"selectors,omitempty"`
	Status              *string                   `json:"status,omitempty"`
	Subnets             []*string                 `json:"subnets,omitempty"`
	Tags                map[string]*string        `json:"tags,omitempty"`
}

// An object representing an identity provider.
type Identity struct {
	// An object representing the OpenID Connect (https://openid.net/connect/) (OIDC)
	// identity provider information for the cluster.
	OIDC *OIDC `json:"oidc,omitempty"`
}

// An object representing an identity provider configuration.
type IdentityProviderConfig struct {
	Name *string `json:"name,omitempty"`
	Type *string `json:"type_,omitempty"`
}

// An object representing an issue with an Amazon EKS resource.
type Issue struct {
	Code        *string   `json:"code,omitempty"`
	Message     *string   `json:"message,omitempty"`
	ResourceIDs []*string `json:"resourceIDs,omitempty"`
}

// The Kubernetes network configuration for the cluster.
type KubernetesNetworkConfigRequest struct {
	ServiceIPv4CIDR *string `json:"serviceIPv4CIDR,omitempty"`
}

// The Kubernetes network configuration for the cluster.
type KubernetesNetworkConfigResponse struct {
	ServiceIPv4CIDR *string `json:"serviceIPv4CIDR,omitempty"`
}

// An object representing a node group launch template specification. The launch
// template cannot include SubnetId (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateNetworkInterface.html),
// IamInstanceProfile (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IamInstanceProfile.html),
// RequestSpotInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotInstances.html),
// HibernationOptions (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_HibernationOptionsRequest.html),
// or TerminateInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TerminateInstances.html),
// or the node group deployment or update will fail. For more information about
// launch templates, see CreateLaunchTemplate (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateLaunchTemplate.html)
// in the Amazon EC2 API Reference. For more information about using launch
// templates with Amazon EKS, see Launch template support (https://docs.aws.amazon.com/eks/latest/userguide/launch-templates.html)
// in the Amazon EKS User Guide.
//
// Specify either name or id, but not both.
type LaunchTemplateSpecification struct {
	ID      *string `json:"id,omitempty"`
	Name    *string `json:"name,omitempty"`
	Version *string `json:"version,omitempty"`
}

// An object representing the enabled or disabled Kubernetes control plane logs
// for your cluster.
type LogSetup struct {
	Enabled *bool     `json:"enabled,omitempty"`
	Types   []*string `json:"types,omitempty"`
}

// An object representing the logging configuration for resources in your cluster.
type Logging struct {
	ClusterLogging []*LogSetup `json:"clusterLogging,omitempty"`
}

// An object representing the health status of the node group.
type NodegroupHealth struct {
	Issues []*Issue `json:"issues,omitempty"`
}

// An object representing the resources associated with the node group, such
// as Auto Scaling groups and security groups for remote access.
type NodegroupResources struct {
	AutoScalingGroups         []*AutoScalingGroup `json:"autoScalingGroups,omitempty"`
	RemoteAccessSecurityGroup *string             `json:"remoteAccessSecurityGroup,omitempty"`
}

// An object representing the scaling configuration details for the Auto Scaling
// group that is associated with your node group. When creating a node group,
// you must specify all or none of the properties. When updating a node group,
// you can specify any or none of the properties.
type NodegroupScalingConfig struct {
	DesiredSize *int64 `json:"desiredSize,omitempty"`
	MaxSize     *int64 `json:"maxSize,omitempty"`
	MinSize     *int64 `json:"minSize,omitempty"`
}

type NodegroupUpdateConfig struct {
	MaxUnavailable           *int64 `json:"maxUnavailable,omitempty"`
	MaxUnavailablePercentage *int64 `json:"maxUnavailablePercentage,omitempty"`
}

// An object representing an Amazon EKS managed node group.
type Nodegroup_SDK struct {
	AmiType      *string      `json:"amiType,omitempty"`
	CapacityType *string      `json:"capacityType,omitempty"`
	ClusterName  *string      `json:"clusterName,omitempty"`
	CreatedAt    *metav1.Time `json:"createdAt,omitempty"`
	DiskSize     *int64       `json:"diskSize,omitempty"`
	// An object representing the health status of the node group.
	Health        *NodegroupHealth   `json:"health,omitempty"`
	InstanceTypes []*string          `json:"instanceTypes,omitempty"`
	Labels        map[string]*string `json:"labels,omitempty"`
	// An object representing a node group launch template specification. The launch
	// template cannot include SubnetId (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateNetworkInterface.html),
	// IamInstanceProfile (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IamInstanceProfile.html),
	// RequestSpotInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotInstances.html),
	// HibernationOptions (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_HibernationOptionsRequest.html),
	// or TerminateInstances (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_TerminateInstances.html),
	// or the node group deployment or update will fail. For more information about
	// launch templates, see CreateLaunchTemplate (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateLaunchTemplate.html)
	// in the Amazon EC2 API Reference. For more information about using launch
	// templates with Amazon EKS, see Launch template support (https://docs.aws.amazon.com/eks/latest/userguide/launch-templates.html)
	// in the Amazon EKS User Guide.
	//
	// Specify either name or id, but not both.
	LaunchTemplate *LaunchTemplateSpecification `json:"launchTemplate,omitempty"`
	ModifiedAt     *metav1.Time                 `json:"modifiedAt,omitempty"`
	NodeRole       *string                      `json:"nodeRole,omitempty"`
	NodegroupARN   *string                      `json:"nodegroupARN,omitempty"`
	NodegroupName  *string                      `json:"nodegroupName,omitempty"`
	ReleaseVersion *string                      `json:"releaseVersion,omitempty"`
	// An object representing the remote access configuration for the managed node
	// group.
	RemoteAccess *RemoteAccessConfig `json:"remoteAccess,omitempty"`
	// An object representing the resources associated with the node group, such
	// as Auto Scaling groups and security groups for remote access.
	Resources *NodegroupResources `json:"resources,omitempty"`
	// An object representing the scaling configuration details for the Auto Scaling
	// group that is associated with your node group. When creating a node group,
	// you must specify all or none of the properties. When updating a node group,
	// you can specify any or none of the properties.
	ScalingConfig *NodegroupScalingConfig `json:"scalingConfig,omitempty"`
	Status        *string                 `json:"status,omitempty"`
	Subnets       []*string               `json:"subnets,omitempty"`
	Tags          map[string]*string      `json:"tags,omitempty"`
	Taints        []*Taint                `json:"taints,omitempty"`
	UpdateConfig  *NodegroupUpdateConfig  `json:"updateConfig,omitempty"`
	Version       *string                 `json:"version,omitempty"`
}

// An object representing the OpenID Connect (https://openid.net/connect/) (OIDC)
// identity provider information for the cluster.
type OIDC struct {
	Issuer *string `json:"issuer,omitempty"`
}

// An object that represents the configuration for an OpenID Connect (OIDC)
// identity provider.
type OIDCIdentityProviderConfig struct {
	ClientID                   *string            `json:"clientID,omitempty"`
	ClusterName                *string            `json:"clusterName,omitempty"`
	GroupsClaim                *string            `json:"groupsClaim,omitempty"`
	GroupsPrefix               *string            `json:"groupsPrefix,omitempty"`
	IdentityProviderConfigARN  *string            `json:"identityProviderConfigARN,omitempty"`
	IdentityProviderConfigName *string            `json:"identityProviderConfigName,omitempty"`
	IssuerURL                  *string            `json:"issuerURL,omitempty"`
	Tags                       map[string]*string `json:"tags,omitempty"`
	UsernameClaim              *string            `json:"usernameClaim,omitempty"`
	UsernamePrefix             *string            `json:"usernamePrefix,omitempty"`
}

// An object representing an OpenID Connect (OIDC) configuration. Before associating
// an OIDC identity provider to your cluster, review the considerations in Authenticating
// users for your cluster from an OpenID Connect identity provider (https://docs.aws.amazon.com/eks/latest/userguide/authenticate-oidc-identity-provider.html)
// in the Amazon EKS User Guide.
type OIDCIdentityProviderConfigRequest struct {
	ClientID                   *string `json:"clientID,omitempty"`
	GroupsClaim                *string `json:"groupsClaim,omitempty"`
	GroupsPrefix               *string `json:"groupsPrefix,omitempty"`
	IdentityProviderConfigName *string `json:"identityProviderConfigName,omitempty"`
	IssuerURL                  *string `json:"issuerURL,omitempty"`
	UsernameClaim              *string `json:"usernameClaim,omitempty"`
	UsernamePrefix             *string `json:"usernamePrefix,omitempty"`
}

// Identifies the AWS Key Management Service (AWS KMS) key used to encrypt the
// secrets.
type Provider struct {
	KeyARN *string `json:"keyARN,omitempty"`
}

// An object representing the remote access configuration for the managed node
// group.
type RemoteAccessConfig struct {
	EC2SshKey            *string   `json:"ec2SshKey,omitempty"`
	SourceSecurityGroups []*string `json:"sourceSecurityGroups,omitempty"`
}

// A property that allows a node to repel a set of pods.
type Taint struct {
	Effect *string `json:"effect,omitempty"`
	Key    *string `json:"key,omitempty"`
	Value  *string `json:"value,omitempty"`
}

// An object representing an asynchronous update.
type Update struct {
	CreatedAt *metav1.Time   `json:"createdAt,omitempty"`
	Errors    []*ErrorDetail `json:"errors,omitempty"`
	ID        *string        `json:"id,omitempty"`
	Params    []*UpdateParam `json:"params,omitempty"`
	Status    *string        `json:"status,omitempty"`
	Type      *string        `json:"type_,omitempty"`
}

// An object representing a Kubernetes label change for a managed node group.
type UpdateLabelsPayload struct {
	AddOrUpdateLabels map[string]*string `json:"addOrUpdateLabels,omitempty"`
}

// An object representing the details of an update request.
type UpdateParam struct {
	Type  *string `json:"type_,omitempty"`
	Value *string `json:"value,omitempty"`
}

// An object representing the details of an update to a taints payload.
type UpdateTaintsPayload struct {
	AddOrUpdateTaints []*Taint `json:"addOrUpdateTaints,omitempty"`
	RemoveTaints      []*Taint `json:"removeTaints,omitempty"`
}

// An object representing the VPC configuration to use for an Amazon EKS cluster.
type VPCConfigRequest struct {
	EndpointPrivateAccess *bool     `json:"endpointPrivateAccess,omitempty"`
	EndpointPublicAccess  *bool     `json:"endpointPublicAccess,omitempty"`
	PublicAccessCIDRs     []*string `json:"publicAccessCIDRs,omitempty"`
	SecurityGroupIDs      []*string `json:"securityGroupIDs,omitempty"`
	SubnetIDs             []*string `json:"subnetIDs,omitempty"`
}

// An object representing an Amazon EKS cluster VPC configuration response.
type VPCConfigResponse struct {
	ClusterSecurityGroupID *string   `json:"clusterSecurityGroupID,omitempty"`
	EndpointPrivateAccess  *bool     `json:"endpointPrivateAccess,omitempty"`
	EndpointPublicAccess   *bool     `json:"endpointPublicAccess,omitempty"`
	PublicAccessCIDRs      []*string `json:"publicAccessCIDRs,omitempty"`
	SecurityGroupIDs       []*string `json:"securityGroupIDs,omitempty"`
	SubnetIDs              []*string `json:"subnetIDs,omitempty"`
	VPCID                  *string   `json:"vpcID,omitempty"`
}
