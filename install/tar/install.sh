#!/bin/bash

#set -x

options=$@

OSE_KUBERNETES_CONNECTOR=N
WORKER_COUNT=2
INSECURE_TLS=0
UPGRADE=0

function cmdOptions() {

    # An array with all the arguments
    arguments=($options)

    # Loop index
    index=0

    for argument in $options
      do

        index=`expr $index + 1`
        case $argument in

            --workers)
                WORKER_COUNT=${arguments[index]} ;;

            --insecuretls) INSECURE_TLS=1 ;;

            --upgrade) UPGRADE=1 ;;

            --usage) usage; exit 1 ;;

		
        esac
      done

}

function upgrade() {
    echo "Upgrading existing installation at user request."
    oc delete project blackduck-scan

    echo "Wait for delete request to fully complete..."
    sleep 5

    # wait for project resources to be removed
    while true; do
        oc project blackduck-scan

        if [ $? -ne 0 ]
        then
            echo "Delete completed."
            # adm required to ignore quotas
            oc adm new-project blackduck-scan
            break			
        fi
        sleep 3
    done
}

function usage() {

  cat << EOF
  Usage: install.sh <[options]>
  Options:
          --workers        (Optional) The quantity of concurrent scans per node. Default is 2
          --insecuretls    (Optional) If present, relaxes TLS validation
          --upgrade        (Optional) Upgrade existing installation if already installed
          --usage          (Optional) This information

  Environment variables required for non-interactive install:
          BDS_HUB_SERVER   The fully qualified URL for the Black Duck Hub Server
          BDS_HUB_USER     The user name to use for Hub operations
          BDS_HUB_PASSWD   The password for the Hub user.

  Environment variables required when installation occurs remote to the OpenShift Master
          BDS_OCP_SERVER   The name of a master node of the OpenShift cluster
          BDS_OCP_USER     The user name for a cluster admin on the cluster. Value isn't stored.
          BDS_OCP_PASSWD   The password for a cluster admin on the cluster. Value isn't stored.

EOF

}

#
# URI parsing function
# Source: http://wp.vpalos.com/537/uri-parsing-using-bash-built-in-features/
#
# The function creates global variables with the parsed results.
# It returns 0 if parsing was successful or non-zero otherwise.
#
# [schema://][user[:password]@]host[:port][/path][?[arg1=val1]...][#fragment]
#
function uri_parser {
    # uri capture
    uri="$@"

    # safe escaping
    uri="${uri//\`/%60}"
    uri="${uri//\"/%22}"

    # top level parsing
    pattern='^(([a-z]{3,5})://)?((([^:\/]+)(:([^@\/]*))?@)?([^:\/?]+)(:([0-9]+))?)(\/[^?]*)?(\?[^#]*)?(#.*)?$'
    [[ "$uri" =~ $pattern ]] || return 1;

    # component extraction
    uri=${BASH_REMATCH[0]}
    uri_schema=${BASH_REMATCH[2]}
    uri_address=${BASH_REMATCH[3]}
    uri_user=${BASH_REMATCH[5]}
    uri_password=${BASH_REMATCH[7]}
    uri_host=${BASH_REMATCH[8]}
    uri_port=${BASH_REMATCH[10]}
    uri_path=${BASH_REMATCH[11]}
    uri_query=${BASH_REMATCH[12]}
    uri_fragment=${BASH_REMATCH[13]}

    # path parsing
    count=0
    path="$uri_path"
    pattern='^/+([^/]+)'
    while [[ $path =~ $pattern ]]; do
        eval "uri_parts[$count]=\"${BASH_REMATCH[1]}\""
        path="${path:${#BASH_REMATCH[0]}}"
        let count++
    done

    # query parsing
    count=0
    query="$uri_query"
    pattern='^[?&]+([^= ]+)(=([^&]*))?'
    while [[ $query =~ $pattern ]]; do
        eval "uri_args[$count]=\"${BASH_REMATCH[1]}\""
        eval "uri_arg_${BASH_REMATCH[1]}=\"${BASH_REMATCH[3]}\""
        query="${query:${#BASH_REMATCH[0]}}"
        let count++
    done

    if [ "$uri_schema" == "https" -a -z "$uri_port" ];
    then
    	uri_port="443"
    elif [ "$uri_schema" == "http" -a -z "$uri_port" ];
    then
	uri_port="80"
    fi

    # return success
    return 0
}

