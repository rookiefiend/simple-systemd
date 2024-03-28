PWD=$(shell pwd)
output_dir = build
binaries = $(shell basename $(pwd))

.PHONY: build
build:
	mkdir -p build && go build -o ${output_dir}/ .


