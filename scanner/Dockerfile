FROM projectatomic/atomicapp
MAINTAINER Black Duck Hub Team

RUN yum -y update && yum clean all

ARG bds_ver

ENV APP_HOME=/scan.cli-${bds_ver} \
    JAVA_HOME=/scan.cli-${bds_ver}/jre \
    BDS_SCANNER_BASE_DIR="/tmp/bds-scanner" \
    SCAN_CLI_JAR_NAME=scan.cli-${bds_ver}-standalone.jar \
    LANG=C.UTF-8

ENV PATH=${APP_HOME}/bin:${JAVA_HOME}/bin:${PATH}

COPY ./output/ose_scanner /ose_scanner

COPY ./hub_scanner/scan.cli/scan.cli-${bds_ver} /scan.cli-${bds_ver}

LABEL vendor="Black Duck Software" \
      com.blackducksoftware.version="${bds_ver}" \
      com.blackducksoftware.product="Hub" \
      com.blackducksoftware.platform="OpenShift" \
      com.blackducksoftware.platform.version="${bds_ver}" \
      com.blackducksoftware.platform.component="Hub Scanner" 

ENTRYPOINT [ "/ose_scanner" ]
EXPOSE 9036