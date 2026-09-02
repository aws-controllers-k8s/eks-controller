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
    TESTS_DEFAULT_KUBERNETES_VERSION_1_33,
    TESTS_DEFAULT_KUBERNETES_VERSION_1_34,
    TESTS_DEFAULT_KUBERNETES_VERSION_1_35,
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

def get_failed_version_update(eks_client, cluster_name):
    """Return a FAILED VersionUpdate for the cluster, or None.

    Scoping/ordering notes:
    - The version test creates a dedicated cluster (function-scoped fixture,
      unique random name) that no other test touches, so every update on this
      cluster belongs to this test. We therefore don't need to identify a
      specific updateId.
    - EKS does not guarantee an ordering for updateIds, so we scan all of them
      and match on type/status rather than assuming the newest is first.

    A failed control-plane version upgrade leaves the cluster ACTIVE at its
    previous version with the Update resource marked 'Failed'. Detecting this
    lets the test fail fast (with the EKS error) instead of waiting out the
    full upgrade timeout.
    """
    update_ids = eks_client.list_updates(name=cluster_name).get("updateIds", [])
    for update_id in update_ids:
        update = eks_client.describe_update(
            name=cluster_name, updateId=update_id,
        )["update"]
        if update.get("type") == "VersionUpdate" and update.get("status") == "Failed":
            return update
    return None

