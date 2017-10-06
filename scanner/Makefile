BDS_SCANNER ?= 3.6.2
BUILD_TIME := $(shell cat ../buildtime.txt)
LAST_COMMIT := $(shell git rev-parse HEAD)

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output

	rm -Rf ./build \
	mkdir build

	GOOS=linux GOARCH=amd64 go build -o ./output/ose_scanner .

	rm -Rf ./hub_scanner/scan.cli
	unzip ./hub_scanner/scan.cli-$(BDS_SCANNER).zip -d ./hub_scanner/scan.cli
	\cp ../LICENSE ./
	
	docker build --rm --build-arg bds_ver=$(BDS_SCANNER)  --build-arg "LASTCOMMIT=$(LAST_COMMIT)" --build-arg "BUILDTIME=$(BUILD_TIME)" --build-arg "BUILD=$(BUILD_TAG)" -t hub_ose_scanner:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_scanner:$(BDS_SCANNER) > ./output/hub_ose_scanner.tar

	rm -Rf ./hub_scanner/scan.cli

travis:
	mkdir output
	GOOS=linux GOARCH=amd64 go build -o ./output/ose_scanner .
