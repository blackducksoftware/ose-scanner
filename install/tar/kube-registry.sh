#!/bin/bash
set -x

sudo docker load < hub_ose_controller.tar
sudo docker load < hub_ose_scanner.tar
sudo docker load < hub_ose_arbiter.tar

if [[ -z $DEFAULT_REPOSITORY ]] ; then 
	echo "ERROR: need a default repository argument to proceed"
	echo "Set DEFAULT_REPOSITORY which images will be pushed to, and retry"
	exit 1
fi

for image_name in hub_ose_scanner hub_ose_controller hub_ose_arbiter ; do
	img=$( sudo docker images --format='{{.ID}} {{.Repository}}' | grep ${image_name} | grep -v $DEFAULT_REPOSITORY | cut -d' ' -f 1 )
	sudo docker tag $img ${DEFAULT_REPOSITORY}/${image_name}:4.4.0
	sudo docker push $DEFAULT_REPOSITORY/${image_name}:4.4.0
done
