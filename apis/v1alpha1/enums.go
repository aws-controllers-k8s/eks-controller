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

type AMITypes string

const (
	AMITypes_AL2023_ARM_64_STANDARD     AMITypes = "AL2023_ARM_64_STANDARD"
	AMITypes_AL2023_x86_64_NEURON       AMITypes = "AL2023_x86_64_NEURON"
	AMITypes_AL2023_x86_64_NVIDIA       AMITypes = "AL2023_x86_64_NVIDIA"
	AMITypes_AL2023_x86_64_STANDARD     AMITypes = "AL2023_x86_64_STANDARD"
	AMITypes_AL2_ARM_64                 AMITypes = "AL2_ARM_64"
	AMITypes_AL2_x86_64                 AMITypes = "AL2_x86_64"
	AMITypes_AL2_x86_64_GPU             AMITypes = "AL2_x86_64_GPU"
	AMITypes_BOTTLEROCKET_ARM_64        AMITypes = "BOTTLEROCKET_ARM_64"
	AMITypes_BOTTLEROCKET_ARM_64_NVIDIA AMITypes = "BOTTLEROCKET_ARM_64_NVIDIA"
	AMITypes_BOTTLEROCKET_x86_64        AMITypes = "BOTTLEROCKET_x86_64"
	AMITypes_BOTTLEROCKET_x86_64_NVIDIA AMITypes = "BOTTLEROCKET_x86_64_NVIDIA"
	AMITypes_CUSTOM                     AMITypes = "CUSTOM"
	AMITypes_WINDOWS_CORE_2019_x86_64   AMITypes = "WINDOWS_CORE_2019_x86_64"
	AMITypes_WINDOWS_CORE_2022_x86_64   AMITypes = "WINDOWS_CORE_2022_x86_64"
	AMITypes_WINDOWS_FULL_2019_x86_64   AMITypes = "WINDOWS_FULL_2019_x86_64"
	AMITypes_WINDOWS_FULL_2022_x86_64   AMITypes = "WINDOWS_FULL_2022_x86_64"
)

type AccessScopeType string

const (
	AccessScopeType_cluster   AccessScopeType = "cluster"
	AccessScopeType_namespace AccessScopeType = "namespace"
)

type AddonIssueCode string

const (
	AddonIssueCode_AccessDenied                 AddonIssueCode = "AccessDenied"
	AddonIssueCode_AddonPermissionFailure       AddonIssueCode = "AddonPermissionFailure"
	AddonIssueCode_AddonSubscriptionNeeded      AddonIssueCode = "AddonSubscriptionNeeded"
	AddonIssueCode_AdmissionRequestDenied       AddonIssueCode = "AdmissionRequestDenied"
	AddonIssueCode_ClusterUnreachable           AddonIssueCode = "ClusterUnreachable"
	AddonIssueCode_ConfigurationConflict        AddonIssueCode = "ConfigurationConflict"
	AddonIssueCode_InsufficientNumberOfReplicas AddonIssueCode = "InsufficientNumberOfReplicas"
	AddonIssueCode_InternalFailure              AddonIssueCode = "InternalFailure"
	AddonIssueCode_K8sResourceNotFound          AddonIssueCode = "K8sResourceNotFound"
	AddonIssueCode_UnsupportedAddonModification AddonIssueCode = "UnsupportedAddonModification"
)

type AddonStatus_SDK string

const (
	AddonStatus_SDK_ACTIVE        AddonStatus_SDK = "ACTIVE"
	AddonStatus_SDK_CREATE_FAILED AddonStatus_SDK = "CREATE_FAILED"
	AddonStatus_SDK_CREATING      AddonStatus_SDK = "CREATING"
	AddonStatus_SDK_DEGRADED      AddonStatus_SDK = "DEGRADED"
	AddonStatus_SDK_DELETE_FAILED AddonStatus_SDK = "DELETE_FAILED"
	AddonStatus_SDK_DELETING      AddonStatus_SDK = "DELETING"
	AddonStatus_SDK_UPDATE_FAILED AddonStatus_SDK = "UPDATE_FAILED"
	AddonStatus_SDK_UPDATING      AddonStatus_SDK = "UPDATING"
)

