BDS_SCANNER := 3.4.0

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output

	rm -Rf ./build \
	mkdir build

	GOOS=linux GOARCH=amd64 go build -o ./output/ose_scanner .

	rm -Rf ./hub_scanner/scan.cli
	unzip ./hub_scanner/scan.cli-$(BDS_SCANNER).zip -d ./hub_scanner/scan.cli
	
	docker build --build-arg bds_ver=$(BDS_SCANNER) -t hub_ose_scanner:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_scanner:$(BDS_SCANNER) > ./output/hub_ose_scanner.tar