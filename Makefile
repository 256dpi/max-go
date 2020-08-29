sdk:
	git clone https://github.com/Cycling74/max-sdk.git

fmt:
	go fmt .
	go vet .
	golint .
	clang-format  -style "{BasedOnStyle: Google, ColumnLimit: 120}" -i max/*.c max/*.h
