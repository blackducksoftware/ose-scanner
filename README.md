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

The build process assumes you have Docker installed. For Linux, install Docker using your local package manager. If you're using a Mac, then you'll want to follow the [Docker for Mac](https://docs.docker.com/engine/installation/mac/) instructions.

To build, you will need to have a local copy of the Linux scan engine from your Hub server. To obtain the correct version of the scan engine, login to the Hub server. Then click on your name in the upper right corner and select "Tools". Under the Hub Scanner section, you'll find the Linux scanner. Download it and save it to the ```ose-scanner\scanner\hub_scanner``` directory using the file format of ```scan.cli-[dotted-version].zip```. For Hub version 3.4.0, this would become ```scan.cli-3.4.0.zip```.

Pick a workspace dir for this project. For example ~/work
Set the GOPATH, then make other directories within it

```
export GOPATH=~/work
cd $GOPATH
mkdir bin
mkdir pkg
mkdir -p src/github.com/blackducksoftware
cd src/github.com/blackducksoftware
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

# Execution

Note: There is a known permissions issue with the ```hub_ose_controller```. Until that is resolved, please use the ```controller``` syntax below.

## Parameters

* scanner 	Specifies the Docker image ID for the ```hub_ose_scanner```. While it is possible to use the Docker name with tag, specifying the image ID is more reliable.
* h 		Specifies the hostname of the Black Duck Hub instance. 
* p 		Specifies the port for the Black Duck Hub instance.
* s 		Specifies the scheme (http/https) for the Black Duck Hub instance
* u			Specifies the user name for the account in the Black Duck Hub
* w			Specifies the password for the username
* workers	Specifies the number of concurrent Hub scanners. If not specified, the controller will start up to five scanners.

## Controller Syntax

``` ./controller --scanner [id] --h [host] --p 443 --s https --u [user] --w [[password]] --workers 2

# Debugging

If you find your hub_scanner unable to reach the hub server, this most likely means you're lacking an IPv4 route. Run the following to resolve this:

```
sysctl -w net.ipv4.ip_forward=1
```

## Unzip errors
If during a build you receive an error from ```unzip```, this will come from one of two sources.

### No scan engine
To build, you will need to have a local copy of the Linux scan engine from your Hub server. To obtain the correct version of the scan engine, login to the Hub server. Then click on your name in the upper right corner and select "Tools". Under the Hub Scanner section, you'll find the Linux scanner. Download it and save it to the ```ose-scanner\scanner\hub_scanner``` directory using the file format of ```scan.cli-[dotted-version].zip```. For Hub version 3.4.0, this would become ```scan.cli-3.4.0.zip```.

### Unzip error, but scan engine present
If the scan engine is present, and correctly named, an error during unzip indicates a version mismatch. To resolve, verify the version of your scan engine, and edit the ```Makefile``` for both ```scanner``` and ```controller``` to set the ```BDS_VERSION``` to match the version of your scan engine.

## Code location stuck "In Progress" or no files present for a project
If a code location is stuck in progress, you will see no files for the project. One observed scenario is if the OpenShift cluster manager has insufficient free memory and OOM killer runs on a Hub scanner. In such a situation, the initial data for the project may have been uploaded, but the file contents were never transmitted due to the container being killed prior to completion. To verify this situation, look in the controller log for a message of ```signal killed```.