type AuthenticationMode string

const (
	AuthenticationMode_API                AuthenticationMode = "API"
	AuthenticationMode_API_AND_CONFIG_MAP AuthenticationMode = "API_AND_CONFIG_MAP"
	AuthenticationMode_CONFIG_MAP         AuthenticationMode = "CONFIG_MAP"
)

type CapacityTypes string

const (
	CapacityTypes_CAPACITY_BLOCK CapacityTypes = "CAPACITY_BLOCK"
	CapacityTypes_ON_DEMAND      CapacityTypes = "ON_DEMAND"
	CapacityTypes_SPOT           CapacityTypes = "SPOT"
)

type Category string

const (
	Category_UPGRADE_READINESS Category = "UPGRADE_READINESS"
)

type ClusterIssueCode string

const (
	ClusterIssueCode_AccessDenied                ClusterIssueCode = "AccessDenied"
	ClusterIssueCode_ClusterUnreachable          ClusterIssueCode = "ClusterUnreachable"
	ClusterIssueCode_ConfigurationConflict       ClusterIssueCode = "ConfigurationConflict"
	ClusterIssueCode_Ec2SecurityGroupNotFound    ClusterIssueCode = "Ec2SecurityGroupNotFound"
	ClusterIssueCode_Ec2ServiceNotSubscribed     ClusterIssueCode = "Ec2ServiceNotSubscribed"
	ClusterIssueCode_Ec2SubnetNotFound           ClusterIssueCode = "Ec2SubnetNotFound"
	ClusterIssueCode_IamRoleNotFound             ClusterIssueCode = "IamRoleNotFound"
	ClusterIssueCode_InsufficientFreeAddresses   ClusterIssueCode = "InsufficientFreeAddresses"
	ClusterIssueCode_InternalFailure             ClusterIssueCode = "InternalFailure"
	ClusterIssueCode_KmsGrantRevoked             ClusterIssueCode = "KmsGrantRevoked"
	ClusterIssueCode_KmsKeyDisabled              ClusterIssueCode = "KmsKeyDisabled"
	ClusterIssueCode_KmsKeyMarkedForDeletion     ClusterIssueCode = "KmsKeyMarkedForDeletion"
	ClusterIssueCode_KmsKeyNotFound              ClusterIssueCode = "KmsKeyNotFound"
	ClusterIssueCode_Other                       ClusterIssueCode = "Other"
	ClusterIssueCode_ResourceLimitExceeded       ClusterIssueCode = "ResourceLimitExceeded"
	ClusterIssueCode_ResourceNotFound            ClusterIssueCode = "ResourceNotFound"
	ClusterIssueCode_StsRegionalEndpointDisabled ClusterIssueCode = "StsRegionalEndpointDisabled"
	ClusterIssueCode_UnsupportedVersion          ClusterIssueCode = "UnsupportedVersion"
	ClusterIssueCode_VpcNotFound                 ClusterIssueCode = "VpcNotFound"
)

type ClusterStatus_SDK string

const (
	ClusterStatus_SDK_ACTIVE   ClusterStatus_SDK = "ACTIVE"
	ClusterStatus_SDK_CREATING ClusterStatus_SDK = "CREATING"
	ClusterStatus_SDK_DELETING ClusterStatus_SDK = "DELETING"
	ClusterStatus_SDK_FAILED   ClusterStatus_SDK = "FAILED"
	ClusterStatus_SDK_PENDING  ClusterStatus_SDK = "PENDING"
	ClusterStatus_SDK_UPDATING ClusterStatus_SDK = "UPDATING"
)

type ConfigStatus string

const (
	ConfigStatus_ACTIVE   ConfigStatus = "ACTIVE"
	ConfigStatus_CREATING ConfigStatus = "CREATING"
	ConfigStatus_DELETING ConfigStatus = "DELETING"
)

type ConnectorConfigProvider string

