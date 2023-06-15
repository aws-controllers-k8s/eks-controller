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

import datetime
import time
import typing

import boto3
import pytest

DEFAULT_WAIT_UNTIL_TIMEOUT_SECONDS = 60*10
DEFAULT_WAIT_UNTIL_INTERVAL_SECONDS = 15
DEFAULT_WAIT_UNTIL_DELETED_TIMEOUT_SECONDS = 60*20
DEFAULT_WAIT_UNTIL_DELETED_INTERVAL_SECONDS = 30

ClusterMatchFunc = typing.NewType(
    'ClusterMatchFunc',
    typing.Callable[[dict], bool],
)

class StatusMatcher:
    def __init__(self, status):
        self.match_on = status

    def __call__(self, record: dict) -> bool:
        return (record is not None and 'status' in record
                and record['status'] == self.match_on)


def status_matches(status: str) -> ClusterMatchFunc:
    return StatusMatcher(status)


def wait_until(
        eks_cluster_name: str,
        match_fn: ClusterMatchFunc,
        timeout_seconds: int = DEFAULT_WAIT_UNTIL_TIMEOUT_SECONDS,
        interval_seconds: int = DEFAULT_WAIT_UNTIL_INTERVAL_SECONDS,
    ) -> None:
    """Waits until an EKS cluster with the supplied name is returned from the EKS API
    and the matching functor returns True.

    Usage:
        from e2e.common.waiter import wait_until, status_matches

        wait_until(
            cluster_name,
            status_matches("ACTIVE"),
        )

    Raises:
        pytest.fail upon timeout
    """
    now = datetime.datetime.now()
    timeout = now + datetime.timedelta(seconds=timeout_seconds)

    while not match_fn(get(eks_cluster_name)):
        if datetime.datetime.now() >= timeout:
            pytest.fail("failed to match Cluster before timeout")
        time.sleep(interval_seconds)


def wait_until_deleted(
        eks_cluster_name: str,
        timeout_seconds: int = DEFAULT_WAIT_UNTIL_DELETED_TIMEOUT_SECONDS,
        interval_seconds: int = DEFAULT_WAIT_UNTIL_DELETED_INTERVAL_SECONDS,
    ) -> None:
    """Waits until a DB cluster with a supplied ID is no longer returned from
    the RDS API.

    Usage:
        from e2e.common.waiter import wait_until_deleted

        wait_until_deleted(cluster_name)

    Raises:
        pytest.fail upon timeout or if the EKS cluster goes to any other status
        other than 'DELETING'
    """
    now = datetime.datetime.now()
    timeout = now + datetime.timedelta(seconds=timeout_seconds)

    while True:
        if datetime.datetime.now() >= timeout:
            pytest.fail("Timed out waiting for cluster to be deleted in EKS API")
        time.sleep(interval_seconds)

        latest = get(eks_cluster_name)
        if latest is None:
            break

        if latest['status'] != "DELETING":
            pytest.fail(
                "Status is not 'DELETING' for EKS cluster that was "
                "deleted. Status is " + latest['status']
            )

def get(eks_cluster_name):
    """Returns a dict containing the EKS cluster record from the EKS API.

    If no such cluster exists, returns None.
    """
    c = boto3.client('eks')
    try:
        resp = c.describe_cluster(name=eks_cluster_name)
        assert 'cluster' in resp
        return resp['cluster']
    except c.exceptions.ResourceNotFoundException:
        return None