def aws_control_plane_egress_mode(eks_client, cluster_name):
    """Returns the controlPlaneEgressMode EKS reports for the cluster, or None
    when the installed botocore is too old to model the member.

    botocore silently drops response members it does not know about, so on an
    older release the key is simply absent rather than raising. Returning None
    lets a test keep its CR-side assertions and skip only the AWS-side
    comparison, instead of failing for an unrelated reason.
    """
    vpc_config = eks_client.describe_cluster(
        name=cluster_name)["cluster"]["resourcesVpcConfig"]
    mode = vpc_config.get("controlPlaneEgressMode")
    if mode is None:
        logging.warning(
            "installed botocore does not report "
            "resourcesVpcConfig.controlPlaneEgressMode; skipping the AWS-side "
            "assertion. Upgrade botocore for full dual verification."
        )
    return mode

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
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_35

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
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_33

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
def control_plane_scaling_cluster(eks_client):
    cluster_name = random_suffix_name("cps-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_35
    replacements["CONTROL_PLANE_SCALING_TIER"] = "standard"

    resource_data = load_eks_resource(
        "cluster_control_plane_scaling",
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
def control_plane_egress_mode_cluster(eks_client):
    cluster_name = random_suffix_name("cpem-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_35
    # Non-default on purpose: AWS_MANAGED at create time is indistinguishable
    # from the server default, so it would not prove the field was transmitted.
    replacements["CONTROL_PLANE_EGRESS_MODE"] = "CUSTOMER_ROUTED"

    resource_data = load_eks_resource(
        "cluster_control_plane_egress_mode",
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
def component_config_cluster(eks_client):
    cluster_name = random_suffix_name("cc-cluster", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_35
    replacements["EVENT_TTL"] = "30m"
    replacements["HPA_SYNC_PERIOD"] = "10s"
    replacements["TERMINATED_POD_GC_THRESHOLD"] = "12500"

    resource_data = load_eks_resource(
        "cluster_component_config",
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
def partial_component_config_cluster(eks_client):
    cluster_name = random_suffix_name("cc-partial", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_35
    replacements["EVENT_TTL"] = "45m"

    resource_data = load_eks_resource(
        "cluster_component_config_partial",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, CLUSTER_RESOURCE_PLURAL,
        cluster_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref, wait_periods=15)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    try:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
        wait_until_deleted(cluster_name)
    except:
        pass

@pytest.fixture
def adoption_cluster(eks_client):
    adopted_cluster = get_bootstrap_resources().AdoptionCluster
    cluster_name = adopted_cluster.name
    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_ADOPTION_NAME"] = cluster_name
    replacements["ADOPTION_POLICY"] = "adopt"
    replacements["ADOPTION_FIELDS"] = f"{{\\\"name\\\": \\\"{cluster_name}\\\"}}"

    resource_data = load_eks_resource(
        "cluster_adoption",
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

    _, deleted = k8s.delete_custom_resource(ref, 3, 10)
    assert deleted


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

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

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

        # Bump two minor versions: 1.33 -> 1.35.
        #
        # EKS only allows upgrading one minor version at a time, so the
        # controller intentionally increments a single minor version per
        # reconcile (1.33 -> 1.34 -> 1.35). This means reaching the desired
        # version requires TWO sequential control-plane upgrades. We therefore
        # poll until the cluster actually reports the target version rather than
        # asserting after a single active cycle (which would catch the cluster
        # mid-way at 1.34).
        updates = {
            "spec": {
                "version": TESTS_DEFAULT_KUBERNETES_VERSION_1_35
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Drive the cluster through each upgrade hop until it reaches 1.35.
        #
        # We poll describe_cluster directly (rather than wait_for_cluster_active)
        # so the per-hop budget below is the single authoritative bound: the
        # shared waiter caps at 20 min, which is shorter than a single EKS minor
        # upgrade can take, so relying on it here would fail healthy-but-slow
        # hops before the budget applied.
        #
        # HOP_TIMEOUT_SECONDS is the max time allowed for one hop to complete.
        # A small/empty EKS cluster minor upgrade typically finishes in
        # ~10-25 min; 40 min is a padded ceiling so a slow-but-healthy upgrade
        # doesn't flake. The timer resets every time the cluster advances a
        # minor version, so the legitimate two-hop path gets the full budget per
        # hop.
        #
        # We fail fast in two cases so a broken upgrade can't hang the test:
        #   1. A hop stalls with no version progress past HOP_TIMEOUT_SECONDS.
        #   2. EKS reports a FAILED VersionUpdate while the cluster is ACTIVE and
        #      hasn't advanced (e.g. 1.34 -> 1.35 fails and rolls back to ACTIVE
        #      at 1.34). The failed-update history is only consulted while ACTIVE
        #      and not progressing, so a hop still in flight isn't misreported.
        HOP_TIMEOUT_SECONDS = 40 * 60
        POLL_INTERVAL_SECONDS = 30

        cluster_version = eks_client.describe_cluster(
            name=cluster_name,
        )["cluster"]["version"]
        hop_deadline = time.time() + HOP_TIMEOUT_SECONDS

        while cluster_version != TESTS_DEFAULT_KUBERNETES_VERSION_1_35:
            if time.time() > hop_deadline:
                pytest.fail(
                    f"Cluster '{cluster_name}' stalled at version "
                    f"{cluster_version}; did not progress toward "
                    f"{TESTS_DEFAULT_KUBERNETES_VERSION_1_35} within "
                    f"{HOP_TIMEOUT_SECONDS // 60} minutes."
                )

            aws_res = eks_client.describe_cluster(name=cluster_name)
            status = aws_res["cluster"]["status"]
            observed_version = aws_res["cluster"]["version"]

            if status == "ACTIVE" and observed_version != cluster_version:
                # Advanced to the next minor version; reset the hop timer and
                # let the controller pick up the remaining delta.
                cluster_version = observed_version
                hop_deadline = time.time() + HOP_TIMEOUT_SECONDS
                continue

            if status == "ACTIVE":
                # Cluster is settled at the same version. If EKS recorded a
                # failed version upgrade, surface it now instead of waiting out
                # the stall timeout.
                failed = get_failed_version_update(eks_client, cluster_name)
                if failed is not None:
                    pytest.fail(
                        f"EKS version upgrade failed for cluster "
                        f"'{cluster_name}' at version {cluster_version}: "
                        f"{failed.get('errors')}"
                    )

            # Either an upgrade is in flight (UPDATING) or we're waiting for the
            # controller to kick off the next hop. Keep polling under the budget.
            time.sleep(POLL_INTERVAL_SECONDS)

        # the cluster should be active again at the desired version 1.35
        assert cluster_version == TESTS_DEFAULT_KUBERNETES_VERSION_1_35

        # Once the final version is reached, the controller should converge
        # and mark the resource synced.
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)

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

        # At this point, the cluster should be active again at version 1.31
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert len(aws_res["cluster"]["encryptionConfig"]) == 1
        assert aws_res["cluster"]["encryptionConfig"][0]["resources"] == ["secrets"]
        assert aws_res["cluster"]["encryptionConfig"][0]["provider"]["keyArn"] == ACK_KMS_KEY_ARN

    def test_update_cluster_update_policy(self, eks_client, simple_cluster):
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
                "upgradePolicy": {
                    "supportType": "STANDARD"
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS*2)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        # At this point, the cluster should be active again at version 1.32
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["upgradePolicy"]["supportType"] == "STANDARD"
    
    def test_cluster_adopt_update(self, eks_client, adoption_cluster):
        (ref, cr) = adoption_cluster

        assert 'spec' in cr
        assert 'name' in cr['spec']
        cluster_name = cr["spec"]["name"]

        wait_for_cluster_active(eks_client, cluster_name)

        assert 'upgradePolicy' in cr['spec']
        assert 'supportType' in cr['spec']['upgradePolicy']
        support_type = cr['spec']['upgradePolicy']['supportType']
        if support_type == 'STANDARD':
            support_type = 'EXTENDED'
        else:
            support_type = 'STANDARD'

        # Update the cluster name
        updates = {
            "spec": {
                "upgradePolicy": {
                    "supportType": support_type
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS*2)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        # At this point, the cluster should be active again at version 1.31
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res
        assert aws_res["cluster"]["upgradePolicy"]["supportType"] == support_type

    def test_update_cluster_control_plane_scaling_config(self, eks_client, control_plane_scaling_cluster):
        (ref, cr) = control_plane_scaling_cluster

        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        wait_for_cluster_active(eks_client, cluster_name)

        # Verify the cluster was created with the expected tier
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["controlPlaneScalingConfig"]["tier"] == "standard"

        # Update the control plane scaling config tier
        updates = {
            "spec": {
                "controlPlaneScalingConfig": {
                    "tier": "tier-xl"
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        get_and_assert_status(ref, 'ACTIVE', True)

        # Verify via describe_cluster that the tier has been updated
        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["controlPlaneScalingConfig"]["tier"] == "tier-xl"

    def test_create_update_cluster_component_config(self, eks_client, component_config_cluster):
        (ref, cr) = component_config_cluster

        cluster_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_cluster(name=cluster_name)
            assert aws_res is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find cluster '{cluster_name}' in EKS")

        wait_for_cluster_active(eks_client, cluster_name)

        # Verify the control plane component configs were applied at create time.
        # NOTE: boto3/describe_cluster returns the API wire names, which differ
        # from the ACK spec names: kubeApiServerConfig (lowercase "pi") vs the
        # ACK spec's kubeAPIServerConfig.
        aws_res = eks_client.describe_cluster(name=cluster_name)
        cluster = aws_res["cluster"]
        assert cluster["kubeApiServerConfig"]["eventTtl"] == "30m"
        assert cluster["kubeApiServerConfig"]["serviceNodePortRange"]["minPort"] == 30000
        assert cluster["kubeApiServerConfig"]["serviceNodePortRange"]["maxPort"] == 32000
        assert cluster["kubeControllerManagerConfig"]["horizontalPodAutoscalerControllerConfig"]["horizontalPodAutoscalerSyncPeriod"] == "10s"
        assert cluster["kubeControllerManagerConfig"]["podGcControllerConfig"]["terminatedPodGcThreshold"] == 12500
        assert cluster["kubeSchedulerConfig"]["nodeResourcesFit"]["scoringStrategy"]["type"] == "LeastAllocated"

        # Update the component configs. The controller batches every changed
        # component config into a single UpdateClusterConfig call
        # (updateComponentConfig).
        updates = {
            "spec": {
                "kubeAPIServerConfig": {
                    "eventTTL": "1h",
                    "serviceNodePortRange": {
                        "minPort": 30000,
                        "maxPort": 32000,
                    },
                },
                "kubeControllerManagerConfig": {
                    "horizontalPodAutoscalerControllerConfig": {
                        "horizontalPodAutoscalerSyncPeriod": "15s",
                    },
                    "podGcControllerConfig": {
                        "terminatedPodGcThreshold": 12000,
                    },
                },
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Wait for the updating to become active again
        wait_for_cluster_active(eks_client, cluster_name)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)

        get_and_assert_status(ref, 'ACTIVE', True)

        aws_res = eks_client.describe_cluster(name=cluster_name)
        cluster = aws_res["cluster"]
        assert cluster["kubeApiServerConfig"]["eventTtl"] == "1h"
        assert cluster["kubeControllerManagerConfig"]["horizontalPodAutoscalerControllerConfig"]["horizontalPodAutoscalerSyncPeriod"] == "15s"
        assert cluster["kubeControllerManagerConfig"]["podGcControllerConfig"]["terminatedPodGcThreshold"] == 12000

    def test_cluster_component_config_late_initialize(self, eks_client, simple_cluster):
        # This cluster is created WITHOUT any control plane component config in
        # its spec. The EKS backend injects tier-based defaults for these configs
        # and returns them on the read (DescribeCluster) path. late_initialize on
        # KubeAPIServerConfig / KubeSchedulerConfig / KubeControllerManagerConfig
        # must copy those backend defaults into the spec so the controller does
        # NOT treat them as drift and reconcile forever.
        #
        # This test validates a load-bearing assumption of the late_initialize
        # design: the generated incompleteLateInitialization() treats a nil
        # component config as "late-init not complete" and requeues. If EKS ever
        # returns these configs as null (e.g. unset, or tier-gated), late-init
        # would never complete and the resource would never reach Synced=True.
        # A hang here (rather than a value mismatch) points at that assumption.
        (ref, cr) = simple_cluster

        cluster_name = cr["spec"]["name"]
        wait_for_cluster_active(eks_client, cluster_name)

        # The resource must converge to Synced=True even though the spec set no
        # component configs (late-init populated them from the backend defaults).
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)
        get_and_assert_status(ref, 'ACTIVE', True)

        # The spec should now carry the late-initialized backend defaults.
        cr = k8s.get_resource(ref)
        assert cr["spec"].get("kubeAPIServerConfig") is not None
        assert cr["spec"].get("kubeSchedulerConfig") is not None
        assert cr["spec"].get("kubeControllerManagerConfig") is not None

        # Confirm there is no perpetual drift: the resource stays Synced across a
        # settle window. A thrashing late-init would flip Synced back to False.
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)

    def test_cluster_component_config_partial_late_initialize(self, eks_client, partial_component_config_cluster):
        # This cluster sets ONLY kubeAPIServerConfig.eventTTL. The EKS backend
        # fills tier defaults for every omitted field and returns the complete
        # effective config on read. A top-level late_initialize alone would not
        # help here: kubeAPIServerConfig is non-nil (eventTTL is set), so the
        # controller would treat the server-defaulted sibling serviceNodePortRange
        # -- and the entirely-omitted kubeSchedulerConfig / kubeControllerManagerConfig
        # -- as drift and reconcile forever. The per-field (nested) late_initialize
        # must adopt those backend defaults so the resource settles at Synced=True.
        (ref, cr) = partial_component_config_cluster

        cluster_name = cr["spec"]["name"]
        wait_for_cluster_active(eks_client, cluster_name)

        # Must converge despite the partial spec.
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)
        get_and_assert_status(ref, 'ACTIVE', True)

        cr = k8s.get_resource(ref)
        spec = cr["spec"]

        # The user-provided value is preserved (late-init only fills nil fields).
        assert spec["kubeAPIServerConfig"]["eventTTL"] == "45m"

        # The nested sibling that was omitted is late-initialized from the backend
        # default -- this is the case a top-level late_initialize would miss.
        snpr = spec["kubeAPIServerConfig"].get("serviceNodePortRange")
        assert snpr is not None
        assert snpr.get("minPort") is not None
        assert snpr.get("maxPort") is not None

        # The entirely-omitted sibling configs are late-initialized too.
        assert spec.get("kubeSchedulerConfig") is not None
        assert spec.get("kubeControllerManagerConfig") is not None

        # No perpetual drift: stays Synced across a settle window. If nested
        # late-init were missing, serviceNodePortRange would keep diffing and flip
        # Synced back to False here.
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=3)

    def test_cluster_control_plane_egress_mode(self, eks_client, simple_cluster):
        (ref, cr) = simple_cluster

        cluster_name = cr["spec"]["name"]
        wait_for_cluster_active(eks_client, cluster_name)
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)
        get_and_assert_status(ref, 'ACTIVE', True)

        # cluster_simple omits controlPlaneEgressMode. EKS defaults it to
        # AWS_MANAGED and returns it on read, so late_initialize must adopt that
        # value into the spec. Without it the controller sees the returned
        # default as drift against a nil spec value and never converges.
        cr = k8s.get_resource(ref)
        assert cr["spec"]["resourcesVPCConfig"]["controlPlaneEgressMode"] == "AWS_MANAGED"

        aws_mode = aws_control_plane_egress_mode(eks_client, cluster_name)
        if aws_mode is not None:
            assert aws_mode == "AWS_MANAGED"

        # No perpetual drift: stays Synced across a settle window.
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)

        # Clearing the field leaves desired nil while the observed state still
        # holds a value, which fires the delta with nothing to send. That must
        # not be treated as a revert (the cluster is AWS_MANAGED already) and
        # must not be dereferenced; late-init simply re-adopts the default.
        k8s.patch_custom_resource(
            ref,
            {"spec": {"resourcesVPCConfig": {"controlPlaneEgressMode": None}}},
        )
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=10)
        get_and_assert_status(ref, 'ACTIVE', True)
        terminal = k8s.get_resource_condition(ref, "ACK.Terminal")
        assert terminal is None or str(terminal.get('status')) != str(True)
        cr = k8s.get_resource(ref)
        assert cr["spec"]["resourcesVPCConfig"]["controlPlaneEgressMode"] == "AWS_MANAGED"

        # Regression guard for the reason this field was originally ignored in
        # generator.yaml: EKS allows only one type of update per
        # UpdateClusterConfig call, so an endpoint-access change must not carry
        # controlPlaneEgressMode along. If it does, EKS rejects the call with
        # "Only one type of update can be allowed" and endpoint-access updates
        # break for every cluster that has an egress mode set.
        k8s.patch_custom_resource(
            ref,
            {"spec": {"resourcesVPCConfig": {"endpointPrivateAccess": False}}},
        )
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)
        wait_for_cluster_active(eks_client, cluster_name)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)
        get_and_assert_status(ref, 'ACTIVE', True)

        aws_res = eks_client.describe_cluster(name=cluster_name)
        assert aws_res["cluster"]["resourcesVpcConfig"]["endpointPrivateAccess"] is False
        # The egress mode is untouched by an endpoint-access update.
        aws_mode = aws_control_plane_egress_mode(eks_client, cluster_name)
        if aws_mode is not None:
            assert aws_mode == "AWS_MANAGED"

        # AWS_MANAGED -> CUSTOMER_ROUTED is the supported direction and is
        # reconciled as its own update type.
        endpoint_access_keys = (
            "endpointPrivateAccess", "endpointPublicAccess", "publicAccessCidrs")
        aws_res = eks_client.describe_cluster(name=cluster_name)
        endpoint_access_before = {
            key: aws_res["cluster"]["resourcesVpcConfig"][key]
            for key in endpoint_access_keys
        }

        k8s.patch_custom_resource(
            ref,
            {"spec": {"resourcesVPCConfig": {"controlPlaneEgressMode": "CUSTOMER_ROUTED"}}},
        )
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)
        wait_for_cluster_active(eks_client, cluster_name)

        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)
        get_and_assert_status(ref, 'ACTIVE', True)

        aws_mode = aws_control_plane_egress_mode(eks_client, cluster_name)
        if aws_mode is not None:
            assert aws_mode == "CUSTOMER_ROUTED"

        # The egress-mode update must not disturb the endpoint access settings,
        # whatever the preceding steps left them as.
        aws_res = eks_client.describe_cluster(name=cluster_name)
        endpoint_access_after = {
            key: aws_res["cluster"]["resourcesVpcConfig"][key]
            for key in endpoint_access_keys
        }
        assert endpoint_access_after == endpoint_access_before

        # EKS does not support returning to AWS_MANAGED once a cluster is
        # CUSTOMER_ROUTED. The controller does not encode that rule; it sends the
        # update and surfaces whatever EKS says. The refusal arrives as
        # InvalidParameterException, which is not in the resource's
        # terminal_codes, so it is reported as recoverable and retried. What
        # matters here is that the rejection is visible on the CR and that the
        # cluster is left alone.
        #
        # If InvalidParameterException is ever added to terminal_codes, this
        # becomes an ACK.Terminal assertion instead.
        k8s.patch_custom_resource(
            ref,
            {"spec": {"resourcesVPCConfig": {"controlPlaneEgressMode": "AWS_MANAGED"}}},
        )
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        assert k8s.wait_on_condition(ref, "ACK.Recoverable", "True", wait_periods=10)

        recoverable_condition = "ACK.Recoverable"
        cond = k8s.get_resource_condition(ref, recoverable_condition)
        if cond is None:
            msg = (f"Failed to find {recoverable_condition} condition in "
                f"resource {ref}")
            pytest.fail(msg)
        # The condition carries the service's own message. Match only the stable
        # parts of it; the request ID and prefix vary per call.
        message = str(cond.get('message'))
        assert "ControlPlaneEgressMode" in message
        assert "not supported" in message

        # The rejected revert left the cluster untouched.
        aws_mode = aws_control_plane_egress_mode(eks_client, cluster_name)
        if aws_mode is not None:
            assert aws_mode == "CUSTOMER_ROUTED"

    def test_create_cluster_control_plane_egress_mode(self, eks_client, control_plane_egress_mode_cluster):
        (ref, cr) = control_plane_egress_mode_cluster

        cluster_name = cr["spec"]["name"]
        wait_for_cluster_active(eks_client, cluster_name)
        assert k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=30)
        get_and_assert_status(ref, 'ACTIVE', True)

        # The value has to survive the create path rather than being applied by a
        # follow-up update, so the cluster must come up CUSTOMER_ROUTED.
        cr = k8s.get_resource(ref)
        assert cr["spec"]["resourcesVPCConfig"]["controlPlaneEgressMode"] == "CUSTOMER_ROUTED"

        aws_mode = aws_control_plane_egress_mode(eks_client, cluster_name)
        if aws_mode is not None:
            assert aws_mode == "CUSTOMER_ROUTED"

        # A user-supplied value must not be re-reconciled into an update.
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)
        assert eks_client.list_updates(name=cluster_name)["updateIds"] == []
