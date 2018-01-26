cp -a ./install/tar/*.yaml ./output/$1/tar/
cp -a ./install/tar/install-kube.sh ./output/$1/tar/
cp -a ./LICENSE ./output/$1/tar/
cp -a ./build.txt ./output/$1/tar/
cd ./output/$1/tar/

tar -czvf bdsocp-$1.tar.gz *.tar *.yaml *.sh build.txt LICENSE

cd ../../..

mkdir -p ./release/

\cp ./output/$1/tar/bdsocp-$1.tar.gz ./release/
