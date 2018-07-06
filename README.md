
# We have moved! 

The original 1.x ose-scanner has now been replaced with the OpsSight 2x product (based on the Black Duck "Perceptor" upstream project for distributed cloud-native security scanning).

Thanks to everyone, especially Tim Mackey, for all the hard work on this great product!

# The OpsSight product

The downstream implementation of this project, hosted by Synopsys (Black Duck) is here: https://github.com/blackducksoftware/opssight-connector/. If you have questions feel free to use the contact list there, or just file an upstream issue.  Also feel free to contact jbeakley@synopsys.com, ngoldman@synopsys.com for a demo!

# The Perceptor open-source project (the upstream source of OpsSight)

OpsSight now has a 'pure' upstream implementation and is thus broken up into several microservices that are composable, and developed in a pure, non-proprietary manner (and also useable out of the box), and exist in their own independent development lifecycles.  These can be adopted and recomposed in any manner, and are all ASF licensed.  Feel free to get involved and create an issue on any of them if you have any questions.  You can contact our development team directly as well.

https://github.com/blackducksoftware/perceptor
https://github.com/blackducksoftware/perceivers
https://github.com/blackducksoftware/perceptor-protoform
https://github.com/blackducksoftware/perceptor-scanner


----------------------------------------------------------------------------------------------------------------------

# Legacy content below ~ only for archival purposes.

## Overview

The **Black Duck OpsSight Connector** (ose-scanner) automates the discovery of open-source components and security vulnerabilities in container images as they are instantiated in container-orchestration platforms. OpsSight helps you prevent known open source vulnerabilities from being deployed into production environments.

With OpsSight, you can:

* Scan and inventory open source in images as they are instantiated in a container-orchestration platform.
* Identify and highlight any images that contain known security vulnerabilities.
* Flag containers that violate open source security policies to prevent them from being deployed to production.
* Receive automated alerts when any newly discovered vulnerabilities may affect containers in your cluster.

The OpsSight Connector detects when new pods are added to a cluster environment (OpsSight on OpenShift also detects new images through ImageStreams), scans those containers, sends information back to the Black Duck Hub, then annotates and labels containers to indicate risks detected in the containers' open-source components. Detailed scan results are available in your Hub instance.  Container annotations can be used to enforce security policies and ensure vulnerable containers are not deployed in production environments.

The end-to-end OpsSight solution therefore requires a Black Duck Hub with an OpsSight feature license.

The latest version of the OpsSight Connector supports both [Kubernetes](https://kubernetes.io/) and [Red Hat OpenShift](https://www.openshift.com/).

For more information on the OpsSight Connector, see the:

* [ose-scanner wiki](https://github.com/blackducksoftware/ose-scanner/wiki) for help with building, running and debugging
* OpsSight Connector [Installation Guide](https://blackducksoftware.atlassian.net/wiki/spaces/BDLM/pages/107741300/OpsSight+Installation+Guide) (Includes Release Notes and Supported Platforms)
* OpsSight Connector [Security Disclosures](https://blackducksoftware.atlassian.net/wiki/spaces/BDLM/pages/107741331/OpsSight+Security+Disclosures)

## Build

See below for Build Status and the Go Report Card for the ose-scanner project.

[![Build Status](https://travis-ci.org/blackducksoftware/ose-scanner.svg?branch=master)](https://travis-ci.org/blackducksoftware/ose-scanner)
[![Go Report Card](https://goreportcard.com/badge/github.com/blackducksoftware/ose-scanner)](https://goreportcard.com/report/github.com/blackducksoftware/ose-scanner)

## Release Status

This project is under active development and has had several official releases, with more in the pipeline.

## Contributing

We welcome all contributions. If you identify an issue, please raise it, and feel free to propose a solution.

Contributions are best done via pull request. It is recommended to start small, and propose changes first. Raising an issue and holding a discussion will help reduce the iterations during pull-request review.

When contributing to this project, please ensure that all changes have been verified using ``go vet`` and formatted per ``go fmt``. This can be done using ``make vet`` within the project.

Please note that PRs that change code (as opposed to readme/docs) and that lack information related to the orchestration-cluster and Hub versions tested against will take longer to approve.

## Community Values

Part of any successful open source project is the ethos that the health of the community is more important than the code itself.

In that spirit, we adopt the values of Cris Nova's [kubicorn project](https://github.com/kris-nova/kubicorn), which is both extremely successful and also an excellent example of how to build a community around an open-source infrastructure software.

We restate the kubicorn core values below and adopt them as our own:

* *Infrastructure as software*: We believe that the important layer of infrastructure should be represented as software (not as code!).
* *Rainbows and Unicorns*: We believe that sharing is important, and encouraging our peers is even more important. Part of contributing to (ose_scanner) means respecting, encouraging, and welcoming others to the project.

## License

[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)
