cp -a ./*.yaml ./output/
cp -a ./install.sh ./output/
cp -a ./LICENSE ./output/
cp -a ./build.txt ./output/
cd ./output

tar -czvf bdsocp$1.tar.gz *.tar *.yaml *.sh build.txt LICENSE

cd ..

\cp ./output/bdsocp$1.tar.gz ./release/