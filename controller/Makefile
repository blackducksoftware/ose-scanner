BDS_SCANNER := 3.4.0

all:
	#clean up from the past
	rm -Rf ./output \
	mkdir output

	GOOS=linux GOARCH=amd64 go build -o ./output/controller ./cmd/controller
	
	docker build --build-arg bds_ver=$(BDS_SCANNER) -t hub_ose_controller:$(BDS_SCANNER) -f Dockerfile .
	docker save hub_ose_controller:$(BDS_SCANNER) > ./output/hub_ose_controller.tar