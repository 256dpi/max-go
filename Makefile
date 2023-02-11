.PHONY: lib

lib:
	rm -rf ./lib
	rm -rf ./sdk
	git clone https://github.com/Cycling74/max-sdk-base.git sdk
	mkdir -p lib/max lib/msp
	cp -r sdk/c74support/max-includes/ ./lib/max
	cp -r sdk/c74support/msp-includes/ ./lib/msp
	rm -rf sdk

fmt:
	go fmt ./...
	go vet ./...
	golint ./...
	clang-format  -style "{BasedOnStyle: Google, ColumnLimit: 120}" -i *.c *.h

install:
	go install ./cmd/maxgo

build: install
	cd example; maxgo -name maxgo -install maxgo

build-cross: install
	cd example; maxgo -name maxgo -cross -install maxgo
