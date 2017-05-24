cp -a ./*.yaml ./output/
cp -a ./install.sh ./output/
cd ./output

tar -czvf bdsocp$1.tar.gz *.tar *.yaml *.sh

cd ..

\cp ./output/bdsocp$1.tar.gz ./release/