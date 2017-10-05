BDS_SCANNER ?= 3.6.2
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output

	#Assumes build on Cent/RHEL7.x
	\cp /etc/ssl/certs/ca-bundle.crt ca-certificates.crt
	\cp ../LICENSE ./

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.bds_version=$(BDS_SCANNER) -X main.build_num=$(OCP_BUILD_NUMBER)" -o ./output/controller ./cmd/controller
	
	docker build --rm --build-arg bds_ver=$(BDS_SCANNER) -t hub_ose_controller:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_controller:$(BDS_SCANNER) > ./output/hub_ose_controller.tar

vet:
	go fmt ${PKG_LIST}
	go vet ${PKG_LIST}

travis:
	mkdir output
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.bds_version=$(BDS_SCANNER) -X main.build_num=$(OCP_BUILD_NUMBER)" -o ./output/controller ./cmd/controller
