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
from e2e.fixtures import assert_tagging_functionality

from .test_cluster import simple_cluster, wait_for_cluster_active
from .test_nodegroup import simple_nodegroup, wait_for_nodegroup_active
from .test_pod_identity_association import PIA_ROLE

RESOURCE_PLURAL = 'addons'

CREATE_WAIT_AFTER_SECONDS = 10

def wait_for_addon_active(eks_client, cluster_name, addon_name):
    waiter = eks_client.get_waiter('addon_active')
    waiter.wait(clusterName=cluster_name, addonName=addon_name)

def wait_for_addon_deleted(eks_client, cluster_name, addon_name):
    waiter = eks_client.get_waiter('addon_deleted')
    waiter.wait(clusterName=cluster_name, addonName=addon_name)

@pytest.fixture
def cluster_with_addons(request, eks_client, simple_nodegroup, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    (ref, cr) = simple_nodegroup
    cluster_name = cr["spec"]["clusterName"]

    wait_for_nodegroup_active(eks_client, cluster_name, cr["spec"]["name"])

    addons = []
    marker = request.node.get_closest_marker("resource_data")
    if marker is not None and 'addons' in marker.args[0]:
        addons = marker.args[0]['addons']

    created = []

    for addon in addons:
        cr_name = random_suffix_name(f"addon-{addon['name']}", 32)

        replacements = REPLACEMENT_VALUES.copy()
        replacements["CLUSTER_NAME"] = cluster_name
        replacements["CR_NAME"] = cr_name
        replacements["ADDON_NAME"] = addon["name"]
        replacements["ADDON_VERSION"] = addon["version"]
        replacements["CONFIGURATION_VALUES"] = addon["configurationValues"] if "configurationValues" in addon else "\{\}"
        replacements["RESOLVE_CONFLICTS"] = addon["resolveConflicts"] if "resolveConflicts" in addon else "NONE"
        replacements["ROLE_ARN"] = addon["roleARN"] if "roleARN" in addon else ""

        file_name = "addon_simple"
        if "roleARN" in addon:
            file_name = "addon_pia"

        resource_data = load_eks_resource(
            file_name,
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

        created += [(ref, cr)]

    yield created

    for ref, _ in created:
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted

        wait_for_addon_deleted(eks_client, cluster_name, ref.name)

@service_marker
class TestAddon:
    @pytest.mark.resource_data({'addons': [
        {
            "name": "coredns",
            "version": "v1.11.1-eksbuild.9",
            "configurationValues": json.dumps({"resources": {"limits": {"memory": "64Mi"}, "requests": {"cpu": "10m", "memory": "64Mi"}}}),
            "resolveConflicts": "OVERWRITE", 
        },
    ]})
    def test_create_delete_addon(self, cluster_with_addons, eks_client, simple_cluster):
        created = cluster_with_addons

        assert len(created) == 1
        (ref, cr) = created[0]

        cluster_name = cr["spec"]["clusterName"]
        wait_for_addon_active(eks_client, cluster_name, "coredns")

        cr_name = ref.name

        addon_name = cr["spec"]["name"]
        configuration_values = cr["spec"]["configurationValues"]

        try:
            aws_res = eks_client.describe_addon(
                clusterName=cluster_name,
                addonName=addon_name
            )
            assert aws_res is not None

            assert aws_res["addon"]["addonName"] == addon_name
            assert aws_res["addon"]["configurationValues"] == configuration_values
            assert aws_res["addon"]["addonArn"] is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Addon '{cr_name}' in EKS")

        assert_tagging_functionality(ref, cr["status"]["ackResourceMetadata"]["arn"])

    @pytest.mark.skip(reason="Super hard setup, needs cert-manager and other things that we don't really want to mess with at this stage.")
    @pytest.mark.resource_data({'addons': [
        {
            "name": "eks-pod-identity-agent",
            "version": "v1.3.0-eksbuild.1",
            "configurationValues": json.dumps({"resources": {"limits": {"memory": "64Mi"}, "requests": {"cpu": "10m", "memory": "64Mi"}}}),
            "resolveConflicts": "NONE",
        },
        {
            "name": "adot",
            "version": "v0.94.1-eksbuild.1",
            "roleARN": PIA_ROLE,
        },
    ]})
    def test_addon_pod_identity_associations(self, eks_client, cluster_with_addons, simple_cluster):
        created = cluster_with_addons
        assert len(created) == 2
        ref, cr = created[1] # adot addon

        cluster_name = cr["spec"]["clusterName"]
        wait_for_addon_active(eks_client, cluster_name, "eks-pod-identity-agent")
        wait_for_addon_active(eks_client, cluster_name, "adot")

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
            assert aws_res["addon"]["status"] == "ACTIVE"
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find Addon '{cr_name}' in EKS")

        wait_for_addon_active(eks_client, cluster_name, "adot")

        aws_res = eks_client.describe_addon(name=cluster_name, addonName="adot")
        assert len(aws_res["addon"]["podIdentityAssociations"]) == 2

        # update pod identity association
        patch = {
            "spec": {
                "podIdentityAssociation": [
                    {
                        "roleARN": PIA_ROLE,
                        "serviceAccount": "adot-col-container-logs",
                    },
                    {
                        "roleARN": PIA_ROLE,
                        "serviceAccount": "adot-col-prom-metrics",
                    },
                    {
                        "roleARN": PIA_ROLE,
                        "serviceAccount": "adot-col-otlp-ingest",
                    }
                ]
            }
        }

        k8s.patch_custom_resource(ref, patch)
        wait_for_addon_active(eks_client, cluster_name, "adot")

        aws_res = eks_client.describe_addon(name=cluster_name, addonName="adot")
        assert len(aws_res["addon"]["podIdentityAssociations"]) == 3
