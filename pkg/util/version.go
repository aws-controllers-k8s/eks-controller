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

package util

import (
	"fmt"
	"strconv"
	"strings"
)

// NOTE(a-hilaly): We can generalize this function and port it to the aws-controllers-k8s/pkg repository.
// we could also consider importing a SemVer/Upstream library to help us with utility - just keeping things
// simple now + we need some ATTRIBUTION.md generation automation

var (
	// ErrInvalidEKSKubernetesVersion is an error that is returned when the given EKS kubernetes version is invalid.
	ErrInvalidEKSKubernetesVersion = fmt.Errorf("invalid EKS kubernetes version")
	// ErrInvalidEKSKubernetesReleaseVersion is an error that is returned when the given EKS kubernetes release version is invalid.
	ErrInvalidEKSKubernetesReleaseVersion = fmt.Errorf("invalid EKS kubernetes release version")
)

// IncrementVersionMajor increments the minor version of the given EKS kubernetes version
// and returns the new version. It returns an error if the given version is not in the
// expected format.
//
// For example, given "1.16", it returns "1.17"
func IncrementEKSMinorVersion(version string) (string, error) {
	major, minor, err := parseEKSKubernetesVersion(version)
	if err != nil {
		return "", fmt.Errorf("failed to parse EKS kubernetes version: %w", err)
	}

	return fmt.Sprintf("%d.%d", major, minor+1), nil
}

// GetEKSVersionFromReleaseVersion returns the EKS kubernetes version from the given release version.
// It returns an error if the given version is not in the expected format.
//
// For example, given "1.16.8-01012024", it returns "1.16"
func GetEKSVersionFromReleaseVersion(version string) (string, error) {
	// First, we split the version into the kubernetes version and the release version.
	parts := strings.Split(version, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("%w: %s: expected a version of format major.minor.patch-release", ErrInvalidEKSKubernetesReleaseVersion, version)
	}

	kubernetesVersion := parts[0]

	// Then, we split the kubernetes version into its parts.
	parts = strings.Split(kubernetesVersion, ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", fmt.Errorf("invalid release version: %s", version)
	}
	return fmt.Sprintf("%s.%s", parts[0], parts[1]), nil
}

// CompareEKSKubernetesVersions compares two EKS kubernetes versions and returns 0 if they are equal,
// -1 if version1 is less than version2, and 1 if version1 is greater than version2. It returns an
// error if the given versions are not in the expected format.
func CompareEKSKubernetesVersions(version1, version2 string) (int, error) {
	majorVersion1, minorVersion1, err := parseEKSKubernetesVersion(version1)
	if err != nil {
		return 0, fmt.Errorf("failed to parse EKS kubernetes version: %w", err)
	}
	majorVersion2, minorVersion2, err := parseEKSKubernetesVersion(version2)
	if err != nil {
		return 0, fmt.Errorf("failed to parse EKS kubernetes version: %w", err)
	}

	// 1.9 < 2.0
	if majorVersion1 < majorVersion2 {
		return -1, nil
	}
	// 2.0 > 1.9
	if majorVersion1 > majorVersion2 {
		return 1, nil
	}
	// 1.9 < 1.17
	if minorVersion1 < minorVersion2 {
		return -1, nil
	}
	// 1.17 > 1.9
	if minorVersion1 > minorVersion2 {
		return 1, nil
	}
	return 0, nil
}

// parseEKSKubernetesVersion parses the given EKS kubernetes version and returns the major and minor versions.
// It returns an error if the given version is not in the EKS version format (major.minor).
func parseEKSKubernetesVersion(version string) (int, int, error) {
	// First, we split the version into its parts
	parts := strings.Split(version, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, fmt.Errorf("%w: %s: expected a version of format major.minor", ErrInvalidEKSKubernetesVersion, version)
	}

	// Then, we convert the parts to integers
	majorVersion := parts[0]
	minorVersion := parts[1]

	majorVersionInteger, err := strconv.Atoi(majorVersion)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse minor version: %w", err)
	}
	minorVersionInteger, err := strconv.Atoi(minorVersion)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse minor version: %w", err)
	}

	if majorVersionInteger < 0 || minorVersionInteger < 0 {
		return 0, 0, fmt.Errorf("%w: %s: expected positive integers", ErrInvalidEKSKubernetesVersion, version)
	}
	return majorVersionInteger, minorVersionInteger, nil
}
