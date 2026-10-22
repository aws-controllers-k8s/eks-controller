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

"""Integration tests for the EKS Capability resource
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

from .test_cluster import simple_cluster, wait_for_cluster_active, get_and_assert_status

RESOURCE_PLURAL = 'capabilities'

# Time to wait after creating the CR for the status to be populated
CREATE_WAIT_AFTER_SECONDS = 10

# Time to wait after modifying the CR for the status to change
MODIFY_WAIT_AFTER_SECONDS = 10

# Time to wait after the capability has changed status, for the CR to update
CHECK_STATUS_WAIT_SECONDS = 10

# Number of status checks to allow before giving up on the Synced condition.
# The capability is already ACTIVE in EKS by the time the test body runs, so
# this only has to cover the reconcile that observes it.
SYNCED_WAIT_PERIODS = 12

# Budget for the capability to reach ACTIVE in EKS. An ACK capability installs
# a controller and its CRDs into the cluster, which is the slowest of the
# capability types; it usually takes a few minutes.
CAPABILITY_ACTIVE_TIMEOUT_SECONDS = 900
CAPABILITY_ACTIVE_POLL_SECONDS = 15

CAPABILITY_TRANSIENT_STATUSES = ("CREATING", "UPDATING")

def wait_for_capability_active(eks_client, cluster_name, capability_name):
    """Poll DescribeCapability until the capability settles on ACTIVE.

    There is no botocore waiter for capabilities. Waiting here rather than on
    the CR's Synced condition keeps the two concerns apart: this establishes
    that EKS finished its work, and the test then asserts that ACK observed
    it.
    """
    deadline = time.time() + CAPABILITY_ACTIVE_TIMEOUT_SECONDS
    capability = {}
    while time.time() < deadline:
        capability = eks_client.describe_capability(
            clusterName=cluster_name,
            capabilityName=capability_name,
        )["capability"]
        if capability["status"] not in CAPABILITY_TRANSIENT_STATUSES:
            break
        time.sleep(CAPABILITY_ACTIVE_POLL_SECONDS)

    status = capability.get("status")
    # An install that cannot finish, for instance because a CRD it ships
    # conflicts with one already on the cluster, sits in CREATING with the
    # reason in health.issues rather than moving to CREATE_FAILED. Report the
    # issues or the timeout says nothing about why.
    issues = capability.get("health", {}).get("issues", [])
    assert status == "ACTIVE", (
        f"Capability '{capability_name}' did not reach ACTIVE within "
        f"{CAPABILITY_ACTIVE_TIMEOUT_SECONDS}s, last status '{status}', "
        f"health issues: {issues}"
    )

@pytest.fixture
def simple_capability(eks_client, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    (ref, cr) = simple_cluster
    cluster_name = cr["spec"]["name"]

    wait_for_cluster_active(eks_client, cluster_name)

    capability_name = random_suffix_name("capability", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["CAPABILITY_NAME"] = capability_name
    replacements["ROLE_ARN"] = capability_name

    resource_data = load_eks_resource(
        "capability_simple",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        capability_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)

    # The wait sits inside the try so that a capability which never reaches
    # ACTIVE is still deleted. Leaving it behind blocks the cluster fixture's
    # own teardown.
    try:
        wait_for_capability_active(eks_client, cluster_name, capability_name)
        yield (ref, cr)
    finally:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted

@service_marker
class TestCapability:
    def test_create_update_delete_capability(self, simple_capability, eks_client):
        (ref, cr) = simple_capability

        cluster_name = cr["spec"]["clusterName"]
        cr_name = ref.name

        capability_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_capability(
                clusterName=cluster_name,
                capabilityName=capability_name
            )
            assert aws_res is not None

            assert aws_res["capability"]["capabilityName"] == capability_name
            assert aws_res["capability"]["roleArn"] is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Capability '{cr_name}' in EKS")

        assert k8s.wait_on_condition(ref, condition.CONDITION_TYPE_RESOURCE_SYNCED, "True", wait_periods=SYNCED_WAIT_PERIODS, period_length=CHECK_STATUS_WAIT_SECONDS)

        # The ACK configuration declared at create time reaches EKS. Read it
        # back after the capability is synced rather than reusing the response
        # from above, which was captured before the wait.
        aws_res = eks_client.describe_capability(
            clusterName=cluster_name,
            capabilityName=capability_name
        )
        assert aws_res["capability"]["configuration"]["ack"]["enableCrossNamespace"] is True
        assert aws_res["capability"]["configuration"]["ack"]["disabledServices"] == ["s3"]

        cr = k8s.get_resource(ref)
        assert cr["spec"]["configuration"]["ack"]["enableCrossNamespace"] is True
        assert cr["spec"]["configuration"]["ack"]["disabledServices"] == ["s3"]

        # Updating the ACK configuration replaces disabledServices rather than
        # merging it, and carries the cross-namespace flag through
        updates = {
            "spec": {
                "configuration": {
                    "ack": {
                        "enableCrossNamespace": False,
                        "disabledServices": ["ec2"],
                    }
                }
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        # The update puts the capability back into UPDATING, which takes as
        # long as EKS needs to reconfigure the installed controller.
        wait_for_capability_active(eks_client, cluster_name, capability_name)

        assert k8s.wait_on_condition(ref, condition.CONDITION_TYPE_RESOURCE_SYNCED, "True", wait_periods=SYNCED_WAIT_PERIODS, period_length=CHECK_STATUS_WAIT_SECONDS)

        aws_res = eks_client.describe_capability(
            clusterName=cluster_name,
            capabilityName=capability_name
        )
        assert aws_res["capability"]["configuration"]["ack"]["enableCrossNamespace"] is False
        assert aws_res["capability"]["configuration"]["ack"]["disabledServices"] == ["ec2"]

        cr = k8s.get_resource(ref)
        assert cr["spec"]["configuration"]["ack"]["enableCrossNamespace"] is False
        assert cr["spec"]["configuration"]["ack"]["disabledServices"] == ["ec2"]

        assert_tagging_functionality(ref, cr["status"]["ackResourceMetadata"]["arn"])
