# ACK service controller for Amazon Elastic Kubernetes Service (EKS)

This repository contains source code for the AWS Controllers for Kubernetes
(ACK) service controller for Amazon EKS.

Please [log issues][ack-issues] and feedback on the main AWS Controllers for
Kubernetes Github project.

[ack-issues]: https://github.com/aws-controllers-k8s/community/issues

## Getting started

To install the `eks-controller` on your cluster, follow the the
[installation][ack-install] instructions.

Currently, the `eks-controller` is GA and supports the following resources:
- `Cluster`
- `Nodegroup`
- `FargateProfile`
- `Addon`
- `PodIdentityAssociation`

A detailed list of the resources supported specifications can be found in the
[references][ack-references] section.

## Annotations

For some resources, the `eks-controller` supports annotations to customize the
behavior of the controller. The following annotations are supported:

- **Nodegroup**
    - `eks.service.k8s.aws/desired-size-managed-by`: used to control whether the
      controller should manage the desiredSize of the Nodegroup `spec.ScalingConfig.DesiredSize`.
      It supports the following values:
        - `ack-eks-controller`: If set the controller will be responsible for
          managing the desired size of the nodegroup.
        - `external-autoscaler`: If set will ignore any changes to the
          `spec.ScalingConfig.DesiredSize` and will not manage the desired size
          of the nodegroup.

      If not set, the controller will default to `ack-eks-controller`.

## Contributing

We welcome community contributions and pull requests.

See our [contribution guide](/CONTRIBUTING.md) for more information on how to
report issues, set up a development environment, and submit code.

We adhere to the [Amazon Open Source Code of Conduct][coc].

You can also learn more about our [Governance](/GOVERNANCE.md) structure.

[coc]: https://aws.github.io/code-of-conduct

## License

This project is [licensed](/LICENSE) under the Apache-2.0 License.

[ack-references]:https://aws-controllers-k8s.github.io/community/reference
[ack-install]:https://aws-controllers-k8s.github.io/community/docs/user-docs/install/