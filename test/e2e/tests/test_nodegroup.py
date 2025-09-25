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

"""Integration tests for the EKS Nodegroup resource
"""

import logging
import time
from typing import Dict, Tuple

import pytest

from acktest.k8s import resource as k8s, condition
from acktest.resources import random_suffix_name
from e2e import CRD_VERSION, service_marker, CRD_GROUP, load_eks_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.fixtures import assert_tagging_functionality
from e2e.common import (
    TESTS_DEFAULT_KUBERNETES_VERSION_1_31,
    TESTS_DEFAULT_KUBERNETES_VERSION_1_32,
    TESTS_DEFAULT_KUBERNETES_RELEASE_VERSION_1_31,
    TESTS_DEFAULT_KUBERNETES_RELEASE_VERSION_1_32,
)

from .test_cluster import simple_cluster, wait_for_cluster_active, get_and_assert_status

RESOURCE_PLURAL = 'nodegroups'

# Time to wait after creating the CR for the status to be populated
CREATE_WAIT_AFTER_SECONDS = 10

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 10

# Time to wait after the nodegroup has changed status, for the CR to update
CHECK_STATUS_WAIT_SECONDS = 10

def wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name):
    waiter = eks_client.get_waiter('nodegroup_active')
    waiter.wait(clusterName=cluster_name, nodegroupName=nodegroup_name)

def wait_for_nodegroup_deleted(eks_client, cluster_name, nodegroup_name):
    waiter = eks_client.get_waiter('nodegroup_deleted')
    waiter.wait(clusterName=cluster_name, nodegroupName=nodegroup_name)

