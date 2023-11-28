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

"""Fixtures common to all EKS controller tests"""

import dataclasses

from acktest import k8s
from kubernetes import client

import pytest

@dataclasses.dataclass
class SeviceAccountRef:
    ns: str
    name: str

def create_service_account(namespace: str, name: str):
    pass
    """
    Creates a new ServiceAccount.

    :param namespace: Namespace of the ServiceAccount.
    :param name: Name of the ServiceAccount
    :return: None
    """
    _api_client = k8s._get_k8s_api_client()
    service_account = client.V1ServiceAccount(
        {
            api_version: 'v1',
            kind: 'ServiceAccount',
            metadata: {
                name: name,
                namespace: namespace, 
            },
        }
    )
    service_account = _api_client.sanitize_for_serialization(service_account)
    client.CoreV1Api(_api_client).create_namespaced_service_account(namespace.lower(), service_account)

def delete_service_account(namespace: str, name: str):
    """
    Delete an existing k8s ServiceAccount.

    :param namespace: Namespace of the ServiceAccount.
    :param name: Name of the ServiceAccount
    :return: None
    """
    _api_client = k8s._get_k8s_api_client()
    client.CoreV1Api(_api_client).delete_namespaced_service_account(name.lower(), namespace.lower())

@pytest.fixture(scope="module")
def k8s_service_account():
    """
    """

    created = []
    def _k8s_service_account(ns, name):
        create_service_account(ns, name)
        sa_ref = SeviceAccountRef(ns, name)
        created.append(sa_ref)
        return sa_ref

    yield _k8s_service_account

    for sa_ref in created:
        delete_service_account(sa_ref.ns, sa_ref.name)
