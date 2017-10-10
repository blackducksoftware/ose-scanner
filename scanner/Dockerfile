FROM centos:centos7
MAINTAINER Black Duck Hub Team

RUN yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical --setopt=tsflags=nodocs && yum clean all

ARG bds_ver
ARG LASTCOMMIT
ARG BUILDTIME
ARG BUILD

ENV APP_HOME=/scan.cli-${bds_ver} \
    JAVA_HOME=/scan.cli-${bds_ver}/jre \
    BDS_SCANNER_BASE_DIR="/tmp/bds-scanner" \
    SCAN_CLI_JAR_NAME=scan.cli-${bds_ver}-standalone.jar \
    LANG=C.UTF-8

ENV PATH=${APP_HOME}/bin:${JAVA_HOME}/bin:${PATH}

COPY ./output/ose_scanner /ose_scanner
COPY ./hub_scanner/scan.cli/scan.cli-${bds_ver}-zip /scan.cli-${bds_ver}
COPY ./LICENSE /LICENSE

# Container catalog requirements
COPY ./LICENSE /licenses/
COPY ./help.1 /help.1

LABEL name="Black Duck Insight for OpenShift Scanner" \
      vendor="Black Duck Software" \
      version="${bds_ver}" \
      release="1" \
      summary="Black Duck Insight for OpenShift Scanner" \
      description="Scan engine for Insight for OpenShift" \
      com.blackducksoftware.version="${bds_ver}" \
      com.blackducksoftware.product="Hub" \
      com.blackducksoftware.platform="OpenShift" \
      com.blackducksoftware.platform.version="${bds_ver}" \
      com.blackducksoftware.platform.component="Insight for OpenShift Scanner" \ 
      com.blackducksoftware.insight.lastcommit="$LASTCOMMIT" \
      com.blackducksoftware.insight.buildtime="$BUILDTIME" \
      com.blackducksoftware.insight.build="$BUILD"

ENTRYPOINT [ "/ose_scanner" ]
EXPOSE 9036
