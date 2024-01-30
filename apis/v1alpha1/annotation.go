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

package v1alpha1

import "fmt"

var (
	// DesiredSizeManagedByAnnotation is the annotation key used to set the management style for the
	// desired size of a nodegroup scaling configuration. This annotation can only be set on a
	// nodegroup custom resource.
	//
	// The value of this annotation must be one of the following:
	//
	// - 'external-autoscaler': The desired size is managed by an external entity. Causing the
	//                          controller to completly ignore the `scalingConfig.desiredSize` field
	// 						    and not reconcile the desired size of a nodegroup.
	//
	// - 'ack-eks-controller':  The desired size is managed by the ACK controller. Causing the
	//                          controller to reconcile the desired size of the nodegroup with the
	//                          value of the `spec.scalingConfig.desiredSize` field.
	//
	// By default the desired size is managed by the controller. If the annotation is not set, or
	// the value is not one of the above, the controller will default to managing the desired size
	// as if the annotation was set to "controller".
	DesiredSizeManagedByAnnotation = fmt.Sprintf("%s/desired-size-managed-by", GroupVersion.Group)
	// ForceNodeGroupUpdateVersionAnnotation is the annotation key used to force an update of the
	// nodegroup version. This annotation can only be set on a nodegroup custom resource.
	// The value of this annotation must be a boolean value. If the value is "true", the controller
	// will force an update of the nodegroup version to the value specified in the `version` field
	// of the `spec` object. If the value is "false", the controller will not force an update of the
	// nodegroup version.
	ForceNodeGroupUpdateVersionAnnotation = fmt.Sprintf("%s/force-update-version", GroupVersion.Group)
)

const (
	// DesiredSizeManagedByExternalAutoscaler is the value of the DesiredSizeManagedByAnnotation
	// annotation that indicates that the desired size of a nodegroup is managed by an external
	// autoscaler.
	DesiredSizeManagedByExternalAutoscaler = "external-autoscaler"
	// DesiredSizeManagedByACKController is the value of the DesiredSizeManagedByAnnotation
	// annotation that indicates that the desired size of a nodegroup is managed by the ACK
	// controller.
	DesiredSizeManagedByACKController = "ack-eks-controller"
)
