# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for EKS Auto-Mode Cluster updates"""

import boto3
import logging
import time
import pytest

from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e.common import TESTS_DEFAULT_KUBERNETES_VERSION_1_32
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_eks_resource
from e2e.common.types import CLUSTER_RESOURCE_PLURAL
from e2e.common.waiter import wait_until_deleted
from e2e.replacement_values import REPLACEMENT_VALUES

# Time (seconds) to wait after creating/patching the Cluster CR so the controller
# can reconcile and issue any needed AWS API calls before we assert intermediate state.
MODIFY_WAIT_AFTER_SECONDS = 5

# Time (seconds) to wait after EKS DescribeCluster reports the cluster ACTIVE before
# re-reading the CR. This gives the controller a chance to observe the external state
# transition and update the CR status fields.
CHECK_STATUS_WAIT_SECONDS = 30


def wait_for_cluster_active(eks_client, cluster_name):
    waiter = eks_client.get_waiter("cluster_active")
    waiter.config.delay = 5
    waiter.config.max_attempts = 240
    waiter.wait(name=cluster_name)


def get_and_assert_status(
    ref: k8s.CustomResourceReference, expected_status: str, expected_synced: bool
):
    cr = k8s.get_resource(ref)
    assert cr is not None
    assert "status" in cr
    assert "status" in cr["status"]
    assert cr["status"]["status"] == expected_status

    if expected_synced:
        condition.assert_synced(ref)
    else:
        condition.assert_not_synced(ref)


@pytest.fixture(scope="module")
def eks_client():
    return boto3.client("eks")


@pytest.fixture
def simple_cluster(eks_client):
    cluster_name = random_suffix_name("simple-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_32

    resource_data = load_eks_resource(
        "cluster_simple",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP,
        CRD_VERSION,
        CLUSTER_RESOURCE_PLURAL,
        cluster_name,
        namespace="default",
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
    except Exception:
        pass


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
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
        wait_until_deleted(cluster_name)
    except Exception:
        pass


@service_marker
@pytest.mark.canary
class TestAutoModeClusterUpdates:
    def test_enable_auto_mode_on_standard_cluster(self, eks_client, simple_cluster):
        (ref, cr) = simple_cluster
        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, "ACTIVE", True)

        # Patch to enable auto-mode
        patch_enable_auto_mode = {
            "spec": {
                "computeConfig": {"enabled": True},
                "storageConfig": {"blockStorage": {"enabled": True}},
                "kubernetesNetworkConfig": {"elasticLoadBalancing": {"enabled": True}},
            }
        }
        logging.info(f"Applying patch to enable auto-mode: {patch_enable_auto_mode}")
        k8s.patch_custom_resource(ref, patch_enable_auto_mode)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        get_and_assert_status(ref, "UPDATING", False)

        cr_updating = k8s.get_resource(ref)
        eks_describe_updating = eks_client.describe_cluster(name=cluster_name)

        # Wait for cluster to become active after update
        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        get_and_assert_status(ref, "ACTIVE", True)

        cr_update_done = k8s.get_resource(ref)

        # Verify on AWS EKS API that auto-mode is enabled
        aws_res = eks_client.describe_cluster(name=cluster_name)
        logging.info(
            f"custom resource while updating: {cr_updating} ###### eks:DescribeCluster while updating: {eks_describe_updating} ###### custom resource after transitioning to EKS Auto Mode: {cr_update_done} ###### eks:DescribeCluster response: {aws_res}"
        )

        # Check compute config
        compute_config = aws_res["cluster"].get("computeConfig")
        assert compute_config is not None, "computeConfig should be present"
        assert compute_config.get("enabled") is True, (
            f"computeConfig.enabled should be True, got: {compute_config.get('enabled')}"
        )

        # Check storage config
        storage_config = aws_res["cluster"].get("storageConfig")
        assert storage_config is not None, "storageConfig should be present"
        block_storage = storage_config.get("blockStorage", {})
        assert block_storage.get("enabled") is True, (
            f"storageConfig.blockStorage.enabled should be True, got: {block_storage.get('enabled')}"
        )

        # Check elastic load balancing config
        k8s_network_config = aws_res["cluster"].get("kubernetesNetworkConfig", {})
        elb_config = k8s_network_config.get("elasticLoadBalancing")
        assert elb_config is not None, (
            "kubernetesNetworkConfig.elasticLoadBalancing should be present"
        )
        assert elb_config.get("enabled") is True, (
            f"kubernetesNetworkConfig.elasticLoadBalancing.enabled should be True, got: {elb_config.get('enabled')}"
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
        logging.info(
            f"Applying patch with incorrect parameters: {patch_disable_auto_mode_incorrectly}"
        )
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
        logging.info(f"Applying patch to disable auto-mode: {patch_disable_auto_mode}")
        k8s.patch_custom_resource(ref, patch_disable_auto_mode)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        get_and_assert_status(ref, "UPDATING", False)

        # Wait for cluster to become active after update
        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        get_and_assert_status(ref, "ACTIVE", True)

        # Verify auto-mode is disabled
        aws_res = eks_client.describe_cluster(name=cluster_name)

        # Check compute config - should be absent or disabled
        compute_config = aws_res["cluster"].get("computeConfig")
        if compute_config is not None:
            assert compute_config.get("enabled") is False, (
                f"computeConfig.enabled should be False or absent, got: {compute_config.get('enabled')}"
            )

        # Check storage config - should be absent or disabled
        storage_config = aws_res["cluster"].get("storageConfig")
        if storage_config is not None:
            block_storage = storage_config.get("blockStorage", {})
            if block_storage:
                assert block_storage.get("enabled") is False, (
                    f"storageConfig.blockStorage.enabled should be False or absent, got: {block_storage.get('enabled')}"
                )

        # Check elastic load balancing config - should be absent or disabled
        k8s_network_config = aws_res["cluster"].get("kubernetesNetworkConfig", {})
        elb_config = k8s_network_config.get("elasticLoadBalancing")
        if elb_config is not None:
            assert elb_config.get("enabled") is False, (
                f"kubernetesNetworkConfig.elasticLoadBalancing.enabled should be False or absent, got: {elb_config.get('enabled')}"
            )
