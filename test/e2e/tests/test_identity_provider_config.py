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

"""Integration tests for the EKS IdentityProviderConfig resource
"""

import json
import boto3
import logging
import time
from typing import Dict, Tuple

import pytest
from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e import CRD_VERSION, service_marker, CRD_GROUP, load_eks_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.fixtures import k8s_service_account, assert_tagging_functionality

from .test_cluster import simple_cluster, wait_for_cluster_active

RESOURCE_PLURAL = 'identityproviderconfigs'

CREATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10
CHECK_STATUS_WAIT_SECONDS = 40

def wait_for_cluster_active(eks_client, cluster_name):
    waiter = eks_client.get_waiter('cluster_active')
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

@service_marker
@pytest.mark.canary
class TestPodIdentityAssociation:
    def test_create_delete_identity_provider_config(self, simple_cluster, eks_client):
        (ref, cr) = simple_cluster
        cluster_name = cr["spec"]["name"]
        wait_for_cluster_active(eks_client, cluster_name)

        cr_name = random_suffix_name("id-provider-cfg", 32)

        identity_provider_name = "idp"
        replacements = REPLACEMENT_VALUES.copy()
        replacements["CR_NAME"] = cr_name
        replacements["CLUSTER_NAME"] = cluster_name
        replacements["IDP_NAME"] = identity_provider_name

        resource_data = load_eks_resource(
            "identity_provider_config",
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

        try:
            aws_res = eks_client.describe_identity_provider_config(
                clusterName=cluster_name,
                identityProviderConfig={
                    'type': 'oidc',
                    'name': identity_provider_name,
                }
            )
            assert aws_res is not None

            assert aws_res["identityProviderConfig"]["oidc"]["status"] == "CREATING"
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find IdentityProviderConfig '{ref.name}' in EKS")

        get_and_assert_status(ref, 'CREATING', False)

        wait_for_cluster_active(eks_client, cluster_name)

        # Ensure status is updated properly once it has become active
        time.sleep(CHECK_STATUS_WAIT_SECONDS)
        get_and_assert_status(ref, 'ACTIVE', True)

        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted

        time.sleep(DELETE_WAIT_AFTER_SECONDS)
        wait_for_cluster_active(eks_client, cluster_name)

        try:
            aws_res = eks_client.describe_identity_provider_config(
                clusterName=cluster_name,
                identityProviderConfig={
                    'type': 'oidc',
                    'name': identity_provider_name,
                }
            )
            pytest.fail(f"IdentityProviderConfig '{ref.name}' in EKS is not deleted")
        except eks_client.exceptions.ResourceNotFoundException:
            pass

