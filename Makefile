.PHONY=build dist

BINARY=ecbgd
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

build:
	go build -o ${BINARY} .

dist: build
	mkdir -p dist/${GOOS}/${GOARCH}
	mv ecbgd dist/${GOOS}/${GOARCH}/

clean:
	rm -f ${BINARY}
	rm -rf dist