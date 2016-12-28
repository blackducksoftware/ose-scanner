# hub_scanner

This is a Docker based Black Duck Hub scanner for use with OpenShift

# Configuration

The Docker image containing the Hub scanner should have a version of the Linux scan engine embedded in it. The directory, ```hub_scanner```, contains the zip file for the scanner. When updating the scanner, ensure the version number in the ```Makefile``` matches that of the ```scan.cli``` zip file.