clear
echo "============================================"
echo "Black Duck OpsSight for OpenShift Installation"
echo "============================================"

# Docker push will fail otherwise
if [ $UID -ne 0 ]; then
  echo -e "\nThis script must be run as root\n"
  exit 1
fi

cmdOptions

INTERACTIVE="false"

if [ "$BDS_HUB_SERVER" == "" ]; then
	INTERACTIVE="true"
else
	echo "Non-interactive installation requested"
fi

#set defaults
DEF_WORKERS="2"
DEF_HUBUSER="sysadmin"
DEF_OSSERVER=`hostname -f`
DEF_OSSERVER="https://$DEF_OSSERVER:8443"
DEF_MASTER=0
allow_insecure="false"

if [ "$INTERACTIVE" == "true" ]; then
	echo " "
	echo "============================================"
	echo "Black Duck Hub Configuration Information"
	echo "============================================"

	read -p "Hub server url (e.g. https://hub.mydomain.com:port): " huburl

	uri_parser "${huburl}" || { echo "Malformed Hub server url!"; exit 1; }

	if [ "$uri_schema" == "https" ]; then
		echo "Do you wish to validate HTTPS certificates?"
		select yn in "Yes" "No"; do
    			case $yn in
        			Yes ) allow_insecure="false"; break;;
	        		No ) allow_insecure="true"; break;;
    			esac
		done
	fi
	read -p "Hub user name [$DEF_HUBUSER]: " hubuser
	read -sp "Hub password: " hubpassword

	echo " "
	read -p "Maximum concurrent scans [$DEF_WORKERS]: " workers
else
	
	huburl=$BDS_HUB_SERVER
	hubuser=$BDS_HUB_USER
	hubpassword=$BDS_HUB_PASSWD
	workers=$WORKER_COUNT

	uri_parser "${huburl}" || { echo "Malformed Hub server url: ${huburl}"; exit 1; }

	if [ "$uri_schema" == "https" ]; then
		if [ $INSECURE_TLS == 1 ]; then
			allow_insecure="true"
		fi
	fi
fi

# Are we running on a master node or not?
if [ -e /etc/origin/master/master-config.yaml ]; then
    osserver=`grep masterPublicURL /etc/origin/master/master-config.yaml | egrep -o "https://.*[0-9]$" | head -n 1`
    echo "Running on a Master --- public URL: $osserver"

    isclusteradmin=`oc describe clusterPolicyBindings | sed -n '/Role:[[:space:]]*cluster-admin/,/Groups:/p' | grep "Users:" | grep $(oc whoami) | wc -l`
    if [ $? -ne 0 ]
    then
	 oc describe clusterrolebindings | grep Name | grep -q cluster-admin ; 
	 if [[ $? -ge 1 ]]; then
	 	echo "Unable to validate user. User must have cluster-admin rights."
                exit 1
	 else
	 	echo "Found cluster role bindings for admin, you must be on 3.7"
	 fi  
    fi
	
    DEF_MASTER=1
else

	if [ "$INTERACTIVE" == "true" ]; then

		echo "============================================"
		echo "OpenShift Configuration"
		echo "============================================"
		echo " "

    		read -p "OpenShift Cluster [$DEF_OSSERVER]: " osserver
    		read -p "Cluster admin user name: " osuser
    		read -sp "Cluster admin password: " ospassword
	else
		osserver=$BDS_OCP_SERVER
		osuser=$BDS_OCP_USER
		ospassword=$BDS_OCP_PASSWD
	fi

    	oc login $osserver -u $osuser -p $ospassword

    	if [ $? -ne 0 ]
    	then
        	exit 1
    	fi
fi

echo " "

