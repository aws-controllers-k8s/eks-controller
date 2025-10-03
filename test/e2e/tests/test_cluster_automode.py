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

"""Integration tests for the EKS Auto-Mode Cluster
"""

import boto3
import logging
import time
import pytest
import json

from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e.common import TESTS_DEFAULT_KUBERNETES_VERSION_1_32
from e2e import (
    service_marker,
    CRD_GROUP,
    CRD_VERSION,
    load_eks_resource
)
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.common.types import CLUSTER_RESOURCE_PLURAL
from e2e.common.waiter import wait_until_deleted
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.test_cluster import simple_cluster

MODIFY_WAIT_AFTER_SECONDS = 60
CHECK_STATUS_WAIT_SECONDS = 30


def wait_for_cluster_active(eks_client, cluster_name):
    waiter = eks_client.get_waiter('cluster_active')
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
def auto_mode_cluster(eks_client):
    cluster_name = random_suffix_name("auto-mode-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["CLUSTER_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_32

    resource_data = load_eks_resource(
        "cluster_automode",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        CLUSTER_RESOURCE_PLURAL,
        cluster_name,
        namespace="default",
    )

    # Create the CR
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref, wait_periods=15)
    assert cr is not None, "Cluster CR was not created in Kubernetes"
    assert k8s.get_resource_exists(ref), "Could not find the Cluster CR in K8s"

    yield (ref, cr)

    # Try to delete, if doesn't already exist
    try:
        _, deleted = k8s.delete_custom_resource(ref, 9, 10)
        assert deleted
        wait_until_deleted(cluster_name)
    except Exception:
        pass

@service_marker
@pytest.mark.canary
class TestAutoModeCluster:
    def test_create_auto_mode_cluster(self, eks_client, auto_mode_cluster):
        (ref, cr) = auto_mode_cluster
        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
            logging.info(f"Initial cluster state: {aws_res}")
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        wait_for_cluster_active(eks_client, cluster_name)

        # Give the cluster some time to fully stabilize
        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        # First verify the cluster is in ACTIVE state
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["status"] == "ACTIVE"
        logging.info(f"Cluster is active: {aws_res}")

        # Get the nodepool role ARN from bootstrap resources
        nodepool_role = get_bootstrap_resources().NodepoolRole
        logging.info(f"Using nodepool role ARN: {nodepool_role.arn}")

        patch_remove_system_pool = {
            "spec": {
                "computeConfig": {
                    "enabled": True,
                    "nodePools": ["general-purpose"],
                    "nodeRoleARN": nodepool_role.arn
                }
            }
        }
        logging.info(f"Applying patch: {patch_remove_system_pool}")
        k8s.patch_custom_resource(ref, patch_remove_system_pool)

        # Wait for cluster to become active after update
        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        # Clean up
        _, deleted = k8s.delete_custom_resource(ref, 9, 10)
        assert deleted
        wait_until_deleted(cluster_name)


@service_marker
@pytest.mark.canary
class TestAutoModeClusterUpdates:
    def test_enable_auto_mode_on_standard_cluster(self, eks_client, simple_cluster):
        (ref, cr) = simple_cluster
        cluster_name = cr["spec"]["name"]

        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res is not None

        # Wait for the cluster to be ACTIVE and let controller refresh status
        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, "ACTIVE", True)

        # Patch to enable auto-mode
        patch_enable_auto_mode = {
            "spec": {
                "computeConfig": {"enabled": True},
                "storageConfig": {"blockStorage": {"enabled": True}},
                "kubernetesNetworkConfig": {
                    "elasticLoadBalancing": {"enabled": True},
                    "ipFamily": "ipv4",
                },
            }
        }
        k8s.patch_custom_resource(ref, patch_enable_auto_mode)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)
        get_and_assert_status(ref, "UPDATING", False)

        # Wait for cluster to become active after update
        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, "ACTIVE", True)

        # Verify auto-mode activation via EKS update history (since DescribeCluster may not reflect the fields immediately)
        updates_summary = eks_client.list_updates(name=cluster_name)

        update_ids = updates_summary.get("updateIds", [])
        assert len(update_ids) == 1, (
            f"Expected exactly 1 update, got {len(update_ids)}: {update_ids}"
        )

        update_id = update_ids[0]
        upd_desc = eks_client.describe_update(name=cluster_name, updateId=update_id)

        update_info = upd_desc["update"]

        # Verify update type and status
        assert update_info["type"] == "AutoModeUpdate", (
            f"Expected AutoModeUpdate, got: {update_info['type']}"
        )
        assert update_info["status"] == "Successful", (
            f"Expected Successful status, got: {update_info['status']}"
        )

    def test_disable_auto_mode_incorrectly(self, eks_client, auto_mode_cluster):
        (ref, cr) = auto_mode_cluster
        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, "ACTIVE", True)

        # Patch with incorrect parameters to disable auto-mode
        patch_disable_auto_mode_incorrectly = {
            "spec": {
                "computeConfig": {"enabled": False},
                "storageConfig": {
                    "blockStorage": {
                        "enabled": True  # Should be False
                    }
                },
                "kubernetesNetworkConfig": {"elasticLoadBalancing": {"enabled": False}},
            }
        }

        k8s.patch_custom_resource(ref, patch_disable_auto_mode_incorrectly)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # The controller should detect the invalid configuration and set a terminal condition.
        terminal_condition = "ACK.Terminal"
        cond = k8s.get_resource_condition(ref, terminal_condition)
        if cond is None:
            pytest.fail(
                f"Failed to find {terminal_condition} condition in resource {ref}"
            )

        cond_status = cond.get("status", None)
        if str(cond_status) != str(True):
            pytest.fail(
                f"Expected {terminal_condition} condition to have status True but found {cond_status}"
            )

        # Verify the error message contains information about invalid Auto Mode configuration
        assert "invalid Auto Mode configuration" in cond.get("message", "")

    def test_disable_auto_mode_correctly(self, eks_client, auto_mode_cluster):
        (ref, cr) = auto_mode_cluster
        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, "ACTIVE", True)

        # Patch to disable auto-mode correctly
        patch_disable_auto_mode = {
            "spec": {
                "computeConfig": {"enabled": False},
                "storageConfig": {"blockStorage": {"enabled": False}},
                "kubernetesNetworkConfig": {"elasticLoadBalancing": {"enabled": False}},
            }
        }

        k8s.patch_custom_resource(ref, patch_disable_auto_mode)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS )
        get_and_assert_status(ref, "UPDATING", False)

        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, "ACTIVE", True)

        # Verify auto-mode is disabled
        aws_res = eks_client.describe_cluster(name=cluster_name)
        compute_config = aws_res["cluster"].get("computeConfig")
        if compute_config is not None:
            assert compute_config.get("enabled") is False, (
                f"computeConfig.enabled should be False or absent, got: {compute_config.get('enabled')}"
            )

        storage_config = aws_res["cluster"].get("storageConfig")
        if storage_config is not None:
            block_storage = storage_config.get("blockStorage", {})
            if block_storage:
                assert block_storage.get("enabled") is False, (
                    f"storageConfig.blockStorage.enabled should be False or absent, got: {block_storage.get('enabled')}"
                )

        k8s_network_config = aws_res["cluster"].get("kubernetesNetworkConfig", {})
        elb_config = k8s_network_config.get("elasticLoadBalancing")
        if elb_config is not None:
            assert elb_config.get("enabled") is False, (
                f"kubernetesNetworkConfig.elasticLoadBalancing.enabled should be False or absent, got: {elb_config.get('enabled')}"
            )
