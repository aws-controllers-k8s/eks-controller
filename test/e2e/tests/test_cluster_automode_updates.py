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
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.common.types import CLUSTER_RESOURCE_PLURAL
from e2e.common.waiter import wait_until_deleted
from e2e.replacement_values import REPLACEMENT_VALUES

MODIFY_WAIT_AFTER_SECONDS = 240
CHECK_STATUS_WAIT_SECONDS = 240


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

        # Get the nodepool role ARN from bootstrap resources
        nodepool_role = get_bootstrap_resources().NodepoolRole
        logging.info(f"Using nodepool role ARN: {nodepool_role.arn}")

        # Patch to enable auto-mode
        patch_enable_auto_mode = {
            "spec": {
                "computeConfig": {"enabled": True, "nodeRoleARN": nodepool_role.arn},
                "storageConfig": {"blockStorage": {"enabled": True}},
                "kubernetesNetworkConfig": {"elasticLoadBalancing": {"enabled": True}},
            }
        }
        logging.info(f"Applying patch to enable auto-mode: {patch_enable_auto_mode}")
        k8s.patch_custom_resource(ref, patch_enable_auto_mode)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        get_and_assert_status(ref, "UPDATING", False)

        # Wait for cluster to become active after update
        wait_for_cluster_active(eks_client, cluster_name)
        time.sleep(CHECK_STATUS_WAIT_SECONDS)

        get_and_assert_status(ref, "ACTIVE", True)

        # Verify auto-mode is enabled
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["compute"] is not None
        assert aws_res["cluster"]["storage"] is not None
        assert (
            aws_res["cluster"]["kubernetesNetworkConfig"]["elasticLoadBalancing"]
            is not None
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
        assert (
            "compute" not in aws_res["cluster"] or aws_res["cluster"]["compute"] is None
        )
        assert (
            "storage" not in aws_res["cluster"] or aws_res["cluster"]["storage"] is None
        )
        assert (
            "elasticLoadBalancing" not in aws_res["cluster"]["kubernetesNetworkConfig"]
            or aws_res["cluster"]["kubernetesNetworkConfig"]["elasticLoadBalancing"]
            is None
        )