const (
	ConnectorConfigProvider_AKS          ConnectorConfigProvider = "AKS"
	ConnectorConfigProvider_ANTHOS       ConnectorConfigProvider = "ANTHOS"
	ConnectorConfigProvider_EC2          ConnectorConfigProvider = "EC2"
	ConnectorConfigProvider_EKS_ANYWHERE ConnectorConfigProvider = "EKS_ANYWHERE"
	ConnectorConfigProvider_GKE          ConnectorConfigProvider = "GKE"
	ConnectorConfigProvider_OPENSHIFT    ConnectorConfigProvider = "OPENSHIFT"
	ConnectorConfigProvider_OTHER        ConnectorConfigProvider = "OTHER"
	ConnectorConfigProvider_RANCHER      ConnectorConfigProvider = "RANCHER"
	ConnectorConfigProvider_TANZU        ConnectorConfigProvider = "TANZU"
)

type EKSAnywhereSubscriptionLicenseType string

const (
	EKSAnywhereSubscriptionLicenseType_Cluster EKSAnywhereSubscriptionLicenseType = "Cluster"
)

type EKSAnywhereSubscriptionStatus string

const (
	EKSAnywhereSubscriptionStatus_ACTIVE   EKSAnywhereSubscriptionStatus = "ACTIVE"
	EKSAnywhereSubscriptionStatus_CREATING EKSAnywhereSubscriptionStatus = "CREATING"
	EKSAnywhereSubscriptionStatus_DELETING EKSAnywhereSubscriptionStatus = "DELETING"
	EKSAnywhereSubscriptionStatus_EXPIRED  EKSAnywhereSubscriptionStatus = "EXPIRED"
	EKSAnywhereSubscriptionStatus_EXPIRING EKSAnywhereSubscriptionStatus = "EXPIRING"
	EKSAnywhereSubscriptionStatus_UPDATING EKSAnywhereSubscriptionStatus = "UPDATING"
)

type EKSAnywhereSubscriptionTermUnit string

const (
	EKSAnywhereSubscriptionTermUnit_MONTHS EKSAnywhereSubscriptionTermUnit = "MONTHS"
)

type ErrorCode string

const (
	ErrorCode_AccessDenied                 ErrorCode = "AccessDenied"
	ErrorCode_AdmissionRequestDenied       ErrorCode = "AdmissionRequestDenied"
	ErrorCode_ClusterUnreachable           ErrorCode = "ClusterUnreachable"
	ErrorCode_ConfigurationConflict        ErrorCode = "ConfigurationConflict"
	ErrorCode_EniLimitReached              ErrorCode = "EniLimitReached"
	ErrorCode_InsufficientFreeAddresses    ErrorCode = "InsufficientFreeAddresses"
	ErrorCode_InsufficientNumberOfReplicas ErrorCode = "InsufficientNumberOfReplicas"
	ErrorCode_IpNotAvailable               ErrorCode = "IpNotAvailable"
	ErrorCode_K8sResourceNotFound          ErrorCode = "K8sResourceNotFound"
	ErrorCode_NodeCreationFailure          ErrorCode = "NodeCreationFailure"
	ErrorCode_OperationNotPermitted        ErrorCode = "OperationNotPermitted"
	ErrorCode_PodEvictionFailure           ErrorCode = "PodEvictionFailure"
	ErrorCode_SecurityGroupNotFound        ErrorCode = "SecurityGroupNotFound"
	ErrorCode_SubnetNotFound               ErrorCode = "SubnetNotFound"
	ErrorCode_Unknown                      ErrorCode = "Unknown"
	ErrorCode_UnsupportedAddonModification ErrorCode = "UnsupportedAddonModification"
	ErrorCode_VpcIdNotFound                ErrorCode = "VpcIdNotFound"
)

type FargateProfileIssueCode string

const (
	FargateProfileIssueCode_AccessDenied                 FargateProfileIssueCode = "AccessDenied"
	FargateProfileIssueCode_ClusterUnreachable           FargateProfileIssueCode = "ClusterUnreachable"
	FargateProfileIssueCode_InternalFailure              FargateProfileIssueCode = "InternalFailure"
	FargateProfileIssueCode_PodExecutionRoleAlreadyInUse FargateProfileIssueCode = "PodExecutionRoleAlreadyInUse"
)

type FargateProfileStatus_SDK string

