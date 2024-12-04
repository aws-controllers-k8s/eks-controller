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
"""Bootstraps the resources required to run the EKS integration tests.
"""
import logging

import boto3
from acktest.aws.identity import get_region
from acktest.bootstrapping import Resources, BootstrapFailureException
from acktest.bootstrapping.iam import Role
from acktest.bootstrapping.vpc import VPC
from acktest.bootstrapping.eks import Cluster
from e2e import bootstrap_directory
from e2e.bootstrap_resources import BootstrapResources

def get_availability_zone_names():
    ec2_client = boto3.client("ec2", region_name=get_region())
    zones = ec2_client.describe_availability_zones()
    return list(map(lambda x: x['ZoneName'], zones['AvailabilityZones']))

def service_bootstrap() -> Resources:
    logging.getLogger().setLevel(logging.INFO)
    
    zones = get_availability_zone_names()
    # We create one subnet more than the number of AZs in order to have the last subnet in the same AZ as the first one
    num_public_subnet=len(zones) + 1

    resources = BootstrapResources(
        ClusterRole=Role("cluster-role", "eks.amazonaws.com", managed_policies=["arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"]),
        FargatePodRole=Role("fargate-pod-role", "eks-fargate-pods.amazonaws.com", managed_policies=["arn:aws:iam::aws:policy/AmazonEKSFargatePodExecutionRolePolicy"]),
        NodegroupRole=Role("nodegroup-role", "ec2.amazonaws.com", managed_policies=[
            "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
            "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
            "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
        ]),
        PodIdentityAssociationRole=Role(
            "ack-pod-identity-association-role",
            "pods.eks.amazonaws.com",
            managed_policies=[
                "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
            ]
        ),
        AccessEntryPrincipalRole=Role(
            "ack-access-entry-principal-role",
            "eks.amazonaws.com",
        ),
        ClusterVPC=VPC(name_prefix="cluster-vpc", num_public_subnet=num_public_subnet, num_private_subnet=2),
        AdoptionCluster=Cluster(name_prefix="adoption-cluster", num_managed_nodes=1)
    )

    try:
        resources.bootstrap()
    except BootstrapFailureException as ex:
        exit(254)

    return resources

if __name__ == "__main__":
    config = service_bootstrap()
    # Write config to current directory by default
    config.serialize(bootstrap_directory)
