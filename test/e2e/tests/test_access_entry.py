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

"""Integration tests for the EKS AccessEntry resource
"""

import logging
import time
from typing import Dict, Tuple

import pytest
from acktest import tags
from acktest.k8s import resource as k8s
from acktest.resources import random_suffix_name
from e2e import CRD_VERSION, service_marker, CRD_GROUP, load_eks_resource
from e2e.bootstrap_resources import get_bootstrap_resources
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.fixtures import assert_tagging_functionality

from .test_cluster import simple_cluster, wait_for_cluster_active

RESOURCE_PLURAL = 'accessentries'

CREATE_WAIT_AFTER_SECONDS = 10

@pytest.fixture
def access_entry(eks_client, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    cr_name = random_suffix_name("access-entry", 24)
    
    (ref, cr) = simple_cluster
    cluster_name = cr["spec"]["name"]

    wait_for_cluster_active(eks_client, cluster_name)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CR_NAME"] = cr_name
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["PRINCIPAL_ARN"] = get_bootstrap_resources().AccessEntryPrincipalRole.arn
    replacements["ACCESS_POLICY_ARN"] = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"

    resource_data = load_eks_resource(
        "access_entry_simple",
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


@service_marker
class TestAccessEntry:
    def test_create_delete_access_entry(self, access_entry, eks_client):
        (ref, cr) = access_entry

        cluster_name = cr["spec"]["clusterName"]
        principal_arn = cr["spec"]["principalARN"]

        try:
            ae = eks_client.describe_access_entry(
                clusterName=cluster_name,
                principalArn=principal_arn,
            )
            assert ae is not None

            policies = eks_client.list_associated_access_policies(
                clusterName=cluster_name,
                principalArn=principal_arn,
                maxResults=100,
            )
            assert len(policies["associatedAccessPolicies"]) == 1
            assert policies["associatedAccessPolicies"][0]["policyArn"] == cr["spec"]["accessPolicies"][0]["policyARN"]

        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find AccessEntry '{ref.name}' in EKS")

        assert_tagging_functionality(ref, cr["status"]["ackResourceMetadata"]["arn"])
    
    def test_update_access_entry_tags(self, access_entry, eks_client):
        (ref, cr) = access_entry

        cluster_name = cr["spec"]["clusterName"]
        principal_arn = cr["spec"]["principalARN"]

        try:
            ae = eks_client.describe_access_entry(
                clusterName=cluster_name,
                principalArn=principal_arn,
            )
            assert ae is not None
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find AccessEntry '{ref.name}' in EKS")

        # Update the AccessEntry
        new_tags = {
            "key1": "value1",
            "key2": "value2",
        }
        cr["spec"]["tags"] = new_tags

        k8s.patch_custom_resource(ref, cr)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        ae_tags = eks_client.list_tags_for_resource(
            resourceArn=cr["status"]["ackResourceMetadata"]["arn"],
        )["tags"]
        
        tags.assert_ack_system_tags(
            tags=ae_tags,
        )
        tags.assert_equal_without_ack_tags(
            expected=new_tags,
            actual=ae_tags,
        )

    def test_update_access_entry_policies(self, access_entry, eks_client):
        (ref, cr) = access_entry

        cluster_name = cr["spec"]["clusterName"]
        principal_arn = cr["spec"]["principalARN"]

        try:
            ae = eks_client.describe_access_entry(
                clusterName=cluster_name,
                principalArn=principal_arn,
            )
            assert ae is not None

            policies = eks_client.list_associated_access_policies(
                clusterName=cluster_name,
                principalArn=principal_arn,
                maxResults=100,
            )
            assert len(policies["associatedAccessPolicies"]) == 1
            assert policies["associatedAccessPolicies"][0]["policyArn"] == cr["spec"]["accessPolicies"][0]["policyARN"]

        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find AccessEntry '{ref.name}' in EKS")

        # Update the AccessEntry
        cr["spec"]["accessPolicies"] = [
            {
                "policyARN": "arn:aws:eks::aws:cluster-access-policy/AmazonEKSAdminPolicy",
                "accessScope": {
                    "type": "namespace",
                    "namespaces": ["prod-4"],
                }
            },
            {
                "policyARN": "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy",
                "accessScope": {
                    "type": "namespace",
                    "namespaces": ["prod-3"],
                }
            },
            {
                "policyARN": "arn:aws:eks::aws:cluster-access-policy/AmazonEKSEditPolicy",
                "accessScope": {
                    "type": "namespace",
                    "namespaces": ["prod-2"],
                }
            },
        ]

        k8s.patch_custom_resource(ref, cr)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        try:
            ae = eks_client.describe_access_entry(
                clusterName=cluster_name,
                principalArn=principal_arn,
            )
            assert ae is not None

            policies = eks_client.list_associated_access_policies(
                clusterName=cluster_name,
                principalArn=principal_arn,
                maxResults=100,
            )
            assert len(policies["associatedAccessPolicies"]) == 3
            assert policies["associatedAccessPolicies"][0]["policyArn"] == cr["spec"]["accessPolicies"][0]["policyARN"]
            assert policies["associatedAccessPolicies"][1]["policyArn"] == cr["spec"]["accessPolicies"][1]["policyARN"]
            assert policies["associatedAccessPolicies"][2]["policyArn"] == cr["spec"]["accessPolicies"][2]["policyARN"]

        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find AccessEntry '{ref.name}' in EKS")
