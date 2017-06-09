BDS_SCANNER ?= 3.6.2
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output
	#Assumes build on Cent/RHEL7.x
	\cp /etc/ssl/certs/ca-bundle.crt ca-certificates.crt
	\cp ../LICENSE ./
	
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.bds_version=$(BDS_SCANNER)" -o ./output/ose_arbiter ./cmd/arbiter
	
	docker build --rm --build-arg bds_ver=$(BDS_SCANNER) -t hub_ose_arbiter:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_arbiter:$(BDS_SCANNER) > ./output/hub_ose_arbiter.tar

travis:
	mkdir output
	# Travis isn't executing tests at this point
	#cp /etc/pki/ca-trust/extracted/openssl/ca-bundle.trust.crt ca-certificates.crt
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "-X main.bds_version=$(BDS_SCANNER)" -o ./output/ose_arbiter ./cmd/arbiter

vet:
	go fmt ${PKG_LIST}
	go vet ${PKG_LIST}


