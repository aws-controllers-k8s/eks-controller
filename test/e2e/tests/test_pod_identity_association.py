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
from e2e.fixtures import k8s_service_account, assert_tagging_functionality

from .test_cluster import simple_cluster, wait_for_cluster_active

RESOURCE_PLURAL = 'podidentityassociations'

#TODO(a-hilaly): Dynamically create this role...
PIA_ROLE = "arn:aws:iam::632556926448:role/ack-eks-controller-pia-role"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10

POLICY_S3_READ = '''{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:ListBucket"
            ],
            "Resource": "*"
        }
    ]}
'''

POLICY_S3_READWRITE = '''{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:GetObject",
                "s3:PutObject",
                "s3:ListBucket"
            ],
            "Resource": "*"
        }
    ]}
'''

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


@pytest.fixture
def pod_identity_association_with_policy(k8s_service_account, eks_client, simple_cluster) -> Tuple[k8s.CustomResourceReference, Dict]:
    cr_name = random_suffix_name("pia-with-policy", 24)
    namespace = "default"
    service_account_name = random_suffix_name("pia-policy-service-account", 32)
    _ = k8s_service_account(namespace, service_account_name)

    (ref, cr) = simple_cluster
    cluster_name = cr["spec"]["name"]

    wait_for_cluster_active(eks_client, cluster_name)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["CR_NAME"] = cr_name
    replacements["CLUSTER_NAME"] = cluster_name
    replacements["NAMESPACE"] = namespace
    replacements["ROLE_ARN"] = PIA_ROLE
    replacements["SERVICE_ACCOUNT"] = service_account_name
    replacements["POLICY"] = POLICY_S3_READ

    resource_data = load_eks_resource(
        "pod_identity_association_with_policy",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

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
            assert aws_res["association"]["roleArn"] == role_arn
            assert aws_res["association"]["associationId"] == association_id
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find PodIdentityAssociation '{ref.name}' in EKS")

        assert_tagging_functionality(ref, cr["status"]["ackResourceMetadata"]["arn"])

    def test_create_pod_identity_association_with_policy(self, pod_identity_association_with_policy, eks_client):
        (ref, cr) = pod_identity_association_with_policy

        cluster_name = cr["spec"]["clusterName"]
        association_id = cr["status"]["associationID"]

        try:
            aws_res = eks_client.describe_pod_identity_association(
                clusterName=cluster_name,
                associationId=association_id
            )
            assert aws_res is not None

            association = aws_res["association"]
            assert association["namespace"] == cr["spec"]["namespace"]
            assert association["roleArn"] == cr["spec"]["roleARN"]
            assert association["associationId"] == association_id

            # Verify the policy was set on the AWS resource
            assert "policy" in association
            aws_policy = json.loads(association["policy"])
            expected_policy = json.loads(POLICY_S3_READ)
            assert aws_policy == expected_policy
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find PodIdentityAssociation '{ref.name}' in EKS")

        # Update the policy to a broader one
        k8s.patch_custom_resource(ref, {
            "spec": {
                "policy": POLICY_S3_READWRITE
            }
        })
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Verify the policy was updated in AWS
        try:
            aws_res = eks_client.describe_pod_identity_association(
                clusterName=cluster_name,
                associationId=association_id
            )
            assert aws_res is not None

            association = aws_res["association"]
            assert "policy" in association
            aws_policy = json.loads(association["policy"])
            expected_policy = json.loads(POLICY_S3_READWRITE)
            assert aws_policy == expected_policy
        except eks_client.exceptions.ResourceNotFoundException:
            pytest.fail(f"Could not find PodIdentityAssociation '{ref.name}' in EKS")