const (
	FargateProfileStatus_SDK_ACTIVE        FargateProfileStatus_SDK = "ACTIVE"
	FargateProfileStatus_SDK_CREATE_FAILED FargateProfileStatus_SDK = "CREATE_FAILED"
	FargateProfileStatus_SDK_CREATING      FargateProfileStatus_SDK = "CREATING"
	FargateProfileStatus_SDK_DELETE_FAILED FargateProfileStatus_SDK = "DELETE_FAILED"
	FargateProfileStatus_SDK_DELETING      FargateProfileStatus_SDK = "DELETING"
)

type IPFamily string

const (
	IPFamily_ipv4 IPFamily = "ipv4"
	IPFamily_ipv6 IPFamily = "ipv6"
)

type InsightStatusValue string

const (
	InsightStatusValue_ERROR   InsightStatusValue = "ERROR"
	InsightStatusValue_PASSING InsightStatusValue = "PASSING"
	InsightStatusValue_UNKNOWN InsightStatusValue = "UNKNOWN"
	InsightStatusValue_WARNING InsightStatusValue = "WARNING"
)

type LogType string

const (
	LogType_api               LogType = "api"
	LogType_audit             LogType = "audit"
	LogType_authenticator     LogType = "authenticator"
	LogType_controllerManager LogType = "controllerManager"
	LogType_scheduler         LogType = "scheduler"
)

type NodegroupIssueCode string

const (
	NodegroupIssueCode_AccessDenied                             NodegroupIssueCode = "AccessDenied"
	NodegroupIssueCode_AmiIdNotFound                            NodegroupIssueCode = "AmiIdNotFound"
	NodegroupIssueCode_AsgInstanceLaunchFailures                NodegroupIssueCode = "AsgInstanceLaunchFailures"
	NodegroupIssueCode_AutoScalingGroupInstanceRefreshActive    NodegroupIssueCode = "AutoScalingGroupInstanceRefreshActive"
	NodegroupIssueCode_AutoScalingGroupInvalidConfiguration     NodegroupIssueCode = "AutoScalingGroupInvalidConfiguration"
	NodegroupIssueCode_AutoScalingGroupNotFound                 NodegroupIssueCode = "AutoScalingGroupNotFound"
	NodegroupIssueCode_AutoScalingGroupOptInRequired            NodegroupIssueCode = "AutoScalingGroupOptInRequired"
	NodegroupIssueCode_AutoScalingGroupRateLimitExceeded        NodegroupIssueCode = "AutoScalingGroupRateLimitExceeded"
	NodegroupIssueCode_ClusterUnreachable                       NodegroupIssueCode = "ClusterUnreachable"
	NodegroupIssueCode_Ec2InstanceTypeDoesNotExist              NodegroupIssueCode = "Ec2InstanceTypeDoesNotExist"
	NodegroupIssueCode_Ec2LaunchTemplateDeletionFailure         NodegroupIssueCode = "Ec2LaunchTemplateDeletionFailure"
	NodegroupIssueCode_Ec2LaunchTemplateInvalidConfiguration    NodegroupIssueCode = "Ec2LaunchTemplateInvalidConfiguration"
	NodegroupIssueCode_Ec2LaunchTemplateMaxLimitExceeded        NodegroupIssueCode = "Ec2LaunchTemplateMaxLimitExceeded"
	NodegroupIssueCode_Ec2LaunchTemplateNotFound                NodegroupIssueCode = "Ec2LaunchTemplateNotFound"
	NodegroupIssueCode_Ec2LaunchTemplateVersionMaxLimitExceeded NodegroupIssueCode = "Ec2LaunchTemplateVersionMaxLimitExceeded"
	NodegroupIssueCode_Ec2LaunchTemplateVersionMismatch         NodegroupIssueCode = "Ec2LaunchTemplateVersionMismatch"
	NodegroupIssueCode_Ec2SecurityGroupDeletionFailure          NodegroupIssueCode = "Ec2SecurityGroupDeletionFailure"
	NodegroupIssueCode_Ec2SecurityGroupNotFound                 NodegroupIssueCode = "Ec2SecurityGroupNotFound"
	NodegroupIssueCode_Ec2SubnetInvalidConfiguration            NodegroupIssueCode = "Ec2SubnetInvalidConfiguration"
	NodegroupIssueCode_Ec2SubnetListTooLong                     NodegroupIssueCode = "Ec2SubnetListTooLong"
	NodegroupIssueCode_Ec2SubnetMissingIpv6Assignment           NodegroupIssueCode = "Ec2SubnetMissingIpv6Assignment"
	NodegroupIssueCode_Ec2SubnetNotFound                        NodegroupIssueCode = "Ec2SubnetNotFound"
	NodegroupIssueCode_IamInstanceProfileNotFound               NodegroupIssueCode = "IamInstanceProfileNotFound"
	NodegroupIssueCode_IamLimitExceeded                         NodegroupIssueCode = "IamLimitExceeded"
	NodegroupIssueCode_IamNodeRoleNotFound                      NodegroupIssueCode = "IamNodeRoleNotFound"
	NodegroupIssueCode_IamThrottling                            NodegroupIssueCode = "IamThrottling"
	NodegroupIssueCode_InstanceLimitExceeded                    NodegroupIssueCode = "InstanceLimitExceeded"
	NodegroupIssueCode_InsufficientFreeAddresses                NodegroupIssueCode = "InsufficientFreeAddresses"
	NodegroupIssueCode_InternalFailure                          NodegroupIssueCode = "InternalFailure"
	NodegroupIssueCode_KubernetesLabelInvalid                   NodegroupIssueCode = "KubernetesLabelInvalid"
	NodegroupIssueCode_LimitExceeded                            NodegroupIssueCode = "LimitExceeded"
	NodegroupIssueCode_NodeCreationFailure                      NodegroupIssueCode = "NodeCreationFailure"
	NodegroupIssueCode_NodeTerminationFailure                   NodegroupIssueCode = "NodeTerminationFailure"
	NodegroupIssueCode_PodEvictionFailure                       NodegroupIssueCode = "PodEvictionFailure"
	NodegroupIssueCode_SourceEc2LaunchTemplateNotFound          NodegroupIssueCode = "SourceEc2LaunchTemplateNotFound"
	NodegroupIssueCode_Unknown                                  NodegroupIssueCode = "Unknown"
)

