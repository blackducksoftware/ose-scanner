BDS_VER ?= 3.6.2
BUILD_NUMBER_FILE=build.txt

all:
	$(eval OS_BUILD_NUMBER=$(shell cat $(BUILD_NUMBER_FILE)))

	rm -Rf ./output; mkdir ./output;
	cd ./scanner; make BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./controller; make BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./arbiter; make BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)

	#copy the results up to our output
	cp -a ./scanner/output/*.tar ./output; cp -a ./controller/output/*.tar ./output; cp -a ./arbiter/output/*.tar ./output
	./build_installer.sh $(BDS_VER)

travis:
	$(eval OS_BUILD_NUMBER=$(shell cat $(BUILD_NUMBER_FILE)))
	rm -Rf ./output; mkdir ./output;
	cd ./scanner; make travis BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./controller; make travis BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./arbiter; make travis BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)

release: build-num all

build-num:
	@if ! test -f $(BUILD_NUMBER_FILE); then echo 0 > $(BUILD_NUMBER_FILE); fi
	@echo $$(($$(cat $(BUILD_NUMBER_FILE)) + 1)) > $(BUILD_NUMBER_FILE)
