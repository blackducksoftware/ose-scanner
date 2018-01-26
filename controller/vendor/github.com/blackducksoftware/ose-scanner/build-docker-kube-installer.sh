#!/bin/bash

cp -a ./install/docker/*.yaml ./output/$1/docker/
cp -a ./install/docker/install-kube.sh ./output/$1/docker/
cp -a ./LICENSE ./output/$1/docker/
cp -a ./build.txt ./output/$1/docker/
cd ./output/$1/docker/

tar -czvf bdsocp$1-docker.tar.gz *.yaml *.sh build.txt LICENSE

cd ../../..

\cp ./output/$1/docker/bdsocp$1-docker.tar.gz ./release/
