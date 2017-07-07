#!/bin/bash

#
# Set the version of Hub we want to connect to
#
#

#set -x

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

    # return success
    return 0
}

clear
echo "============================================"
echo "Black Duck Insight for OpenShift Installation"
echo "============================================"

echo " "
echo "============================================"
echo "Black Duck Hub Configuration Information"
echo "============================================"

#set defaults
DEF_WORKERS="2"
DEF_HUBUSER="sysadmin"
DEF_OSSERVER=`hostname -f`
DEF_OSSERVER="https://$DEF_OSSERVER:8443"
DEF_MASTER=0

read -p "Hub server url (e.g. https://hub.mydomain.com:port): " huburl

allow_insecure="false"

uri_parser "${huburl}" || { echo "Malformed Hub url!"; exit 1; }

if [ "$uri_schema" == "https" -a -z "$uri_port" ];
then
	echo "Do you wish to validate HTTPS certificates?"
	select yn in "Yes" "No"; do
    		case $yn in
        		Yes ) allow_insecure="false"; break;;
        		No ) allow_insecure="true"; break;;
    		esac
	done

	uri_port="443"
elif [ "$uri_schema" == "http" -a -z "$uri_port" ];
then
	uri_port="80"
fi

read -p "Hub user name [$DEF_HUBUSER]: " hubuser
read -sp "Hub password: " hubpassword
echo ""
echo "Please select a supported Hub server version: "
options=("3.6.2" "3.7.1" "Quit")
select supported_version in "${options[@]}"
do
    case $supported_version in
        "3.6.2")
            break
            ;;
        "3.7.1")
            break
            ;;

        "Quit")
            exit
            ;;
        *) echo Invalid Hub version;;
    esac
done

echo " "
read -p "Maximum concurrent scans [$DEF_WORKERS]: " workers

echo "============================================"
echo "OpenShift Configuration"
echo "============================================"
echo " "

# Are we running on a master node or not?
if [ -e /etc/origin/master/master-config.yaml ]; then
    osserver=`grep masterPublicURL /etc/origin/master/master-config.yaml | egrep -o "https://.*[0-9]$" | head -n 1`
    echo "Running on a Master --- Public URL: $osserver"

    isclusteradmin=`oc describe clusterPolicyBindings | sed -n '/Role:[[:space:]]*cluster-admin/,/Groups:/p' | grep "Users:" | grep $(oc whoami) | wc -l`
    if [ $? -ne 0 ]
    then
         echo "Unable to validate user. User must have cluster-admin rights."
         exit 1
    fi

    if [ $isclusteradmin -ne "1" ]
    then
         echo "User does not have required cluster-admin rights."
         exit 1
    fi
	
    DEF_MASTER=1
else
    read -p "OpenShift Cluster [$DEF_OSSERVER]: " osserver
    read -p "Cluster admin user name: " osuser
    read -sp "Cluster admin password: " ospassword

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

version=${supported_version}
#
# Handle secrets
#
secretfile=$(mktemp /tmp/hub_ose_controller.XXXXXX)

cp ./secret.yaml ${secretfile}

sed -i "s/%USER%/${hubuser}/g" ${secretfile}
sed -i "s/%PASSWD%/${hubpassword}/g" ${secretfile}
sed -i "s/%HOST%/${uri_host}/g" ${secretfile}
sed -i "s/%SCHEME%/${uri_schema}/g" ${secretfile}
sed -i "s/%PORT%/${uri_port}/g" ${secretfile}
sed -i "s/%INSECURETLS%/${allow_insecure}/g" ${secretfile}

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
# Create Application
#

podfile=$(mktemp /tmp/hub_ose_controller_pod.XXXXXX)
cp ./insight.yaml ${podfile}

# Note using ~ as separator to avoid URL conflict
sed -i "s~%VERSION%~${version}~g" ${podfile}
sed -i "s~%WORKERS%~${workers}~g" ${podfile}


if [ ! -z "`oc get pod | grep scan-controller`" ];
then
	oc replace -f ${podfile}
else
	oc create -f ${podfile}
fi


echo "Installation complete. Validate application execution from within OpenShift."