type NodegroupStatus_SDK string

const (
	NodegroupStatus_SDK_ACTIVE        NodegroupStatus_SDK = "ACTIVE"
	NodegroupStatus_SDK_CREATE_FAILED NodegroupStatus_SDK = "CREATE_FAILED"
	NodegroupStatus_SDK_CREATING      NodegroupStatus_SDK = "CREATING"
	NodegroupStatus_SDK_DEGRADED      NodegroupStatus_SDK = "DEGRADED"
	NodegroupStatus_SDK_DELETE_FAILED NodegroupStatus_SDK = "DELETE_FAILED"
	NodegroupStatus_SDK_DELETING      NodegroupStatus_SDK = "DELETING"
	NodegroupStatus_SDK_UPDATING      NodegroupStatus_SDK = "UPDATING"
)

type ResolveConflicts string

const (
	ResolveConflicts_NONE      ResolveConflicts = "NONE"
	ResolveConflicts_OVERWRITE ResolveConflicts = "OVERWRITE"
	ResolveConflicts_PRESERVE  ResolveConflicts = "PRESERVE"
)

type SupportType string

const (
	SupportType_EXTENDED SupportType = "EXTENDED"
	SupportType_STANDARD SupportType = "STANDARD"
)

type TaintEffect string

const (
	TaintEffect_NO_EXECUTE         TaintEffect = "NO_EXECUTE"
	TaintEffect_NO_SCHEDULE        TaintEffect = "NO_SCHEDULE"
	TaintEffect_PREFER_NO_SCHEDULE TaintEffect = "PREFER_NO_SCHEDULE"
)

type UpdateParamType string

