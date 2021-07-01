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

"""Integration tests for the EKS Cluster resource
"""

import boto3
import datetime
import logging
import time
from typing import Dict

import pytest

from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_eks_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import get_bootstrap_resources

RESOURCE_PLURAL = 'clusters'

DELETE_WAIT_AFTER_SECONDS = 10

@service_marker
@pytest.mark.canary
class TestCluster:
    def test_create_delete_cluster(self, eks_client):
        cluster_name = random_suffix_name("simple-cluster", 32)
        
        replacements = REPLACEMENT_VALUES.copy()
        replacements["CLUSTER_NAME"] = cluster_name
        replacements["CLUSTER_ROLE"] = ...
        replacements["SUBNET_1"] = ...
        replacements["SUBNET_2"] = ...

        resource_data = load_eks_resource(
            "cluster_simple",
            additional_replacements=replacements,
        )
        logging.debug(resource_data)

        # Create the k8s resource
        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
            cluster_name, namespace="default",
        )
        k8s.create_custom_resource(ref, resource_data)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        # Delete the k8s resource on teardown of the module
        k8s.delete_custom_resource(ref)

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Cluster should no longer appear in EKS
        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
            pytest.fail(f"Cluster '{cluster_name}' was not deleted from EKS")
        except eks_client.exceptions.ResourceNotFoundException:
            pass
