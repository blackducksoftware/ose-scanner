## Overview

The ose-scanner provides integration between Black Duck Hub and OpenShift v3. In the current implementation, pre-existing OpenShift images are automatically discovered and any ImageStream activity is monitored. When an image is discovered, the integration kicks off a Black Duck scan engine container to perform the scan and upload the results to your Black Duck Hub instance. Obviously this integration requires both OpenShift and Black Duck Hub!

## Build
[![Build Status](https://travis-ci.org/blackducksoftware/ose-scanner.svg?branch=master)](https://travis-ci.org/blackducksoftware/ose-scanner)
[![Go Report Card](https://goreportcard.com/badge/github.com/blackducksoftware/ose-scanner)](https://goreportcard.com/report/github.com/blackducksoftware/ose-scanner)

## Documentation

All documentation, including compilation, installation and debugging instructions can be found on the [wiki](https://github.com/blackducksoftware/ose-scanner/wiki)


## Release Status

This project is under active development and has had one official releases. We welcome all contributions, and anyone attempting to use the code contained in here should expect some rough edges and operational issues until release. If you identify an issue, please raise it, or better yet propose a solution.

### Contributing

- Note that contributions are best done via pull request. 
- Its recommended to start small, and propose changes first. Raising an issue and holding discussion will help reduce the iterations during PR review.

### Community Values

Part of any successful open source project is the ethos that the health of the community is more important then the code itself.  In that spirit, we adopt the values of Cris Nova's kubicorn project, which is both extremely successful and also an excellent example of how to build a community (https://github.com/kris-nova/kubicorn).

```
- Infrastructure as software.

We believe that the oh-so important layer of infrastructure should be represented as software (not as code!). We hope that our project demonstrates this idea, so the community can begin thinking in the way of the new paradigm.

- Rainbows and Unicorns

We believe that sharing is important, and encouraging our peers is even more important. Part of contributing to (ose_scanner) means respecting, encouraging, and welcoming others to the project.
```

Thanks again kris !
## License

Apache License 2.0 
