BDS_VER ?= 3.6.2
BUILD_NUMBER_FILE=build.txt

all: clean build tar-install release-docker

tar-build: clean build tar-install

clean: 
	rm -Rf ./output/$(BDS_VER); mkdir ./output; mkdir ./output/$(BDS_VER);

build:
	$(eval OS_BUILD_NUMBER=$(shell cat $(BUILD_NUMBER_FILE)))

	cd ./scanner; make BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./controller; make BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./arbiter; make BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)

tar-install:
	mkdir ./output/$(BDS_VER)/tar; cp -a ./scanner/output/*.tar ./output/$(BDS_VER)/tar; cp -a ./controller/output/*.tar ./output/$(BDS_VER)/tar; cp -a ./arbiter/output/*.tar ./output/$(BDS_VER)/tar
	./build-tar-installer.sh $(BDS_VER)

docker-install:
	rm -Rf ./output/$(BDS_VER)/docker; mkdir ./output/$(BDS_VER)/docker
	./build-docker-installer.sh $(BDS_VER)

travis: 
	rm -Rf ./output/$(BDS_VER); mkdir ./output; mkdir ./output/$(BDS_VER);
	cd ./scanner; make travis BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./controller; make travis BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)
	cd ./arbiter; make travis BDS_SCANNER=$(BDS_VER) OCP_BUILD_NUMBER=$(OS_BUILD_NUMBER)

release: build-num all 

build-num:
	@if ! test -f $(BUILD_NUMBER_FILE); then echo 0 > $(BUILD_NUMBER_FILE); fi
	@echo $$(($$(cat $(BUILD_NUMBER_FILE)) + 1)) > $(BUILD_NUMBER_FILE)

build-docker: clean build release-docker

release-docker: docker-install docker-push

docker-push:
	docker login ;\
	docker tag hub_ose_arbiter:$(BDS_VER) blackducksoftware/hub_ose_arbiter:$(BDS_VER) ;\
	docker push blackducksoftware/hub_ose_arbiter:$(BDS_VER) ;\
	docker tag hub_ose_controller:$(BDS_VER) blackducksoftware/hub_ose_controller:$(BDS_VER) ;\
	docker push blackducksoftware/hub_ose_controller:$(BDS_VER) ;\
	docker tag hub_ose_scanner:$(BDS_VER) blackducksoftware/hub_ose_scanner:$(BDS_VER) ;\
	docker push blackducksoftware/hub_ose_scanner:$(BDS_VER) ;\
	docker logout
