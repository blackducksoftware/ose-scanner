FROM projectatomic/atomicapp
MAINTAINER Black Duck Hub Team

ARG bds_ver

ENV JAVA_HOME=/opt/jdk/jre \
    APP_HOME=/scan.cli-${bds_ver} \
    BDS_SCANNER_BASE_DIR="/tmp/bds-scanner" \
    SCAN_CLI_JAR_NAME=scan.cli-${bds_ver}-standalone.jar \
    LANG=C.UTF-8

ENV PATH=${APP_HOME}/bin:${JAVA_HOME}/bin:${PATH}

COPY ./output/ose_scanner /ose_scanner

LABEL vendor="Black Duck Software" \
      com.blackducksoftware.version="${bds_ver}" \
      com.blackducksoftware.product="Hub" \
      com.blackducksoftware.platform="OpenShift" \
      com.blackducksoftware.platform.version="${bds_ver}" \
      com.blackducksoftware.platform.component="Hub Scanner" 

# Clear any existing entry points
ENTRYPOINT []
#ENTRYPOINT [ "/ose_scanner" ]
#CMD [ "-?" ]