const (
	UpdateParamType_AddonVersion             UpdateParamType = "AddonVersion"
	UpdateParamType_AuthenticationMode       UpdateParamType = "AuthenticationMode"
	UpdateParamType_ClusterLogging           UpdateParamType = "ClusterLogging"
	UpdateParamType_ComputeConfig            UpdateParamType = "ComputeConfig"
	UpdateParamType_ConfigurationValues      UpdateParamType = "ConfigurationValues"
	UpdateParamType_DesiredSize              UpdateParamType = "DesiredSize"
	UpdateParamType_EncryptionConfig         UpdateParamType = "EncryptionConfig"
	UpdateParamType_EndpointPrivateAccess    UpdateParamType = "EndpointPrivateAccess"
	UpdateParamType_EndpointPublicAccess     UpdateParamType = "EndpointPublicAccess"
	UpdateParamType_IdentityProviderConfig   UpdateParamType = "IdentityProviderConfig"
	UpdateParamType_KubernetesNetworkConfig  UpdateParamType = "KubernetesNetworkConfig"
	UpdateParamType_LabelsToAdd              UpdateParamType = "LabelsToAdd"
	UpdateParamType_LabelsToRemove           UpdateParamType = "LabelsToRemove"
	UpdateParamType_LaunchTemplateName       UpdateParamType = "LaunchTemplateName"
	UpdateParamType_LaunchTemplateVersion    UpdateParamType = "LaunchTemplateVersion"
	UpdateParamType_MaxSize                  UpdateParamType = "MaxSize"
	UpdateParamType_MaxUnavailable           UpdateParamType = "MaxUnavailable"
	UpdateParamType_MaxUnavailablePercentage UpdateParamType = "MaxUnavailablePercentage"
	UpdateParamType_MinSize                  UpdateParamType = "MinSize"
	UpdateParamType_PlatformVersion          UpdateParamType = "PlatformVersion"
	UpdateParamType_PodIdentityAssociations  UpdateParamType = "PodIdentityAssociations"
	UpdateParamType_PublicAccessCidrs        UpdateParamType = "PublicAccessCidrs"
	UpdateParamType_ReleaseVersion           UpdateParamType = "ReleaseVersion"
	UpdateParamType_ResolveConflicts         UpdateParamType = "ResolveConflicts"
	UpdateParamType_SecurityGroups           UpdateParamType = "SecurityGroups"
	UpdateParamType_ServiceAccountRoleArn    UpdateParamType = "ServiceAccountRoleArn"
	UpdateParamType_StorageConfig            UpdateParamType = "StorageConfig"
	UpdateParamType_Subnets                  UpdateParamType = "Subnets"
	UpdateParamType_TaintsToAdd              UpdateParamType = "TaintsToAdd"
	UpdateParamType_TaintsToRemove           UpdateParamType = "TaintsToRemove"
	UpdateParamType_UpgradePolicy            UpdateParamType = "UpgradePolicy"
	UpdateParamType_Version                  UpdateParamType = "Version"
	UpdateParamType_ZonalShiftConfig         UpdateParamType = "ZonalShiftConfig"
)

type UpdateStatus string

const (
	UpdateStatus_Cancelled  UpdateStatus = "Cancelled"
	UpdateStatus_Failed     UpdateStatus = "Failed"
	UpdateStatus_InProgress UpdateStatus = "InProgress"
	UpdateStatus_Successful UpdateStatus = "Successful"
)

type UpdateType string

const (
	UpdateType_AccessConfigUpdate                 UpdateType = "AccessConfigUpdate"
	UpdateType_AddonUpdate                        UpdateType = "AddonUpdate"
	UpdateType_AssociateEncryptionConfig          UpdateType = "AssociateEncryptionConfig"
	UpdateType_AssociateIdentityProviderConfig    UpdateType = "AssociateIdentityProviderConfig"
	UpdateType_AutoModeUpdate                     UpdateType = "AutoModeUpdate"
	UpdateType_ConfigUpdate                       UpdateType = "ConfigUpdate"
	UpdateType_DisassociateIdentityProviderConfig UpdateType = "DisassociateIdentityProviderConfig"
	UpdateType_EndpointAccessUpdate               UpdateType = "EndpointAccessUpdate"
	UpdateType_LoggingUpdate                      UpdateType = "LoggingUpdate"
	UpdateType_UpgradePolicyUpdate                UpdateType = "UpgradePolicyUpdate"
	UpdateType_VersionUpdate                      UpdateType = "VersionUpdate"
	UpdateType_VpcConfigUpdate                    UpdateType = "VpcConfigUpdate"
	UpdateType_ZonalShiftConfigUpdate             UpdateType = "ZonalShiftConfigUpdate"
)
