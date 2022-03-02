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

package fargate_profile

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
	_ = &svcapitypes.FargateProfile{}
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

	var resp *svcsdk.DescribeFargateProfileOutput
	resp, err = rm.sdkapi.DescribeFargateProfileWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_ONE", "DescribeFargateProfile", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "ResourceNotFoundException" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	if resp.FargateProfile.ClusterName != nil {
		ko.Spec.ClusterName = resp.FargateProfile.ClusterName
	} else {
		ko.Spec.ClusterName = nil
	}
	if resp.FargateProfile.CreatedAt != nil {
		ko.Status.CreatedAt = &metav1.Time{*resp.FargateProfile.CreatedAt}
	} else {
		ko.Status.CreatedAt = nil
	}
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.FargateProfile.FargateProfileArn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.FargateProfile.FargateProfileArn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.FargateProfile.FargateProfileName != nil {
		ko.Spec.Name = resp.FargateProfile.FargateProfileName
	} else {
		ko.Spec.Name = nil
	}
	if resp.FargateProfile.PodExecutionRoleArn != nil {
		ko.Spec.PodExecutionRoleARN = resp.FargateProfile.PodExecutionRoleArn
	} else {
		ko.Spec.PodExecutionRoleARN = nil
	}
	if resp.FargateProfile.Selectors != nil {
		f5 := []*svcapitypes.FargateProfileSelector{}
		for _, f5iter := range resp.FargateProfile.Selectors {
			f5elem := &svcapitypes.FargateProfileSelector{}
			if f5iter.Labels != nil {
				f5elemf0 := map[string]*string{}
				for f5elemf0key, f5elemf0valiter := range f5iter.Labels {
					var f5elemf0val string
					f5elemf0val = *f5elemf0valiter
					f5elemf0[f5elemf0key] = &f5elemf0val
				}
				f5elem.Labels = f5elemf0
			}
			if f5iter.Namespace != nil {
				f5elem.Namespace = f5iter.Namespace
			}
			f5 = append(f5, f5elem)
		}
		ko.Spec.Selectors = f5
	} else {
		ko.Spec.Selectors = nil
	}
	if resp.FargateProfile.Status != nil {
		ko.Status.Status = resp.FargateProfile.Status
	} else {
		ko.Status.Status = nil
	}
	if resp.FargateProfile.Subnets != nil {
		f7 := []*string{}
		for _, f7iter := range resp.FargateProfile.Subnets {
			var f7elem string
			f7elem = *f7iter
			f7 = append(f7, &f7elem)
		}
		ko.Spec.Subnets = f7
	} else {
		ko.Spec.Subnets = nil
	}
	if resp.FargateProfile.Tags != nil {
		f8 := map[string]*string{}
		for f8key, f8valiter := range resp.FargateProfile.Tags {
			var f8val string
			f8val = *f8valiter
			f8[f8key] = &f8val
		}
		ko.Spec.Tags = f8
	} else {
		ko.Spec.Tags = nil
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
	return r.ko.Spec.ClusterName == nil || r.ko.Spec.Name == nil

}

