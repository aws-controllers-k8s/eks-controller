# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Bootstraps the resources required to run the EKS integration tests.
"""
import logging

from acktest.bootstrapping import ServiceBootstrapResources
from acktest.bootstrapping.iam import Role
from e2e import bootstrap_directory
from e2e.bootstrap_resources import (
    TestBootstrapResources,
)

def service_bootstrap() -> ServiceBootstrapResources:
    logging.getLogger().setLevel(logging.INFO)
    
    resources = TestBootstrapResources(
        ClusterRole=Role("cluster-role", "eks.amazonaws.com", ["arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"]),
        FargatePodRole=Role("fargate-pod-role", "eks-fargate-pods.amazonaws.com", ["arn:aws:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"])
    )

    resources.bootstrap()

    return resources

if __name__ == "__main__":
    config = service_bootstrap()
    # Write config to current directory by default
    config.serialize(bootstrap_directory)