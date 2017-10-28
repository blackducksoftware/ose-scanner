# controller

This is the controller service for the Black Duck Hub and OpenShift integration. To make it, perform a top level make.

## Role

The controller is responsible for monitoring an OpenShift node for container images present on it, and then triggering a scan under the guidance of the arbiter. The controller operates as a DaemonSet with its deployment governed by OpenShift policy. The container is built from scratch and has no user space dependencies. It operates with an elevated security context and security attack surface decisions will be part of the PR process.
