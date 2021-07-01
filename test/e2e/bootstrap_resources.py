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

"""Declares the structure of the bootstrapped resources and provides a loader
for them.
"""

from dataclasses import dataclass
from acktest.bootstrapping import ServiceBootstrapResources
from acktest.bootstrapping.iam import Role
from acktest.bootstrapping.vpc import VPC
from e2e import bootstrap_directory

@dataclass
class TestBootstrapResources(ServiceBootstrapResources):
    ClusterVPC: VPC
    ClusterRole: Role
    FargatePodRole: Role

_bootstrap_resources = None

def get_bootstrap_resources(bootstrap_file_name: str = "bootstrap.pkl") -> TestBootstrapResources:
    global _bootstrap_resources
    if _bootstrap_resources is None:
        _bootstrap_resources = TestBootstrapResources.deseralize(bootstrap_directory, bootstrap_file_name=bootstrap_file_name)
    return _bootstrap_resources
