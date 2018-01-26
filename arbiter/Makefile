BDS_SCANNER ?= 4.4.0
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
BUILD_TIME := $(shell cat ../buildtime.txt)
LAST_COMMIT := $(shell git rev-parse HEAD)

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output
	#Assumes build on Cent/RHEL7.x
	\cp /etc/ssl/certs/ca-bundle.crt ca-certificates.crt
	\cp ../LICENSE ./

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.bds_version=$(BDS_SCANNER) -X main.build_num=$(OCP_BUILD_NUMBER)" -o ./output/ose_arbiter ./cmd/arbiter

	docker build --rm --build-arg bds_ver=$(BDS_SCANNER) --build-arg "LASTCOMMIT=$(LAST_COMMIT)" --build-arg "BUILDTIME=$(BUILD_TIME)" --build-arg "BUILD=$(BUILD_TAG)" -t hub_ose_arbiter:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_arbiter:$(BDS_SCANNER) > ./output/hub_ose_arbiter.tar

travis:
	mkdir output
	# Travis isn't executing tests at this point
	#cp /etc/pki/ca-trust/extracted/openssl/ca-bundle.trust.crt ca-certificates.crt
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.bds_version=$(BDS_SCANNER) -X main.build_num=$(OCP_BUILD_NUMBER)" -o ./output/ose_arbiter ./cmd/arbiter

vet:
	go fmt ${PKG_LIST}
	go vet ${PKG_LIST}
