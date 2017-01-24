## Overview

The ose-scanner provides integration between Black Duck Hub and OpenShift v3. In the current implementation, pre-existing OpenShift images are automatically discovered and any ImageStream activity is monitored. When an image is discovered, the integration kicks off a Black Duck scan engine container to perform the scan and upload the results to your Black Duck Hub instance. Obviously this integration requires both OpenShift and Black Duck Hub!

## Build
[![Go Report Card](https://goreportcard.com/badge/github.com/blackducksoftware/ose-scanner)](https://goreportcard.com/report/github.com/blackducksoftware/ose-scanner)

## Documentation

All documentation, including compilation, installation and debugging instructions can be found on the [wiki](https://github.com/blackducksoftware/ose-scanner/wiki)


## Release Status

This project is under active development and has had no official releases. We welcome all contributions, and anyone attempting to use the code contained in here should expect rough edges and operational issues until release. If you identify an issue, please raise it, or better yet propose a solution.

## License

Apache License 2.0




