.PHONY=build dist

BINARY=ecbgd

build:
	go build -o ${BINARY} .

clean:
	rm -f ${BINARY}
