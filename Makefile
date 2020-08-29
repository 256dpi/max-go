.PHONY: lib

lib:
	rm -rf ./lib
	rm -rf ./sdk
	git clone https://github.com/Cycling74/max-sdk.git sdk
	mkdir lib
	cp -r sdk/source/c74support/max-includes/ ./lib/
	rm -rf sdk

fmt:
	go fmt ./...
	go vet ./...
	golint ./...
	clang-format  -style "{BasedOnStyle: Google, ColumnLimit: 120}" -i max/*.c max/*.h

install:
	go install ./cmd/maxgo

build: install
	cd example; maxgo -cross -name foo
