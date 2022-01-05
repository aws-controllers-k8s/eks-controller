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
	"context"
	"reflect"
	"strings"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackcondition "github.com/aws-controllers-k8s/runtime/pkg/condition"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/eks"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &aws.JSONValue{}
	_ = &svcsdk.EKS{}
	_ = &svcapitypes.Cluster{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
	_ = &ackcondition.NotManagedMessage
	_ = &reflect.Value{}
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkFind")
	defer exit(err)
	// If any required fields in the input shape are missing, AWS resource is
	// not created yet. Return NotFound here to indicate to callers that the
	// resource isn't yet created.
	if rm.requiredFieldsMissingFromReadOneInput(r) {
		return nil, ackerr.NotFound
	}

	input, err := rm.newDescribeRequestPayload(r)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.DescribeClusterOutput
	resp, err = rm.sdkapi.DescribeClusterWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "DescribeCluster", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "ResourceNotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.Cluster.Arn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.Cluster.Arn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Cluster.CertificateAuthority != nil {
		f1 := &svcapitypes.Certificate{}
		if resp.Cluster.CertificateAuthority.Data != nil {
			f1.Data = resp.Cluster.CertificateAuthority.Data
		}
		ko.Status.CertificateAuthority = f1
	} else {
		ko.Status.CertificateAuthority = nil
	}
	if resp.Cluster.ClientRequestToken != nil {
		ko.Spec.ClientRequestToken = resp.Cluster.ClientRequestToken
	} else {
		ko.Spec.ClientRequestToken = nil
	}
	if resp.Cluster.CreatedAt != nil {
		ko.Status.CreatedAt = &metav1.Time{*resp.Cluster.CreatedAt}
	} else {
		ko.Status.CreatedAt = nil
	}
	if resp.Cluster.EncryptionConfig != nil {
		f4 := []*svcapitypes.EncryptionConfig{}
		for _, f4iter := range resp.Cluster.EncryptionConfig {
			f4elem := &svcapitypes.EncryptionConfig{}
			if f4iter.Provider != nil {
				f4elemf0 := &svcapitypes.Provider{}
				if f4iter.Provider.KeyArn != nil {
					f4elemf0.KeyARN = f4iter.Provider.KeyArn
				}
				f4elem.Provider = f4elemf0
			}
			if f4iter.Resources != nil {
				f4elemf1 := []*string{}
				for _, f4elemf1iter := range f4iter.Resources {
					var f4elemf1elem string
					f4elemf1elem = *f4elemf1iter
					f4elemf1 = append(f4elemf1, &f4elemf1elem)
				}
				f4elem.Resources = f4elemf1
			}
			f4 = append(f4, f4elem)
		}
		ko.Spec.EncryptionConfig = f4
	} else {
		ko.Spec.EncryptionConfig = nil
	}
	if resp.Cluster.Endpoint != nil {
		ko.Status.Endpoint = resp.Cluster.Endpoint
	} else {
		ko.Status.Endpoint = nil
	}
	if resp.Cluster.Identity != nil {
		f6 := &svcapitypes.Identity{}
		if resp.Cluster.Identity.Oidc != nil {
			f6f0 := &svcapitypes.OIDC{}
			if resp.Cluster.Identity.Oidc.Issuer != nil {
				f6f0.Issuer = resp.Cluster.Identity.Oidc.Issuer
			}
			f6.OIDC = f6f0
		}
		ko.Status.Identity = f6
	} else {
		ko.Status.Identity = nil
	}
	if resp.Cluster.KubernetesNetworkConfig != nil {
		f7 := &svcapitypes.KubernetesNetworkConfigRequest{}
		if resp.Cluster.KubernetesNetworkConfig.ServiceIpv4Cidr != nil {
			f7.ServiceIPv4CIDR = resp.Cluster.KubernetesNetworkConfig.ServiceIpv4Cidr
		}
		ko.Spec.KubernetesNetworkConfig = f7
	} else {
		ko.Spec.KubernetesNetworkConfig = nil
	}
	if resp.Cluster.Logging != nil {
		f8 := &svcapitypes.Logging{}
		if resp.Cluster.Logging.ClusterLogging != nil {
			f8f0 := []*svcapitypes.LogSetup{}
			for _, f8f0iter := range resp.Cluster.Logging.ClusterLogging {
				f8f0elem := &svcapitypes.LogSetup{}
				if f8f0iter.Enabled != nil {
					f8f0elem.Enabled = f8f0iter.Enabled
				}
				if f8f0iter.Types != nil {
					f8f0elemf1 := []*string{}
					for _, f8f0elemf1iter := range f8f0iter.Types {
						var f8f0elemf1elem string
						f8f0elemf1elem = *f8f0elemf1iter
						f8f0elemf1 = append(f8f0elemf1, &f8f0elemf1elem)
					}
					f8f0elem.Types = f8f0elemf1
				}
				f8f0 = append(f8f0, f8f0elem)
			}
			f8.ClusterLogging = f8f0
		}
		ko.Spec.Logging = f8
	} else {
		ko.Spec.Logging = nil
	}
	if resp.Cluster.Name != nil {
		ko.Spec.Name = resp.Cluster.Name
	} else {
		ko.Spec.Name = nil
	}
	if resp.Cluster.PlatformVersion != nil {
		ko.Status.PlatformVersion = resp.Cluster.PlatformVersion
	} else {
		ko.Status.PlatformVersion = nil
	}
	if resp.Cluster.ResourcesVpcConfig != nil {
		f11 := &svcapitypes.VPCConfigRequest{}
		if resp.Cluster.ResourcesVpcConfig.EndpointPrivateAccess != nil {
			f11.EndpointPrivateAccess = resp.Cluster.ResourcesVpcConfig.EndpointPrivateAccess
		}
		if resp.Cluster.ResourcesVpcConfig.EndpointPublicAccess != nil {
			f11.EndpointPublicAccess = resp.Cluster.ResourcesVpcConfig.EndpointPublicAccess
		}
		if resp.Cluster.ResourcesVpcConfig.PublicAccessCidrs != nil {
			f11f3 := []*string{}
			for _, f11f3iter := range resp.Cluster.ResourcesVpcConfig.PublicAccessCidrs {
				var f11f3elem string
				f11f3elem = *f11f3iter
				f11f3 = append(f11f3, &f11f3elem)
			}
			f11.PublicAccessCIDRs = f11f3
		}
		if resp.Cluster.ResourcesVpcConfig.SecurityGroupIds != nil {
			f11f4 := []*string{}
			for _, f11f4iter := range resp.Cluster.ResourcesVpcConfig.SecurityGroupIds {
				var f11f4elem string
				f11f4elem = *f11f4iter
				f11f4 = append(f11f4, &f11f4elem)
			}
			f11.SecurityGroupIDs = f11f4
		}
		if resp.Cluster.ResourcesVpcConfig.SubnetIds != nil {
			f11f5 := []*string{}
			for _, f11f5iter := range resp.Cluster.ResourcesVpcConfig.SubnetIds {
				var f11f5elem string
				f11f5elem = *f11f5iter
				f11f5 = append(f11f5, &f11f5elem)
			}
			f11.SubnetIDs = f11f5
		}
		ko.Spec.ResourcesVPCConfig = f11
	} else {
		ko.Spec.ResourcesVPCConfig = nil
	}
	if resp.Cluster.RoleArn != nil {
		ko.Spec.RoleARN = resp.Cluster.RoleArn
	} else {
		ko.Spec.RoleARN = nil
	}
	if resp.Cluster.Status != nil {
		ko.Status.Status = resp.Cluster.Status
	} else {
		ko.Status.Status = nil
	}
	if resp.Cluster.Tags != nil {
		f14 := map[string]*string{}
		for f14key, f14valiter := range resp.Cluster.Tags {
			var f14val string
			f14val = *f14valiter
			f14[f14key] = &f14val
		}
		ko.Spec.Tags = f14
	} else {
		ko.Spec.Tags = nil
	}
	if resp.Cluster.Version != nil {
		ko.Spec.Version = resp.Cluster.Version
	} else {
		ko.Spec.Version = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// requiredFieldsMissingFromReadOneInput returns true if there are any fields
// for the ReadOne Input shape that are required but not present in the
// resource's Spec or Status
func (rm *resourceManager) requiredFieldsMissingFromReadOneInput(
	r *resource,
) bool {
	return r.ko.Spec.Name == nil

}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.DescribeClusterInput, error) {
	res := &svcsdk.DescribeClusterInput{}

	if r.ko.Spec.Name != nil {
		res.SetName(*r.ko.Spec.Name)
	}

	return res, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a copy of the resource with resource fields (in both Spec and
// Status) filled in with values from the CREATE API operation's Output shape.
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	desired *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkCreate")
	defer exit(err)
	input, err := rm.newCreateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.CreateClusterOutput
	_ = resp
	resp, err = rm.sdkapi.CreateClusterWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateCluster", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.Cluster.Arn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.Cluster.Arn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.Cluster.CertificateAuthority != nil {
		f1 := &svcapitypes.Certificate{}
		if resp.Cluster.CertificateAuthority.Data != nil {
			f1.Data = resp.Cluster.CertificateAuthority.Data
		}
		ko.Status.CertificateAuthority = f1
	} else {
		ko.Status.CertificateAuthority = nil
	}
	if resp.Cluster.ClientRequestToken != nil {
		ko.Spec.ClientRequestToken = resp.Cluster.ClientRequestToken
	} else {
		ko.Spec.ClientRequestToken = nil
	}
	if resp.Cluster.CreatedAt != nil {
		ko.Status.CreatedAt = &metav1.Time{*resp.Cluster.CreatedAt}
	} else {
		ko.Status.CreatedAt = nil
	}
	if resp.Cluster.EncryptionConfig != nil {
		f4 := []*svcapitypes.EncryptionConfig{}
		for _, f4iter := range resp.Cluster.EncryptionConfig {
			f4elem := &svcapitypes.EncryptionConfig{}
			if f4iter.Provider != nil {
				f4elemf0 := &svcapitypes.Provider{}
				if f4iter.Provider.KeyArn != nil {
					f4elemf0.KeyARN = f4iter.Provider.KeyArn
				}
				f4elem.Provider = f4elemf0
			}
			if f4iter.Resources != nil {
				f4elemf1 := []*string{}
				for _, f4elemf1iter := range f4iter.Resources {
					var f4elemf1elem string
					f4elemf1elem = *f4elemf1iter
					f4elemf1 = append(f4elemf1, &f4elemf1elem)
				}
				f4elem.Resources = f4elemf1
			}
			f4 = append(f4, f4elem)
		}
		ko.Spec.EncryptionConfig = f4
	} else {
		ko.Spec.EncryptionConfig = nil
	}
	if resp.Cluster.Endpoint != nil {
		ko.Status.Endpoint = resp.Cluster.Endpoint
	} else {
		ko.Status.Endpoint = nil
	}
	if resp.Cluster.Identity != nil {
		f6 := &svcapitypes.Identity{}
		if resp.Cluster.Identity.Oidc != nil {
			f6f0 := &svcapitypes.OIDC{}
			if resp.Cluster.Identity.Oidc.Issuer != nil {
				f6f0.Issuer = resp.Cluster.Identity.Oidc.Issuer
			}
			f6.OIDC = f6f0
		}
		ko.Status.Identity = f6
	} else {
		ko.Status.Identity = nil
	}
	if resp.Cluster.KubernetesNetworkConfig != nil {
		f7 := &svcapitypes.KubernetesNetworkConfigRequest{}
		if resp.Cluster.KubernetesNetworkConfig.ServiceIpv4Cidr != nil {
			f7.ServiceIPv4CIDR = resp.Cluster.KubernetesNetworkConfig.ServiceIpv4Cidr
		}
		ko.Spec.KubernetesNetworkConfig = f7
	} else {
		ko.Spec.KubernetesNetworkConfig = nil
	}
	if resp.Cluster.Logging != nil {
		f8 := &svcapitypes.Logging{}
		if resp.Cluster.Logging.ClusterLogging != nil {
			f8f0 := []*svcapitypes.LogSetup{}
			for _, f8f0iter := range resp.Cluster.Logging.ClusterLogging {
				f8f0elem := &svcapitypes.LogSetup{}
				if f8f0iter.Enabled != nil {
					f8f0elem.Enabled = f8f0iter.Enabled
				}
				if f8f0iter.Types != nil {
					f8f0elemf1 := []*string{}
					for _, f8f0elemf1iter := range f8f0iter.Types {
						var f8f0elemf1elem string
						f8f0elemf1elem = *f8f0elemf1iter
						f8f0elemf1 = append(f8f0elemf1, &f8f0elemf1elem)
					}
					f8f0elem.Types = f8f0elemf1
				}
				f8f0 = append(f8f0, f8f0elem)
			}
			f8.ClusterLogging = f8f0
		}
		ko.Spec.Logging = f8
	} else {
		ko.Spec.Logging = nil
	}
	if resp.Cluster.Name != nil {
		ko.Spec.Name = resp.Cluster.Name
	} else {
		ko.Spec.Name = nil
	}
	if resp.Cluster.PlatformVersion != nil {
		ko.Status.PlatformVersion = resp.Cluster.PlatformVersion
	} else {
		ko.Status.PlatformVersion = nil
	}
	if resp.Cluster.ResourcesVpcConfig != nil {
		f11 := &svcapitypes.VPCConfigRequest{}
		if resp.Cluster.ResourcesVpcConfig.EndpointPrivateAccess != nil {
			f11.EndpointPrivateAccess = resp.Cluster.ResourcesVpcConfig.EndpointPrivateAccess
		}
		if resp.Cluster.ResourcesVpcConfig.EndpointPublicAccess != nil {
			f11.EndpointPublicAccess = resp.Cluster.ResourcesVpcConfig.EndpointPublicAccess
		}
		if resp.Cluster.ResourcesVpcConfig.PublicAccessCidrs != nil {
			f11f3 := []*string{}
			for _, f11f3iter := range resp.Cluster.ResourcesVpcConfig.PublicAccessCidrs {
				var f11f3elem string
				f11f3elem = *f11f3iter
				f11f3 = append(f11f3, &f11f3elem)
			}
			f11.PublicAccessCIDRs = f11f3
		}
		if resp.Cluster.ResourcesVpcConfig.SecurityGroupIds != nil {
			f11f4 := []*string{}
			for _, f11f4iter := range resp.Cluster.ResourcesVpcConfig.SecurityGroupIds {
				var f11f4elem string
				f11f4elem = *f11f4iter
				f11f4 = append(f11f4, &f11f4elem)
			}
			f11.SecurityGroupIDs = f11f4
		}
		if resp.Cluster.ResourcesVpcConfig.SubnetIds != nil {
			f11f5 := []*string{}
			for _, f11f5iter := range resp.Cluster.ResourcesVpcConfig.SubnetIds {
				var f11f5elem string
				f11f5elem = *f11f5iter
				f11f5 = append(f11f5, &f11f5elem)
			}
			f11.SubnetIDs = f11f5
		}
		ko.Spec.ResourcesVPCConfig = f11
	} else {
		ko.Spec.ResourcesVPCConfig = nil
	}
	if resp.Cluster.RoleArn != nil {
		ko.Spec.RoleARN = resp.Cluster.RoleArn
	} else {
		ko.Spec.RoleARN = nil
	}
	if resp.Cluster.Status != nil {
		ko.Status.Status = resp.Cluster.Status
	} else {
		ko.Status.Status = nil
	}
	if resp.Cluster.Tags != nil {
		f14 := map[string]*string{}
		for f14key, f14valiter := range resp.Cluster.Tags {
			var f14val string
			f14val = *f14valiter
			f14[f14key] = &f14val
		}
		ko.Spec.Tags = f14
	} else {
		ko.Spec.Tags = nil
	}
	if resp.Cluster.Version != nil {
		ko.Spec.Version = resp.Cluster.Version
	} else {
		ko.Spec.Version = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateClusterInput, error) {
	res := &svcsdk.CreateClusterInput{}

	if r.ko.Spec.ClientRequestToken != nil {
		res.SetClientRequestToken(*r.ko.Spec.ClientRequestToken)
	}
	if r.ko.Spec.EncryptionConfig != nil {
		f1 := []*svcsdk.EncryptionConfig{}
		for _, f1iter := range r.ko.Spec.EncryptionConfig {
			f1elem := &svcsdk.EncryptionConfig{}
			if f1iter.Provider != nil {
				f1elemf0 := &svcsdk.Provider{}
				if f1iter.Provider.KeyARN != nil {
					f1elemf0.SetKeyArn(*f1iter.Provider.KeyARN)
				}
				f1elem.SetProvider(f1elemf0)
			}
			if f1iter.Resources != nil {
				f1elemf1 := []*string{}
				for _, f1elemf1iter := range f1iter.Resources {
					var f1elemf1elem string
					f1elemf1elem = *f1elemf1iter
					f1elemf1 = append(f1elemf1, &f1elemf1elem)
				}
				f1elem.SetResources(f1elemf1)
			}
			f1 = append(f1, f1elem)
		}
		res.SetEncryptionConfig(f1)
	}
	if r.ko.Spec.KubernetesNetworkConfig != nil {
		f2 := &svcsdk.KubernetesNetworkConfigRequest{}
		if r.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR != nil {
			f2.SetServiceIpv4Cidr(*r.ko.Spec.KubernetesNetworkConfig.ServiceIPv4CIDR)
		}
		res.SetKubernetesNetworkConfig(f2)
	}
	if r.ko.Spec.Logging != nil {
		f3 := &svcsdk.Logging{}
		if r.ko.Spec.Logging.ClusterLogging != nil {
			f3f0 := []*svcsdk.LogSetup{}
			for _, f3f0iter := range r.ko.Spec.Logging.ClusterLogging {
				f3f0elem := &svcsdk.LogSetup{}
				if f3f0iter.Enabled != nil {
					f3f0elem.SetEnabled(*f3f0iter.Enabled)
				}
				if f3f0iter.Types != nil {
					f3f0elemf1 := []*string{}
					for _, f3f0elemf1iter := range f3f0iter.Types {
						var f3f0elemf1elem string
						f3f0elemf1elem = *f3f0elemf1iter
						f3f0elemf1 = append(f3f0elemf1, &f3f0elemf1elem)
					}
					f3f0elem.SetTypes(f3f0elemf1)
				}
				f3f0 = append(f3f0, f3f0elem)
			}
			f3.SetClusterLogging(f3f0)
		}
		res.SetLogging(f3)
	}
	if r.ko.Spec.Name != nil {
		res.SetName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.ResourcesVPCConfig != nil {
		f5 := &svcsdk.VpcConfigRequest{}
		if r.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess != nil {
			f5.SetEndpointPrivateAccess(*r.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess)
		}
		if r.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess != nil {
			f5.SetEndpointPublicAccess(*r.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess)
		}
		if r.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs != nil {
			f5f2 := []*string{}
			for _, f5f2iter := range r.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs {
				var f5f2elem string
				f5f2elem = *f5f2iter
				f5f2 = append(f5f2, &f5f2elem)
			}
			f5.SetPublicAccessCidrs(f5f2)
		}
		if r.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs != nil {
			f5f3 := []*string{}
			for _, f5f3iter := range r.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs {
				var f5f3elem string
				f5f3elem = *f5f3iter
				f5f3 = append(f5f3, &f5f3elem)
			}
			f5.SetSecurityGroupIds(f5f3)
		}
		if r.ko.Spec.ResourcesVPCConfig.SubnetIDs != nil {
			f5f4 := []*string{}
			for _, f5f4iter := range r.ko.Spec.ResourcesVPCConfig.SubnetIDs {
				var f5f4elem string
				f5f4elem = *f5f4iter
				f5f4 = append(f5f4, &f5f4elem)
			}
			f5.SetSubnetIds(f5f4)
		}
		res.SetResourcesVpcConfig(f5)
	}
	if r.ko.Spec.RoleARN != nil {
		res.SetRoleArn(*r.ko.Spec.RoleARN)
	}
	if r.ko.Spec.Tags != nil {
		f7 := map[string]*string{}
		for f7key, f7valiter := range r.ko.Spec.Tags {
			var f7val string
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		res.SetTags(f7)
	}
	if r.ko.Spec.Version != nil {
		res.SetVersion(*r.ko.Spec.Version)
	}

	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	return rm.customUpdateCluster(ctx, desired, latest, delta)
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer exit(err)
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteClusterOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteClusterWithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteCluster", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteClusterInput, error) {
	res := &svcsdk.DeleteClusterInput{}

	if r.ko.Spec.Name != nil {
		res.SetName(*r.ko.Spec.Name)
	}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.Cluster,
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions(
	r *resource,
	onSuccess bool,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	var recoverableCondition *ackv1alpha1.Condition = nil
	var syncCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeRecoverable {
			recoverableCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			syncCondition = condition
		}
	}

	if rm.terminalAWSError(err) || err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type: ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		var errorMessage = ""
		if err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound {
			errorMessage = err.Error()
		} else {
			awsErr, _ := ackerr.AWSError(err)
			errorMessage = awsErr.Error()
		}
		terminalCondition.Status = corev1.ConditionTrue
		terminalCondition.Message = &errorMessage
	} else {
		// Clear the terminal condition if no longer present
		if terminalCondition != nil {
			terminalCondition.Status = corev1.ConditionFalse
			terminalCondition.Message = nil
		}
		// Handling Recoverable Conditions
		if err != nil {
			if recoverableCondition == nil {
				// Add a new Condition containing a non-terminal error
				recoverableCondition = &ackv1alpha1.Condition{
					Type: ackv1alpha1.ConditionTypeRecoverable,
				}
				ko.Status.Conditions = append(ko.Status.Conditions, recoverableCondition)
			}
			recoverableCondition.Status = corev1.ConditionTrue
			awsErr, _ := ackerr.AWSError(err)
			errorMessage := err.Error()
			if awsErr != nil {
				errorMessage = awsErr.Error()
			}
			recoverableCondition.Message = &errorMessage
		} else if recoverableCondition != nil {
			recoverableCondition.Status = corev1.ConditionFalse
			recoverableCondition.Message = nil
		}
	}
	// Required to avoid the "declared but not used" error in the default case
	_ = syncCondition
	if terminalCondition != nil || recoverableCondition != nil || syncCondition != nil {
		return &resource{ko}, true // updated
	}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
	if err == nil {
		return false
	}
	awsErr, ok := ackerr.AWSError(err)
	if !ok {
		return false
	}
	switch awsErr.Code() {
	case "ResourceLimitExceeded",
		"ResourceNotFound",
		"ResourceInUse",
		"OptInRequired",
		"InvalidParameterCombination",
		"InvalidParameterValue",
		"InvalidParameterException",
		"InvalidQueryParameter",
		"MalformedQueryString",
		"MissingAction",
		"MissingParameter",
		"ValidationError":
		return true
	default:
		return false
	}
}

// newLogging returns a Logging object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newLogging(
	r *resource,
) *svcsdk.Logging {
	res := &svcsdk.Logging{}

	if r.ko.Spec.Logging.ClusterLogging != nil {
		resf0 := []*svcsdk.LogSetup{}
		for _, resf0iter := range r.ko.Spec.Logging.ClusterLogging {
			resf0elem := &svcsdk.LogSetup{}
			if resf0iter.Enabled != nil {
				resf0elem.SetEnabled(*resf0iter.Enabled)
			}
			if resf0iter.Types != nil {
				resf0elemf1 := []*string{}
				for _, resf0elemf1iter := range resf0iter.Types {
					var resf0elemf1elem string
					resf0elemf1elem = *resf0elemf1iter
					resf0elemf1 = append(resf0elemf1, &resf0elemf1elem)
				}
				resf0elem.SetTypes(resf0elemf1)
			}
			resf0 = append(resf0, resf0elem)
		}
		res.SetClusterLogging(resf0)
	}

	return res
}

// newResourcesVpcConfig returns a ResourcesVpcConfig object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newResourcesVpcConfig(
	r *resource,
) *svcsdk.VpcConfigRequest {
	res := &svcsdk.VpcConfigRequest{}

	if r.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess != nil {
		res.SetEndpointPrivateAccess(*r.ko.Spec.ResourcesVPCConfig.EndpointPrivateAccess)
	}
	if r.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess != nil {
		res.SetEndpointPublicAccess(*r.ko.Spec.ResourcesVPCConfig.EndpointPublicAccess)
	}
	if r.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs != nil {
		resf2 := []*string{}
		for _, resf2iter := range r.ko.Spec.ResourcesVPCConfig.PublicAccessCIDRs {
			var resf2elem string
			resf2elem = *resf2iter
			resf2 = append(resf2, &resf2elem)
		}
		res.SetPublicAccessCidrs(resf2)
	}
	if r.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs != nil {
		resf3 := []*string{}
		for _, resf3iter := range r.ko.Spec.ResourcesVPCConfig.SecurityGroupIDs {
			var resf3elem string
			resf3elem = *resf3iter
			resf3 = append(resf3, &resf3elem)
		}
		res.SetSecurityGroupIds(resf3)
	}
	if r.ko.Spec.ResourcesVPCConfig.SubnetIDs != nil {
		resf4 := []*string{}
		for _, resf4iter := range r.ko.Spec.ResourcesVPCConfig.SubnetIDs {
			var resf4elem string
			resf4elem = *resf4iter
			resf4 = append(resf4, &resf4elem)
		}
		res.SetSubnetIds(resf4)
	}

	return res
}
