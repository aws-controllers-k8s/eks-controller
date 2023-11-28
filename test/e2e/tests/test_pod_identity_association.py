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

"""Integration tests for the EKS PodIdentityAssociation resource
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
from e2e.fixtures import k8s_service_account

from .test_cluster import simple_cluster, wait_for_cluster_active

RESOURCE_PLURAL = 'podidentityassociations'

#TODO(a-hilaly): Dynamically create this role...
PIA_ROLE = "arn:aws:iam::632556926448:role/ack-eks-controller-pia-role"

CREATE_WAIT_AFTER_SECONDS = 10

@pytest.fixture
def pod_identity_association(k8s_service_account, eks_client, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    cr_name = random_suffix_name("s3-readonly-pia", 24)
    namespace = "default"
    service_account_name = "s3-readonly-service-account"
    _ = k8s_service_account(namespace, service_account_name)
    
    (ref, cr) = simple_cluster
    cluster_name = cr["spec"]["name"]

    wait_for_cluster_active(eks_client, cluster_name)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CR_NAME"] = cr_name
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["NAMESPACE"] = namespace
    replacements["ROLE_ARN"] = "arn:aws:iam::632556926448:role/ack-eks-controller-pia-role"
    replacements["SERVICE_ACCOUNT"] = service_account_name

    resource_data = load_eks_resource(
        "pod_identity_association",
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
class TestPodIdentityAssociation:
    def test_create_delete_pod_identity_association(self, pod_identity_association, eks_client):
        (ref, cr) = pod_identity_association

        cluster_name = cr["spec"]["clusterName"]
        association_id = cr["status"]["associationID"]
        namespace = cr["spec"]["namespace"]
        role_arn = cr["spec"]["roleARN"]

        try:
            aws_res = eks_client.describe_pod_identity_association(
                clusterName=cluster_name,
                associationId=association_id
            )
            assert aws_res is not None

            assert aws_res["association"]["namespace"] == namespace
            assert aws_res["association"]["role_arn"] == role_arn
            assert aws_res["association"]["associationId"] == association_id
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find PodIdentityAssociation '{ref.name}' in EKS")
