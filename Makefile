.PHONY: build run

build:
	env go build -o ./build/rdproxy ./...
	chmod +x ./build/rdproxy

build-ci:
	env GOOS=linux GOARCH=amd64 go build -o ./build/rdproxy-linux-amd64 -ldflags="-w -s" ./...
	env GOOS=linux GOARCH=arm64 go build -o ./build/rdproxy-linux-arm64 -ldflags="-w -s" ./...
	cd ./build && tar -zcvf ./rdproxy-linux-amd64.tar.gz ./rdproxy-linux-amd64
	cd ./build && tar -zcvf ./rdproxy-linux-arm64.tar.gz ./rdproxy-linux-arm64
