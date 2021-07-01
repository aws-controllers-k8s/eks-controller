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

"""Integration tests for the EKS FargateProfile resource
"""

import boto3
import datetime
import logging
import time
from typing import Dict

import pytest

from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import CRD_VERSION, service_marker, CRD_GROUP, load_eks_resource
from e2e.common.types import CLUSTER_RESOURCE_PLURAL
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import get_bootstrap_resources

RESOURCE_PLURAL = 'fargateprofiles'

DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture
def simple_cluster():
    cluster_name = random_suffix_name("simple-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name

    resource_data = load_eks_resource(
        "cluster_simple",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, CLUSTER_RESOURCE_PLURAL,
        cluster_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    _, deleted = k8s.delete_custom_resource(ref, 3, 10)
    assert deleted

@service_marker
@pytest.mark.canary
class TestFargateProfile:
    def test_create_delete_fargate_profile(self, simple_cluster, eks_client):
        pass
