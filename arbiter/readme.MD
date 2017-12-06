# arbiter

This is the arbiter service for the Black Duck Hub and OpenShift integration. To make it, perform a top level make.

## Role

The arbiter is responsible for monitoring an OpenShift cluster for container images available to it and controller requests to perform scans. When a given controller identifies an image needing scan, it'll request the scan from the arbiter. If the image has previsouly been scanned on a different node, the arbiter ensures it's only scanned once. The arbiter is also responsible for processing any Hub notifications and ensuring CVE and policy status are consistent between the OpenShift Cluster and Hub. The arbiter operates as a Replica with replica count of 1 with its deployment governed by OpenShift policy. The container is built from scratch and has no user space dependencies. It operates with an elevated security context and security attack surface decisions will be part of the PR process.
