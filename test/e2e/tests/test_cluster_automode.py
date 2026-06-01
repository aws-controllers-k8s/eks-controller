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

from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e.common import TESTS_DEFAULT_KUBERNETES_VERSION_1_35
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

CHECK_STATUS_WAIT_SECONDS = 240


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


@pytest.fixture(scope="class")
def auto_mode_cluster(eks_client):
    cluster_name = random_suffix_name("auto-mode-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["CLUSTER_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_35

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

    # Teardown: ensure cluster is cleaned up if a test fails midway.
    # The finalizer-retention test handles its own deletion, but if
    # test_create_auto_mode_cluster fails before deletion is issued,
    # we still need to clean up the AWS resource.
    try:
        if k8s.get_resource_exists(ref):
            k8s.delete_custom_resource(ref, 3, 10)
        wait_until_deleted(cluster_name)
    except Exception as e:
        logging.error(f"Teardown cleanup failed for {cluster_name}: {e}")


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

    def test_finalizer_retained_during_deletion(self, auto_mode_cluster):
        """Validates that the controller retains the finalizer while the
        cluster CR is in DELETING state and only removes it once the
        underlying AWS cluster is fully deleted.

        This prevents a race condition where the IAM controller deletes the
        node role while EKS-managed instance profiles are still attached.
        See: https://github.com/aws-controllers-k8s/iam-controller/pull/181
        """
        (ref, cr) = auto_mode_cluster
        cluster_name = cr["spec"]["name"]

        # Issue delete for the Cluster CR
        k8s.delete_custom_resource(ref, 3, 10)
        logging.info(f"Issued delete for Cluster CR {cluster_name}")

        # Poll the custom resource until its status reflects DELETING.
        # This verifies the controller correctly updates CR status during
        # deletion. For Auto Mode clusters deletion takes 10-15 min, so
        # we will reliably catch DELETING.
        observed_deleting = False
        for _ in range(24):  # up to 4 minutes
            cr_current = k8s.get_resource(ref)
            if cr_current is None:
                pytest.fail(
                    "CR was removed before reaching DELETING status — "
                    "finalizer was released prematurely."
                )
            cr_status = cr_current.get('status', {}).get('status')
            logging.debug(f"Polling CR status: {cr_status}")
            if cr_status == 'DELETING':
                observed_deleting = True
                finalizers = cr_current.get('metadata', {}).get('finalizers', [])
                assert len(finalizers) > 0, (
                    "Cluster CR has no finalizers while CR status is DELETING!"
                )
                logging.info(
                    f"PASS: CR status is DELETING and finalizer retained. "
                    f"Finalizers: {finalizers}"
                )
                break
            time.sleep(10)

        assert observed_deleting, (
            "Never observed CR in DELETING state — either the controller "
            "removed the finalizer prematurely or never set DELETING status."
        )

        # Wait for cluster to be fully deleted in AWS
        wait_until_deleted(cluster_name)
        logging.info(f"Cluster {cluster_name} is fully deleted in AWS")

        # Verify the CR is eventually removed from K8s
        for _ in range(20):  # up to ~5 minutes
            if not k8s.get_resource_exists(ref):
                break
            time.sleep(15)

        assert not k8s.get_resource_exists(ref), (
            "Cluster CR still exists after AWS cluster is fully deleted. "
            "Controller should remove the finalizer once cluster is gone."
        )