#apply defaults
workers="${workers:-$DEF_WORKERS}"
osserver="${osserver:-$DEF_OSSERVER}"
hubuser="${hubuser:-$DEF_HUBUSER}"


oc project blackduck-scan

if [ $? -ne 0 ]
then
	# adm required to ignore quotas
	oc adm new-project blackduck-scan
else
	
	if [ "$INTERACTIVE" == "true" ]; then
		echo "Black Duck OpsSight scanner already installed. Do you wish to upgrade?"
		select yn in "Yes" "No"; do
    			case $yn in
        			Yes ) upgrade; break;;
	        		No ) echo "Aborting upgrade at user request."; exit 2; break;;
    			esac
		done
	else
		echo "Black Duck OpsSight scanner already installed. Processing upgrade..."
		if [ $UPGRADE == 1 ]; then
			upgrade
		else
			echo "Aborting upgrade at user request."
			exit 2
		fi
	fi

fi


# remove default node selector to ensure full scans
oc annotate namespace blackduck-scan openshift.io/node-selector="" --overwrite

#
# Handle service accounts
#
#

oc project blackduck-scan
oc create serviceaccount blackduck-scan

# following allows us to write cluster level metadata for imagestreams
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:blackduck-scan:blackduck-scan

# following allows us to launch priv'd containers for Docker machine access
oc adm policy add-scc-to-user privileged system:serviceaccount:blackduck-scan:blackduck-scan

# start with trying to get remote access securely

if [ $DEF_MASTER -eq 0 ]; then
  dockertoken=`oc whoami -t`
else
  dockertoken=`oc -n blackduck-scan sa get-token blackduck-scan`
fi

oc project default

dockerip=`oc get route -n default | grep docker-registry | tr -s ' ' | cut -d ' ' -f 2`
dockerport=443

echo "Attempting Docker login using secure remote route"
docker login -u blackduck -p ${dockertoken} ${dockerip}:${dockerport}

if [ $? -ne 0 ]
then
	echo "Attempting Docker login using insecure remote route"
	dockerport=80
	docker login -u blackduck -p ${dockertoken} ${dockerip}:${dockerport}
	if [ $? -ne 0 ]
	then
		# Fixed issue if docker registry has Containered Gluster
		dockerip=`oc get svc | egrep "^docker-registry[[:space:]].+$" | tr -s ' ' | cut -d ' ' -f 2`
		dockerport=`oc get svc | egrep "^docker-registry[[:space:]].+$" | tr -s ' ' | cut -d ' ' -f 4 | cut -d '/' -f 1`
		
		docker login -u blackduck -p ${dockertoken} ${dockerip}:${dockerport}
		
		if [ $? -ne 0 ]
		then
			echo "Please validate the docker configuration"
			exit 1
		fi
	fi
fi

oc project blackduck-scan

echo "Loading images into Docker engine. This may take a few minutes..."

docker load < hub_ose_controller.tar
docker load < hub_ose_scanner.tar
docker load < hub_ose_arbiter.tar

version=`docker images | grep "^hub_ose_controller" | sed -n 1p | tr -s ' ' | cut -d ' ' -f 2 `
#controllerimageid=`docker images | grep "^hub_ose_controller" | sed -n 1p | tr -s ' ' | cut -d ' ' -f 3 `

docker tag hub_ose_controller:${version} ${dockerip}:${dockerport}/blackduck-scan/hub_ose_controller:${version}
docker push ${dockerip}:${dockerport}/blackduck-scan/hub_ose_controller:${version}

docker tag hub_ose_scanner:${version} ${dockerip}:${dockerport}/blackduck-scan/hub_ose_scanner:${version}
docker push ${dockerip}:${dockerport}/blackduck-scan/hub_ose_scanner:${version}

docker tag hub_ose_arbiter:${version} ${dockerip}:${dockerport}/blackduck-scan/hub_ose_arbiter:${version}
docker push ${dockerip}:${dockerport}/blackduck-scan/hub_ose_arbiter:${version}

oc project default

