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
import logging
import time

import pytest

from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_eks_resource
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.common.types import CLUSTER_RESOURCE_PLURAL
from e2e.common.waiter import wait_until_deleted
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.fixtures import assert_tagging_functionality
from e2e.common import (
    TESTS_DEFAULT_KUBERNETES_VERSION_1_27,
    TESTS_DEFAULT_KUBERNETES_VERSION_1_29,
)

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 60

# Time to wait after the cluster has changed status, for the CR to update
CHECK_STATUS_WAIT_SECONDS = 30

# Testing KMS Key.
# NOTE(a-hilaly) Can't wait to rewrite all this stuff in Go. The current bootstrapping
# is a mess.
ACK_KMS_KEY_ARN = "arn:aws:kms:us-west-2:632556926448:key/aac8cabd-2a52-43dd-96dc-266c03a9b412"

def wait_for_cluster_active(eks_client, cluster_name):
    waiter = eks_client.get_waiter(
        'cluster_active',
    )
    waiter.config.delay = 5
    waiter.config.max_attempts = 240
    waiter.wait(name=cluster_name)

def get_and_assert_status(ref: k8s.CustomResourceReference, expected_status: str, expected_synced: bool):
    cr = k8s.get_resource(ref)
    assert cr is not None
    assert 'status' in cr
    assert 'status' in cr['status']
    assert cr['status']['status'] == expected_status

    if expected_synced:
        condition.assert_synced(ref)
    else:
        condition.assert_not_synced(ref)

@pytest.fixture(scope="module")
def eks_client():
    return boto3.client('eks')

@pytest.fixture
def simple_cluster(eks_client):
    cluster_name = random_suffix_name("simple-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_29

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
    cr = k8s.wait_resource_consumed_by_controller(ref, wait_periods=15)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    # Try to delete, if doesn't already exist
    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
        wait_until_deleted(cluster_name)
    except:
        pass

@pytest.fixture
def simple_cluster_version_minus_2(eks_client):
    cluster_name = random_suffix_name("simple-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_27

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
    cr = k8s.wait_resource_consumed_by_controller(ref, wait_periods=15)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    # Try to delete, if doesn't already exist
    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
        wait_until_deleted(cluster_name)
    except:
        pass

@service_marker
@pytest.mark.canary
class TestCluster:
    def test_create_update_delete_cluster(self, eks_client, simple_cluster):
        (ref, cr) = simple_cluster

        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")


        wait_for_cluster_active(eks_client, cluster_name)

        # Update VPC endpoint public access config field
        updates = {
            "spec": {
                "resourcesVPCConfig": {
                    "endpointPublicAccess": False
                }
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Ensure status is updating properly and set as not synced
        get_and_assert_status(ref, 'UPDATING', False)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        # Ensure status is updated properly once it has become active
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)

        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["resourcesVpcConfig"]["endpointPublicAccess"] == False

        # Update the VPC subnets config field
        vpc_subnets_ids = get_bootstrap_resources().ClusterVPC.public_subnets.subnet_ids
        # We substitute the first subnet with the last one which is in the same AZ
        subnets_ids = [vpc_subnets_ids[len(vpc_subnets_ids)-1], vpc_subnets_ids[1]]
        
        updates = {
            "spec": {
                "resourcesVPCConfig": {
                    "subnetIDs": subnets_ids,
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_cluster_active(eks_client, cluster_name)

        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert sorted(aws_res["cluster"]["resourcesVpcConfig"]["subnetIds"]) == sorted(subnets_ids)

        # Update the logging fields
        updates = {
            "spec": {
                "logging": {
                    "clusterLogging": [
                        {
                            "enabled": True,
                            "types": ["api"]
                        },
                        {
                            "enabled": False,
                            "types": ["audit", "authenticator", "controllerManager", "scheduler"]
                        },
                    ]
                },
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_cluster_active(eks_client, cluster_name)

        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert len(aws_res["cluster"]["logging"]["clusterLogging"]) > 0
        logging = aws_res["cluster"]["logging"]["clusterLogging"][0]
        assert logging["enabled"] == True
        assert logging["types"] == ["api"]

        # Update the AccessConfig field
        updates = {
            "spec": {
                "accessConfig": {
                    "authenticationMode": "API",
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_cluster_active(eks_client, cluster_name)

        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["accessConfig"]["authenticationMode"] == "API"

        assert_tagging_functionality(ref, cr["status"]["ackResourceMetadata"]["arn"])

        # Delete the k8s resource on teardown of the module
        k8s.delete_custom_resource(ref)
        wait_until_deleted(cluster_name)

    def test_update_cluster_version(self, eks_client, simple_cluster_version_minus_2):
        (ref, cr) = simple_cluster_version_minus_2

        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")


        wait_for_cluster_active(eks_client, cluster_name)

        # Bump two minor versions 1.27 -> 1.29
        updates = {
            "spec": {
                "version": "1.29"
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Ensure status is updating properly and set as not synced
        get_and_assert_status(ref, 'UPDATING', False)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        # At this point, the cluster should be active again at version 1.28
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["version"] == "1.28"

        # So we need to wait again for the CR to be updated.
        time.sleep(CHECK_STATUS_WAIT_SECONDS*1.5)

        # Ensure status is updating properly and set as not synced
        get_and_assert_status(ref, 'UPDATING', False)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        # So we need to wait again for the CR to be updated.
        time.sleep(CHECK_STATUS_WAIT_SECONDS*1.5)
        
        # the cluster should be active again at version 1.29
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["version"] == "1.29"

        # So we need to wait again for the CR to be updated.
        time.sleep(CHECK_STATUS_WAIT_SECONDS*1.5)

        # Ensure status is updating properly and set as not synced
        get_and_assert_status(ref, 'ACTIVE', True)

    def test_associate_cluster_encryption_config(self, eks_client, simple_cluster):
        (ref, cr) = simple_cluster

        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")


        wait_for_cluster_active(eks_client, cluster_name)

        updates = {
            "spec": {
                "encryptionConfig": [
                    {
                        "resources": ["secrets"],
                        "provider": {
                            "keyARN": ACK_KMS_KEY_ARN
                        }
                    }
                ]
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS*2)

        # Ensure status is updating properly and set as not synced
        get_and_assert_status(ref, 'UPDATING', False)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        # At this point, the cluster should be active again at version 1.28
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert len(aws_res["cluster"]["encryptionConfig"]) == 1
        assert aws_res["cluster"]["encryptionConfig"][0]["resources"] == ["secrets"]
        assert aws_res["cluster"]["encryptionConfig"][0]["provider"]["keyArn"] == ACK_KMS_KEY_ARN
