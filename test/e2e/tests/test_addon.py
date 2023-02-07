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

"""Integration tests for the EKS Addon resource
"""

import json
import logging
import time
from typing import Dict, Tuple

import pytest
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import CRD_VERSION, service_marker, CRD_GROUP, load_eks_resource
from e2e.replacement_values import REPLACEMENT_VALUES

from .test_cluster import simple_cluster, wait_for_cluster_active

RESOURCE_PLURAL = 'addons'

CREATE_WAIT_AFTER_SECONDS = 10


def wait_for_addon_deleted(eks_client, cluster_name, addon_name):
    waiter = eks_client.get_waiter('addon_deleted')
    waiter.wait(clusterName=cluster_name, addonName=addon_name)


@pytest.fixture
def coredns_addon(eks_client, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    addon_name = "coredns"
    addon_version = "v1.8.7-eksbuild.3"
    configuration_values = json.dumps(
        {"resources": {"limits": {"memory": "64Mi"}, "requests": {"cpu": "10m", "memory": "64Mi"}}})
    resolve_conflicts = "OVERWRITE"

    (ref, cr) = simple_cluster
    cluster_name = cr["spec"]["name"]

    wait_for_cluster_active(eks_client, cluster_name)

    cr_name = random_suffix_name("addon", 32)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["CR_NAME"] = cr_name
    replacements["ADDON_NAME"] = addon_name
    replacements["ADDON_VERSION"] = addon_version
    replacements["CONFIGURATION_VALUES"] = configuration_values
    replacements["RESOLVE_CONFLICTS"] = resolve_conflicts

    resource_data = load_eks_resource(
        "addon_simple",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create the k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        cr_name, namespace="default",
    )
    k8s.create_custom_resource(ref, resource_data)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)

    yield (ref, cr)

    _, deleted = k8s.delete_custom_resource(ref, 3, 10)
    assert deleted

    wait_for_addon_deleted(eks_client, cluster_name, cr_name)


@service_marker
class TestAddon:
    def test_create_delete_addon(self, coredns_addon, eks_client):
        (ref, cr) = coredns_addon

        cluster_name = cr["spec"]["clusterName"]
        cr_name = ref.name

        addon_name = cr["spec"]["name"]

        try:
            aws_res = eks_client.describe_addon(
                clusterName=cluster_name,
                addonName=addon_name
            )
            assert aws_res is not None

            assert aws_res["addon"]["addonName"] == addon_name
            assert aws_res["addon"]["addonArn"] is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Addon '{cr_name}' in EKS")
