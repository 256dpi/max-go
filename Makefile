sdk:
	git clone https://github.com/Cycling74/max-sdk.git

fmt:
	go fmt .
	go vet .
	golint .
	clang-format -i max/*.c max/*.h
