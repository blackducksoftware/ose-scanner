FROM projectatomic/atomicapp
MAINTAINER Black Duck Hub Team

RUN yum -y update && yum clean all

ARG bds_ver

COPY ./output/controller /ose_controller

LABEL vendor="Black Duck Software" \
      com.blackducksoftware.version="${bds_ver}" \
      com.blackducksoftware.product="Hub" \
      com.blackducksoftware.platform="OpenShift" \
      com.blackducksoftware.platform.version="${bds_ver}" \
      com.blackducksoftware.platform.component="Controller Service" 

# Clear any existing entry points
ENTRYPOINT ["/ose_controller"]