
all:
	rm -Rf ./output; mkdir ./output;
	cd ./scanner; make
	cd ./controller; make

	#copy the results up to our output
	cp -a ./scanner/output/*.tar ./output; cp -a ./controller/output/*.tar ./output