// newDescribeRequestPayload returns SDK-specific struct for the HTTP request
// payload of the Describe API call for the resource
func (rm *resourceManager) newDescribeRequestPayload(
	r *resource,
) (*svcsdk.DescribeFargateProfileInput, error) {
	res := &svcsdk.DescribeFargateProfileInput{}

	if r.ko.Spec.ClusterName != nil {
		res.SetClusterName(*r.ko.Spec.ClusterName)
	}
	if r.ko.Spec.Name != nil {
		res.SetFargateProfileName(*r.ko.Spec.Name)
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

	var resp *svcsdk.CreateFargateProfileOutput
	_ = resp
	resp, err = rm.sdkapi.CreateFargateProfileWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateFargateProfile", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.FargateProfile.ClusterName != nil {
		ko.Spec.ClusterName = resp.FargateProfile.ClusterName
	} else {
		ko.Spec.ClusterName = nil
	}
	if resp.FargateProfile.CreatedAt != nil {
		ko.Status.CreatedAt = &metav1.Time{*resp.FargateProfile.CreatedAt}
	} else {
		ko.Status.CreatedAt = nil
	}
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if resp.FargateProfile.FargateProfileArn != nil {
		arn := ackv1alpha1.AWSResourceName(*resp.FargateProfile.FargateProfileArn)
		ko.Status.ACKResourceMetadata.ARN = &arn
	}
	if resp.FargateProfile.FargateProfileName != nil {
		ko.Spec.Name = resp.FargateProfile.FargateProfileName
	} else {
		ko.Spec.Name = nil
	}
	if resp.FargateProfile.PodExecutionRoleArn != nil {
		ko.Spec.PodExecutionRoleARN = resp.FargateProfile.PodExecutionRoleArn
	} else {
		ko.Spec.PodExecutionRoleARN = nil
	}
	if resp.FargateProfile.Selectors != nil {
		f5 := []*svcapitypes.FargateProfileSelector{}
		for _, f5iter := range resp.FargateProfile.Selectors {
			f5elem := &svcapitypes.FargateProfileSelector{}
			if f5iter.Labels != nil {
				f5elemf0 := map[string]*string{}
				for f5elemf0key, f5elemf0valiter := range f5iter.Labels {
					var f5elemf0val string
					f5elemf0val = *f5elemf0valiter
					f5elemf0[f5elemf0key] = &f5elemf0val
				}
				f5elem.Labels = f5elemf0
			}
			if f5iter.Namespace != nil {
				f5elem.Namespace = f5iter.Namespace
			}
			f5 = append(f5, f5elem)
		}
		ko.Spec.Selectors = f5
	} else {
		ko.Spec.Selectors = nil
	}
	if resp.FargateProfile.Status != nil {
		ko.Status.Status = resp.FargateProfile.Status
	} else {
		ko.Status.Status = nil
	}
	if resp.FargateProfile.Subnets != nil {
		f7 := []*string{}
		for _, f7iter := range resp.FargateProfile.Subnets {
			var f7elem string
			f7elem = *f7iter
			f7 = append(f7, &f7elem)
		}
		ko.Spec.Subnets = f7
	} else {
		ko.Spec.Subnets = nil
	}
	if resp.FargateProfile.Tags != nil {
		f8 := map[string]*string{}
		for f8key, f8valiter := range resp.FargateProfile.Tags {
			var f8val string
			f8val = *f8valiter
			f8[f8key] = &f8val
		}
		ko.Spec.Tags = f8
	} else {
		ko.Spec.Tags = nil
	}

	rm.setStatusDefaults(ko)
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateFargateProfileInput, error) {
	res := &svcsdk.CreateFargateProfileInput{}

	if r.ko.Spec.ClientRequestToken != nil {
		res.SetClientRequestToken(*r.ko.Spec.ClientRequestToken)
	}
	if r.ko.Spec.ClusterName != nil {
		res.SetClusterName(*r.ko.Spec.ClusterName)
	}
	if r.ko.Spec.Name != nil {
		res.SetFargateProfileName(*r.ko.Spec.Name)
	}
	if r.ko.Spec.PodExecutionRoleARN != nil {
		res.SetPodExecutionRoleArn(*r.ko.Spec.PodExecutionRoleARN)
	}
	if r.ko.Spec.Selectors != nil {
		f4 := []*svcsdk.FargateProfileSelector{}
		for _, f4iter := range r.ko.Spec.Selectors {
			f4elem := &svcsdk.FargateProfileSelector{}
			if f4iter.Labels != nil {
				f4elemf0 := map[string]*string{}
				for f4elemf0key, f4elemf0valiter := range f4iter.Labels {
					var f4elemf0val string
					f4elemf0val = *f4elemf0valiter
					f4elemf0[f4elemf0key] = &f4elemf0val
				}
				f4elem.SetLabels(f4elemf0)
			}
			if f4iter.Namespace != nil {
				f4elem.SetNamespace(*f4iter.Namespace)
			}
			f4 = append(f4, f4elem)
		}
		res.SetSelectors(f4)
	}
	if r.ko.Spec.Subnets != nil {
		f5 := []*string{}
		for _, f5iter := range r.ko.Spec.Subnets {
			var f5elem string
			f5elem = *f5iter
			f5 = append(f5, &f5elem)
		}
		res.SetSubnets(f5)
	}
	if r.ko.Spec.Tags != nil {
		f6 := map[string]*string{}
		for f6key, f6valiter := range r.ko.Spec.Tags {
			var f6val string
			f6val = *f6valiter
			f6[f6key] = &f6val
		}
		res.SetTags(f6)
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
	return rm.customUpdate(ctx, desired, latest, delta)
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer exit(err)
	if profileDeleting(r) {
		return r, requeueWaitWhileDeleting
	}

	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteFargateProfileOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteFargateProfileWithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteFargateProfile", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteFargateProfileInput, error) {
	res := &svcsdk.DeleteFargateProfileInput{}

	if r.ko.Spec.ClusterName != nil {
		res.SetClusterName(*r.ko.Spec.ClusterName)
	}
	if r.ko.Spec.Name != nil {
		res.SetFargateProfileName(*r.ko.Spec.Name)
	}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.FargateProfile,
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
