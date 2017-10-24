BDS_SCANNER ?= 3.6.2
BUILD_TIME := $(shell cat ../buildtime.txt)
LAST_COMMIT := $(shell git rev-parse HEAD)

all:
	mkdir -p output build

	GOOS=linux GOARCH=amd64 go build -o ./output/ose_scanner .

	rm -Rf ./hub_scanner/scan.cli
	curl -u "reader":"EGGlaptop4EGG_&" https://test-repo.blackducksoftware.com/artifactory/bds-hub-nightly/com/blackducksoftware/hub/scan.cli/${BDS_VER}/scan.cli-${BDS_VER}-zip.zip > ./hub_scanner/scan.cli-${BDS_VER}.zip
	unzip ./hub_scanner/scan.cli-$(BDS_SCANNER).zip -d ./hub_scanner/scan.cli
	if [ -d ./hub_scanner/scan.cli/scan.cli-$(BDS_SCANNER)-zip ]; then \cp -a ./hub_scanner/scan.cli/scan.cli-$(BDS_SCANNER)-zip/. ./hub_scanner/scan.cli/scan.cli-$(BDS_SCANNER); fi;
	\cp ../LICENSE ./
	
	docker build --rm --build-arg bds_ver=$(BDS_SCANNER)  --build-arg "LASTCOMMIT=$(LAST_COMMIT)" --build-arg "BUILDTIME=$(BUILD_TIME)" --build-arg "BUILD=$(BUILD_TAG)" -t hub_ose_scanner:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_scanner:$(BDS_SCANNER) > ./output/hub_ose_scanner.tar

	rm -Rf ./hub_scanner/scan.cli

travis:
	mkdir -p output
	GOOS=linux GOARCH=amd64 go build -o ./output/ose_scanner .

clean:
	rm -Rf ./output ./build
