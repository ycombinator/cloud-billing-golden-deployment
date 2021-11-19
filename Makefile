.PHONY=build dist

BINARY=ecbgd

build:
	mkdir -p bin/
	go build -o bin/${BINARY} .

clean:
	rm -f bin/*
	rm -rf data/scenarios/*