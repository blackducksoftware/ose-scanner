BDS_VER ?= 3.6.2

all:
	rm -Rf ./output; mkdir ./output;
	cd ./scanner; make BDS_SCANNER=$(BDS_VER)
	cd ./controller; make BDS_SCANNER=$(BDS_VER)
	cd ./arbiter; make BDS_SCANNER=$(BDS_VER)

	#copy the results up to our output
	cp -a ./scanner/output/*.tar ./output; cp -a ./controller/output/*.tar ./output; cp -a ./arbiter/output/*.tar ./output
	./build_installer.sh $(BDS_VER)

travis:
	rm -Rf ./output; mkdir ./output;
	cd ./scanner; make travis BDS_SCANNER=$(BDS_VER)
	cd ./controller; make travis BDS_SCANNER=$(BDS_VER)
	cd ./arbiter; make travis BDS_SCANNER=$(BDS_VER)
