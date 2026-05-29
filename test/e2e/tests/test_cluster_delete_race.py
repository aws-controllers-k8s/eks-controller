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

"""Integration tests for EKS Cluster deletion behavior.

Validates that when an EKS Auto Mode cluster is deleted, the controller
waits for the cluster to be fully gone (ResourceNotFoundException) before
removing the finalizer. This prevents a race condition where the IAM Role
controller tries to delete the node role while EKS-managed instance profiles
are still attached.

See: https://github.com/aws-controllers-k8s/iam-controller/pull/181
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
    load_eks_resource,
)
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.common.types import CLUSTER_RESOURCE_PLURAL
from e2e.common.waiter import wait_until_deleted
from e2e.replacement_values import REPLACEMENT_VALUES

# Time to wait after issuing delete for the cluster to enter DELETING state
DELETE_WAIT_SECONDS = 30

# Maximum time to wait for finalizer to remain while cluster is DELETING
FINALIZER_CHECK_TIMEOUT_SECONDS = 60 * 5

# Maximum time to wait for the cluster to be fully deleted
FULL_DELETE_TIMEOUT_SECONDS = 60 * 20


def wait_for_cluster_active(eks_client, cluster_name):
    waiter = eks_client.get_waiter('cluster_active')
    waiter.config.delay = 5
    waiter.config.max_attempts = 240
    waiter.wait(name=cluster_name)


def cluster_exists(eks_client, cluster_name):
    """Returns the cluster dict if it exists, None otherwise."""
    try:
        resp = eks_client.describe_cluster(name=cluster_name)
        return resp['cluster']
    except eks_client.exceptions.ResourceNotFoundException:
        return None


def get_instance_profiles_for_role(iam_client, role_name):
    """Returns list of instance profile names attached to the given role."""
    try:
        resp = iam_client.list_instance_profiles_for_role(RoleName=role_name)
        return [ip['InstanceProfileName'] for ip in resp['InstanceProfiles']]
    except iam_client.exceptions.NoSuchEntityException:
        return []


@pytest.fixture(scope="module")
def eks_client():
    return boto3.client('eks')


@pytest.fixture(scope="module")
def iam_client():
    return boto3.client('iam')


@pytest.fixture
def auto_mode_cluster_for_delete(eks_client):
    """Creates an Auto Mode cluster for deletion testing."""
    cluster_name = random_suffix_name("del-race-test", 32)

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

    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref, wait_periods=15)
    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    # Cleanup: ensure cluster is fully deleted if test fails midway
    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        if deleted:
            wait_until_deleted(cluster_name)
    except Exception:
        pass


@service_marker
class TestClusterDeleteWaitsForNotFound:
    """Validates that the EKS controller does NOT remove the finalizer until
    the cluster is fully deleted in AWS (DescribeCluster returns NotFound).

    This prevents the race condition where:
    1. Cluster CR is deleted → DeleteCluster API called
    2. Finalizer removed immediately (old behavior)
    3. IAM Role CR tries to delete the node role
    4. EKS still has instance profiles attached → DeleteConflict
    """

    def test_finalizer_retained_during_cluster_deletion(
        self, eks_client, iam_client, auto_mode_cluster_for_delete
    ):
        (ref, cr) = auto_mode_cluster_for_delete
        cluster_name = cr["spec"]["name"]

        # Wait for cluster to become ACTIVE
        wait_for_cluster_active(eks_client, cluster_name)
        logging.info(f"Cluster {cluster_name} is ACTIVE")

        # Get the node role name from bootstrap resources
        nodepool_role = get_bootstrap_resources().NodepoolRole
        role_name = nodepool_role.name
        logging.info(f"Node role: {role_name}")

        # Verify that EKS Auto Mode has attached instance profiles to the node role
        # (this may take a moment after cluster becomes active)
        # Only track profiles belonging to THIS cluster (the node role is shared
        # across parallel tests, so other clusters may also attach profiles).
        profiles = []
        for _ in range(12):  # wait up to 2 minutes
            all_profiles = get_instance_profiles_for_role(iam_client, role_name)
            profiles = [p for p in all_profiles if cluster_name in p]
            if profiles:
                break
            time.sleep(10)

        logging.info(f"Instance profiles attached to {role_name}: {get_instance_profiles_for_role(iam_client, role_name)}")
        logging.info(f"Instance profiles for this cluster ({cluster_name}): {profiles}")
        # Note: if no profiles found, the test still validates the finalizer behavior

        # === Delete the Cluster CR ===
        k8s.delete_custom_resource(ref, 3, 10)
        logging.info(f"Issued delete for Cluster CR {cluster_name}")

        # Wait a moment for the controller to process the deletion
        time.sleep(DELETE_WAIT_SECONDS)

        # === KEY ASSERTION: finalizer must remain while cluster is DELETING ===
        # The controller should NOT remove the finalizer until DescribeCluster
        # returns ResourceNotFoundException.
        cluster_info = cluster_exists(eks_client, cluster_name)
        if cluster_info is not None:
            assert cluster_info['status'] == 'DELETING', (
                f"Expected cluster to be in DELETING state, got: {cluster_info['status']}"
            )

            # The CR should still exist in Kubernetes (finalizer blocks removal)
            cr_current = k8s.get_resource(ref)
            assert cr_current is not None, (
                "Cluster CR was removed from Kubernetes while AWS cluster is still DELETING! "
                "This means the finalizer was removed prematurely, which causes the "
                "DeleteConflict race condition with IAM Role deletion."
            )
            finalizers = cr_current.get('metadata', {}).get('finalizers', [])
            assert len(finalizers) > 0, (
                "Cluster CR has no finalizers while AWS cluster is still DELETING! "
                "The controller should retain the finalizer until the cluster is fully gone."
            )
            logging.info(
                f"PASS: Finalizer retained while cluster is DELETING. "
                f"Finalizers: {finalizers}"
            )

        # === Wait for cluster to be fully deleted ===
        wait_until_deleted(cluster_name)
        logging.info(f"Cluster {cluster_name} is fully deleted in AWS")

        # === Verify instance profiles are cleaned up ===
        profiles_after = get_instance_profiles_for_role(iam_client, role_name)
        # Only check profiles belonging to THIS cluster (node role is shared)
        our_profiles_after = [p for p in profiles_after if cluster_name in p]
        logging.info(
            f"Instance profiles on {role_name} after deletion: {profiles_after}"
        )
        logging.info(
            f"Instance profiles for this cluster after deletion: {our_profiles_after}"
        )
        # EKS should clean up its instance profiles when the cluster is deleted
        # (the profiles that were created by EKS Auto Mode should be gone)
        for p in profiles:
            assert p not in our_profiles_after, (
                f"Instance profile {p} is still attached to role {role_name} "
                f"after cluster deletion. This would cause DeleteConflict if "
                f"the IAM controller tries to delete the role."
            )

        # === Verify the CR is eventually removed from Kubernetes ===
        # After the cluster is gone, the controller should remove the finalizer
        # and the CR should be garbage collected.
        for _ in range(20):  # wait up to ~5 minutes
            if not k8s.get_resource_exists(ref):
                break
            time.sleep(15)

        assert not k8s.get_resource_exists(ref), (
            "Cluster CR still exists in Kubernetes after AWS cluster is fully deleted. "
            "The controller should remove the finalizer once DescribeCluster returns NotFound."
        )
        logging.info("PASS: CR removed after cluster fully deleted. No race condition.")
