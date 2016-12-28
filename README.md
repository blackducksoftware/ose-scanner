# ose-scanner

The ose-scanner provides integration between Black Duck Hub and OpenShift v3. Three components are part of the solution, and each has an independent make file.

# Components

## <a name="controller"></a>Controller

The controller operates as a service. It has three core functional elements and forms the main integration point with OpenShift. The controller executes within the OpenShift master role. At startup, the controller first enumerates all Docker images present in the target OpenShift environment. Those images are queued for scanning by [hub_scanner](#hub_scanner) docker images. The controller currently will process up to five concurrent scans. Once image enumeration completes, the controller monitors for new images and as they are created, these new images are queued for for the [hub_scanner](#hub_scanner)

## <a name="hub_scanner"></a>Hub Scanner

The hub_scanner is a Docker image responsible for extracting the Docker image layers from a Docker image and invoking a Hub scan engine. The version of the Hub scan engine is compiled in during the Docker image creation, but if the Hub server is running a newer version, the scan engine will automatically update to the newer version. Note that in production systems, the hub_scanner Docker image should be updated whenever the Hub server is updated. Doing so will ensure the hub_scanner operates at peak efficiency.

## <a name="notification"></a>Notifications

Hub notifications are sent to the notification service. The notification contains an indication of the policy state for the named Docker image. 

# Compilation Setup

## Setting up the workspace and building

This integration is written primarily in Golang (Go). Go programs are generally compiled within workspaces: https://golang.org/doc/code.html.

Follow these instructions and install Golang on your system. The minimum supported Golang version is 1.6.
* https://golang.org/doc/install

Pick a workspace dir for this project. For example ~/work
Set the GOPATH, then make other directories within it

```
export GOPATH=~/work
cd $GOPATH
mkdir bin
mkdir pkg
mkdir -p src/github.com/blackducksoftware
cd src/blackducksoftware
git clone https://github.com/blackducksoftware/ose-scanner.git
cd ose-scanner
make
```

# Installation

To install the components of this integration, build them and then copy the contents of the ```output``` directory to your OpenShift cluster manager. Then import the containers.

```
docker load < ./hub_ose_scanner.tar
docker load < ./hub_ose_controller.tar

```

# Debugging

If you find your hub_scanner unable to reach the hub server, this most likely means you're lacking an IPv4 route. Run the following to resolve this:

```
sysctl -w net.ipv4.ip_forward=1
```




