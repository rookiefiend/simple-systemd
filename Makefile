PWD=$(shell pwd)
output_dir = build
binaries = $(shell basename $(pwd))

GOPROXY:=https://goproxy.cn,direct
GOARCH:=amd64
CGO_ENABLED:=0
export GOPROXY:=${GOPROXY}
export GOARCH:=${GOARCH}
export CGO_ENABLED:=${CGO_ENABLED}

buildImage = golang:1.22.1-alpine3.19

.PHONY: build
build:
	mkdir -p build && go build -o ${output_dir}/ .

docker-build:
	-docker pull ${buildImage}
	docker run --rm -it -v ${PWD}:/workspace -w /workspace \
		-e GOPROXY \
		-e GOARCH \
		-e CGO_ENABLED \
		${buildImage} sh -c "mkdir -p build && go build -o ${output_dir}/"