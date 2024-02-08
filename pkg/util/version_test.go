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

import "testing"

func TestIncrementEKSMinorVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty string",
			args{version: ""},
			"",
			true,
		},
		{
			"invalid version - no minor",
			args{version: "1."},
			"",
			true,
		},
		{
			"invalid version - no major",
			args{version: ".16"},
			"",
			true,
		},
		{
			"invalid version - no major and minor",
			args{version: "."},
			"",
			true,
		},
		{
			"invalid version - patch versions",
			args{version: "1.16.8"},
			"",
			true,
		},
		{
			"valid version - one digit",
			args{version: "1.0"},
			"1.1",
			false,
		},
		{
			"valid version - 2 digits",
			args{version: "1.16"},
			"1.17",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IncrementEKSMinorVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementEKSMinorVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IncrementEKSMinorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEKSVersionFromReleaseVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty string",
			args{version: ""},
			"",
			true,
		},
		{
			"invalid version - minor",
			args{version: "1."},
			"",
			true,
		},
		{
			"invalid version - no patch",
			args{version: "1.16."},
			"",
			true,
		},
		{
			"invalid version - no release date",
			args{version: "1.16.8-"},
			"",
			true,
		},
		{
			"valid version",
			args{version: "1.16.8-01012024"},
			"1.16",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEKSVersionFromReleaseVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEKSVersionFromReleaseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEKSVersionFromReleaseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareEKSKubernetesVersions(t *testing.T) {
	type args struct {
		version1 string
		version2 string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			"empty string",
			args{version1: "", version2: ""},
			0,
			true,
		},
		{
			"invalid version - no minor",
			args{version1: "1.", version2: "1.16"},
			0,
			true,
		},
		{
			"invalid version - no major",
			args{version1: ".16", version2: "1.16"},
			0,
			true,
		},
		{
			"invalid version - no major and minor",
			args{version1: ".", version2: "1.16"},
			0,
			true,
		},
		{
			"invalid version - patch versions",
			args{version1: "1.16.8", version2: "1.16"},
			0,
			true,
		},
		{
			"valid version - equal",
			args{version1: "1.16", version2: "1.16"},
			0,
			false,
		},
		{
			"valid version - version1 < version2",
			args{version1: "1.16", version2: "1.17"},
			-1,
			false,
		},
		{
			"valid version - version1 > version2",
			args{version1: "1.17", version2: "1.16"},
			1,
			false,
		},
		{
			"valid version - major version1 < major version2",
			args{version1: "1.16", version2: "2.0"},
			-1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareEKSKubernetesVersions(tt.args.version1, tt.args.version2)
			if (err != nil) != tt.wantErr {
				t.Errorf("CompareEKSKubernetesVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CompareEKSKubernetesVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEKSKubernetesVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   int
		wantErr bool
	}{
		{
			"empty string",
			args{version: ""},
			0,
			0,
			true,
		},
		{
			"invalid version - no minor",
			args{version: "1."},
			0,
			0,
			true,
		},
		{
			"invalid version - no major",
			args{version: ".16"},
			0,
			0,
			true,
		},
		{
			"invalid version - no major and minor",
			args{version: "."},
			0,
			0,
			true,
		},
		{
			"invalid version - patch versions",
			args{version: "1.16.8"},
			0,
			0,
			true,
		},
		{
			"invalid version - negative major",
			args{version: "-1.16"},
			0,
			0,
			true,
		},
		{
			"valid version",
			args{version: "1.16"},
			1,
			16,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseEKSKubernetesVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEKSKubernetesVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseEKSKubernetesVersion() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseEKSKubernetesVersion() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