# read the docker ip/port again as we need the internal perspective for the services
## Fixed issue if docker registry has Containered Gluster
dockerip=`oc get svc | egrep "^docker-registry[[:space:]].+$" | tr -s ' ' | cut -d ' ' -f 2`
dockerport=`oc get svc | egrep "^docker-registry[[:space:]].+$" | tr -s ' ' | cut -d ' ' -f 4 | cut -d '/' -f 1`

oc project blackduck-scan

#
# Handle secrets
#
secretfile=$(mktemp /tmp/hub_ose_controller.XXXXXX)

cp ./secret.yaml ${secretfile}

sed -i -e "s/%USER%/${hubuser}/g" ${secretfile}
sed -i -e "s/%PASSWD%/${hubpassword}/g" ${secretfile}
sed -i -e "s/%HOST%/${uri_host}/g" ${secretfile}
sed -i -e "s/%SCHEME%/${uri_schema}/g" ${secretfile}
sed -i -e "s/%PORT%/${uri_port}/g" ${secretfile}
sed -i -e "s/%INSECURETLS%/${allow_insecure}/g" ${secretfile}

if [ ! -z "`oc get secrets | grep bds-controller-credentials`" ];
then
	oc delete secret bds-controller-credentials
fi

oc create -f ${secretfile}

rm ${secretfile}

#
# Done secrets
#

#
# Create DS
#

podfile=$(mktemp /tmp/hub_ose_controller_pod.XXXXXX)
cp ./ds.yaml ${podfile}

scanner=${dockerip}:${dockerport}/blackduck-scan/hub_ose_scanner:${version}
controller=${dockerip}:${dockerport}/blackduck-scan/hub_ose_controller:${version}
arbiter=${dockerip}:${dockerport}/blackduck-scan/hub_ose_arbiter:${version}

# Note using ~ as separator to avoid URL conflict
sed -i -e "s~%SCANNER%~${scanner}~g" ${podfile}
sed -i -e "s~%WORKERS%~${workers}~g" ${podfile}
sed -i -e "s~%CONTROLLER%~${controller}~g" ${podfile}
sed -i -e "s~%ARBITER%~${arbiter}~g" ${podfile}
sed -i -e "s~%OSE_KUBERNETES_CONNECTOR%~${OSE_KUBERNETES_CONNECTOR}~g" ${podfile}

if [ ! -z "`oc get pod | grep scan-controller`" ];
then
	oc replace -f ${podfile}
else
	oc create -f ${podfile}
fi

rm ${podfile}

#
# Create RC
#

podfile=$(mktemp /tmp/hub_ose_controller_pod.XXXXXX)
cp ./rc.yaml ${podfile}

# Note using ~ as separator to avoid URL conflict
sed -i -e "s~%SCANNER%~${scanner}~g" ${podfile}
sed -i -e "s~%WORKERS%~${workers}~g" ${podfile}
sed -i -e "s~%CONTROLLER%~${controller}~g" ${podfile}
sed -i -e "s~%ARBITER%~${arbiter}~g" ${podfile}
sed -i -e "s~%OSE_KUBERNETES_CONNECTOR%~${OSE_KUBERNETES_CONNECTOR}~g" ${podfile}

if [ ! -z "`oc get pod | grep scan-arbiter`" ];
then
	oc replace -f ${podfile}
else
	oc create -f ${podfile}
fi

rm ${podfile}

#
# Create Service 
#

podfile=$(mktemp /tmp/hub_ose_controller_pod.XXXXXX)
cp ./svc.yaml ${podfile}

# Note using ~ as separator to avoid URL conflict
sed -i -e "s~%SCANNER%~${scanner}~g" ${podfile}
sed -i -e "s~%WORKERS%~${workers}~g" ${podfile}
sed -i -e "s~%CONTROLLER%~${controller}~g" ${podfile}
sed -i -e "s~%ARBITER%~${arbiter}~g" ${podfile}

if [ ! -z "`oc get svc | grep scan-arbiter`" ];
then
	oc delete svc scan-arbiter
fi
	
oc create -f ${podfile}


rm ${podfile}

echo "Installation complete. Validate pod execution from within OpenShift."