@pytest.fixture
def simple_nodegroup(eks_client, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    (ref, cr) = simple_cluster
    cluster_name = cr["spec"]["name"]

    wait_for_cluster_active(eks_client, cluster_name)

    nodegroup_name = random_suffix_name("nodegroup", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["NODEGROUP_NAME"] = nodegroup_name
    replacements["K8S_VERSION"] = TESTS_DEFAULT_KUBERNETES_VERSION_1_31
    replacements["RELEASE_VERSION"] = TESTS_DEFAULT_KUBERNETES_RELEASE_VERSION_1_31

    resource_data = load_eks_resource(
        "nodegroup_simple",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        nodegroup_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)

    yield (ref, cr)

    _, deleted = k8s.delete_custom_resource(ref, 3, 10)
    assert deleted
    wait_for_nodegroup_deleted(eks_client, cluster_name, nodegroup_name)


@service_marker
class TestNodegroup:
    def test_create_update_delete_nodegroup(self, simple_nodegroup, eks_client):
        (ref, cr) = simple_nodegroup

        cluster_name = cr["spec"]["clusterName"]
        cr_name = ref.name

        nodegroup_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_nodegroup(
                clusterName=cluster_name,
                nodegroupName=nodegroup_name
            )
            assert aws_res is not None

            assert aws_res["nodegroup"]["nodegroupName"] == nodegroup_name
            assert aws_res["nodegroup"]["nodegroupArn"] is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Nodegroup '{cr_name}' in EKS")

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        # Update the update and scaling configs fields
        updates = {
            "spec": {
                "updateConfig": {
                    "maxUnavailable": None,
                    "maxUnavailablePercentage": 15
                },
                "scalingConfig": {
                    "minSize": 2,
                    "maxSize": 2,
                    "desiredSize": 2,
                }
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Wait for the updating to become active again
        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        # Ensure status is updated properly once it has become active
        k8s.wait_on_condition(ref, condition.CONDITION_TYPE_READY, "True", wait_periods=5, period_length=CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )

        assert aws_res["nodegroup"]["updateConfig"]["maxUnavailablePercentage"] == 15
        assert aws_res["nodegroup"]["scalingConfig"]["minSize"] == 2
        assert aws_res["nodegroup"]["scalingConfig"]["maxSize"] == 2
        assert aws_res["nodegroup"]["scalingConfig"]["desiredSize"] == 2

        updates = {
            "spec": {
                "labels": {
                    "toot": "shoot",
                    "boot": "snoot"
                },
                "taints": [
                    {
                        "key": "ifbooted",
                        "value": "noexecuted",
                        "effect": "NO_EXECUTE"
                    },
                    {
                        "key": "owner",
                        "value": "teamone",
                        "effect": "NO_SCHEDULE"
                    },
                ],
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )

        assert len(aws_res["nodegroup"]["labels"]) == 2
        assert aws_res["nodegroup"]["labels"]["toot"] == "shoot"
        assert aws_res["nodegroup"]["labels"]["boot"] == "snoot"

        assert len(aws_res["nodegroup"]["taints"]) == 2
        assert aws_res["nodegroup"]["taints"][0]["key"] in ["ifbooted", "owner"]
        assert aws_res["nodegroup"]["taints"][0]["value"] in ["noexecuted", "teamone"]
        assert aws_res["nodegroup"]["taints"][0]["effect"] in ["NO_EXECUTE", "NO_SCHEDULE"]
        assert aws_res["nodegroup"]["taints"][1]["key"] in ["ifbooted", "owner"]
        assert aws_res["nodegroup"]["taints"][1]["value"] in ["noexecuted", "teamone"]
        assert aws_res["nodegroup"]["taints"][1]["effect"] in ["NO_EXECUTE", "NO_SCHEDULE"]

        # Remove a label, update a label and remove a taint
        updates = {
            "spec": {
                "labels": {
                    "toot": "updooted",
                    "boot": None
                },
                "taints": [
                    {
                        "key": "ifbooted",
                        "value": "noexecuted",
                        "effect": "NO_EXECUTE"
                    }
                ],
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )

        assert len(aws_res["nodegroup"]["labels"]) == 1
        assert aws_res["nodegroup"]["labels"]["toot"] == "updooted"
        assert "boot" not in aws_res["nodegroup"]["labels"]

        assert len(aws_res["nodegroup"]["taints"]) == 1
        assert aws_res["nodegroup"]["taints"][0]["key"] == "ifbooted"
        assert aws_res["nodegroup"]["taints"][0]["value"] == "noexecuted"
        assert aws_res["nodegroup"]["taints"][0]["effect"] == "NO_EXECUTE"

        # Remove last taint
        updates = {
            "spec": {
                "taints": [],
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )

        assert "taints" not in aws_res["nodegroup"]

        assert_tagging_functionality(ref, cr["status"]["ackResourceMetadata"]["arn"])

    def test_nodegroup_custom_annotations(self, simple_nodegroup, eks_client):
        (ref, cr) = simple_nodegroup

        cluster_name = cr["spec"]["clusterName"]
        cr_name = ref.name

        nodegroup_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_nodegroup(
                clusterName=cluster_name,
                nodegroupName=nodegroup_name
            )
            assert aws_res is not None

            assert aws_res["nodegroup"]["nodegroupName"] == nodegroup_name
            assert aws_res["nodegroup"]["nodegroupArn"] is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Nodegroup '{cr_name}' in EKS")

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        # Set the eks.services.k8s.aws/desired-size-managed-by annotations and
        # update the desiredSize along with the maxSize
        updates = {
            "metadata": {
                "annotations": {
                    "eks.services.k8s.aws/desired-size-managed-by": "external-autoscaler",
                }
            },
            "spec": {
                "scalingConfig": {
                    "minSize": 1,
                    "maxSize": 2,
                    "desiredSize": 3, # Controller should not apply this value
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )
        assert aws_res["nodegroup"]["scalingConfig"]["minSize"] == 1
        assert aws_res["nodegroup"]["scalingConfig"]["maxSize"] == 2
        assert aws_res["nodegroup"]["scalingConfig"]["desiredSize"] == 1 # Should be 1, not 3

        # Set the eks.services.k8s.aws/desired-size-managed-by annotations and
        # update the desiredSize only
        updates = {
            "metadata": {
                "annotations": {
                    "eks.services.k8s.aws/desired-size-managed-by": "external-autoscaler",
                }
            },
            "spec": {
                "scalingConfig": {
                    "minSize": 1,
                    "maxSize": 2,
                    "desiredSize": 10, # Controller should not apply this value
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )
        assert aws_res["nodegroup"]["scalingConfig"]["minSize"] == 1
        assert aws_res["nodegroup"]["scalingConfig"]["maxSize"] == 2
        assert aws_res["nodegroup"]["scalingConfig"]["desiredSize"] == 1 # Should be 1, not 10

    def test_update_nodegroup_version(self, simple_nodegroup, eks_client):
        (ref, cr) = simple_nodegroup

        cluster_name = cr["spec"]["clusterName"]
        cr_name = ref.name

        nodegroup_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_nodegroup(
                clusterName=cluster_name,
                nodegroupName=nodegroup_name
            )
            assert aws_res is not None

            assert aws_res["nodegroup"]["nodegroupName"] == nodegroup_name
            assert aws_res["nodegroup"]["nodegroupArn"] is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Nodegroup '{cr_name}' in EKS")

        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        # Update nodegroup version fields
        updates = {
            "spec": {
                "version": TESTS_DEFAULT_KUBERNETES_VERSION_1_32,
                "releaseVersion": TESTS_DEFAULT_KUBERNETES_RELEASE_VERSION_1_32,
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # Wait for the nodegroup to become active again
        wait_for_nodegroup_active(eks_client, cluster_name, nodegroup_name)

        # Ensure status is updated properly once it has become active
        k8s.wait_on_condition(ref, condition.CONDITION_TYPE_READY, "True", wait_periods=5, period_length=CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)

        aws_res = eks_client.describe_nodegroup(
            clusterName=cluster_name,
            nodegroupName=nodegroup_name
        )

        assert aws_res["nodegroup"]["version"] == TESTS_DEFAULT_KUBERNETES_VERSION_1_32
        assert aws_res["nodegroup"]["releaseVersion"] == TESTS_DEFAULT_KUBERNETES_RELEASE_VERSION_1_32

        # Assert terminal condition is set to True
        updates = {
            "spec": {
                "version": TESTS_DEFAULT_KUBERNETES_VERSION_1_31,
                "releaseVersion": TESTS_DEFAULT_KUBERNETES_RELEASE_VERSION_1_32,
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        terminal_condition = "ACK.Terminal"
        cond = k8s.get_resource_condition(ref, terminal_condition)
        if cond is None:
            msg = (f"Failed to find {terminal_condition} condition in "
                f"resource {ref}")
            pytest.fail(msg)

        cond_status = cond.get('status', None)
        if str(cond_status) != str(True):
            msg = (f"Expected {terminal_condition} condition to "
                f"have status {terminal_condition} but found {cond_status}")
            pytest.fail(msg)