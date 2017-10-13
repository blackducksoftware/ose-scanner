FROM scratch

MAINTAINER Black Duck Hub Team

ARG bds_ver
ARG LASTCOMMIT
ARG BUILDTIME
ARG BUILD

COPY ./output/ose_arbiter /ose_arbiter
COPY ./ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY ./LICENSE /LICENSE
COPY ./NOTICE /NOTICE

# Container catalog requirements
COPY ./LICENSE /licenses/
COPY ./help.1 /help.1

LABEL name="Black Duck OpsSight for OpenShift Arbiter" \
      vendor="Black Duck Software" \
      version="${bds_ver}" \
      release="1" \
      summary="Black Duck OpsSight for OpenShift Arbiter" \
      description="Scan arbiter for OpsSight for OpenShift" \
      com.blackducksoftware.version="${bds_ver}" \
      com.blackducksoftware.product="Hub" \
      com.blackducksoftware.platform="OpenShift" \
      com.blackducksoftware.platform.version="${bds_ver}" \
      com.blackducksoftware.platform.component="OpsSight for OpenShift Arbiter Service" \
      com.blackducksoftware.opssight.lastcommit="$LASTCOMMIT" \
      com.blackducksoftware.opssight.buildtime="$BUILDTIME" \
      com.blackducksoftware.opssight.build="$BUILD"

# Clear any existing entry points
ENTRYPOINT ["/ose_arbiter"]
EXPOSE 9035
