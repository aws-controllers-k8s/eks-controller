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

package addon

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"

	svcapitypes "github.com/aws-controllers-k8s/eks-controller/apis/v1alpha1"
)

// helper to build a *resource with only ConfigurationValues set.
func addonWithConfigurationValues(configValues *string) *resource {
	return &resource{
		ko: &svcapitypes.Addon{
			Spec: svcapitypes.AddonSpec{
				Name:                aws.String("test-addon"),
				ClusterName:         aws.String("test-cluster"),
				ConfigurationValues: configValues,
			},
		},
	}
}

func TestNewResourceDelta_ConfigurationValues_JSONComparison(t *testing.T) {
	tests := []struct {
		name     string
		desired  *string
		latest   *string
		wantDiff bool
	}{
		{
			name:     "both nil produces no diff",
			desired:  nil,
			latest:   nil,
			wantDiff: false,
		},
		{
			name:     "desired nil latest non-nil produces diff",
			desired:  nil,
			latest:   aws.String(`{"key":"value"}`),
			wantDiff: true,
		},
		{
			name:     "desired non-nil latest nil produces diff",
			desired:  aws.String(`{"key":"value"}`),
			latest:   nil,
			wantDiff: true,
		},
		{
			name:     "identical JSON strings produce no diff",
			desired:  aws.String(`{"key":"value"}`),
			latest:   aws.String(`{"key":"value"}`),
			wantDiff: false,
		},
		{
			name:     "trailing newline from YAML block scalar produces no diff (issue 2869)",
			desired:  aws.String("{\n  \"secrets-store-csi-driver\": {\n    \"syncSecret\": {\n      \"enabled\": true\n    }\n  }\n}\n"),
			latest:   aws.String("{\n  \"secrets-store-csi-driver\": {\n    \"syncSecret\": {\n      \"enabled\": true\n    }\n  }\n}"),
			wantDiff: false,
		},
		{
			name:     "pretty vs compact JSON produces no diff (issue 2877)",
			desired:  aws.String("{\n  \"notificationOrigin\": [\n    \"PRODUCT\"\n  ]\n}\n"),
			latest:   aws.String(`{"notificationOrigin":["PRODUCT"]}`),
			wantDiff: false,
		},
		{
			name:     "different key ordering produces no diff",
			desired:  aws.String(`{"b":2,"a":1}`),
			latest:   aws.String(`{"a":1,"b":2}`),
			wantDiff: false,
		},
		{
			name:    "whitespace differences produce no diff",
			desired: aws.String(`{"key":"value","nested":{"a":1}}`),
			latest: aws.String(`{
				"key": "value",
				"nested": {
					"a": 1
				}
			}`),
			wantDiff: false,
		},
		{
			name:     "different values produce diff",
			desired:  aws.String(`{"key":"value1"}`),
			latest:   aws.String(`{"key":"value2"}`),
			wantDiff: true,
		},
		{
			name:     "extra key produces diff",
			desired:  aws.String(`{"a":1}`),
			latest:   aws.String(`{"a":1,"b":2}`),
			wantDiff: true,
		},
		{
			name:     "missing key produces diff",
			desired:  aws.String(`{"a":1,"b":2}`),
			latest:   aws.String(`{"a":1}`),
			wantDiff: true,
		},
		{
			name:     "different nested values produce diff",
			desired:  aws.String(`{"syncSecret":{"enabled":true}}`),
			latest:   aws.String(`{"syncSecret":{"enabled":false}}`),
			wantDiff: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired := addonWithConfigurationValues(tt.desired)
			latest := addonWithConfigurationValues(tt.latest)

			delta := newResourceDelta(desired, latest)
			hasDiff := delta.DifferentAt("Spec.ConfigurationValues")

			assert.Equal(t, tt.wantDiff, hasDiff,
				"DifferentAt(Spec.ConfigurationValues) = %v, want %v", hasDiff, tt.wantDiff)
		})
	}
}
