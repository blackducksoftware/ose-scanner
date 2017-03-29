BDS_SCANNER := 3.5.1
PKG_LIST := $(shell go list ./... | grep -v /vendor/)

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output

	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.bds_version=$(BDS_SCANNER)" -o ./output/ose_arbiter ./cmd/arbiter
	
	docker build --rm --build-arg bds_ver=$(BDS_SCANNER) -t hub_ose_arbiter:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_arbiter:$(BDS_SCANNER) > ./output/hub_ose_arbiter.tar

travis:
	mkdir output
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.bds_version=$(BDS_SCANNER)" -o ./output/ose_arbiter ./cmd/arbiter

vet:
	go fmt ${PKG_LIST}
	go vet ${PKG_LIST}


