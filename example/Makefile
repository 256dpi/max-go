# export MACOSX_DEPLOYMENT_TARGET=10.9
# export SDKROOT=/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk
# export CPATH="/Applications/Xcode.app/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs/MacOSX.sdk/usr/include/";

build:
	go build -buildmode=c-shared -o maxgo
	mkdir -p ./maxgo.mxo/MacOS
	cp ./maxgo ./maxgo.mxo/MacOS
	cp ./static/Info.plist ./maxgo.mxo
	cp ./static/PkgInfo ./maxgo.mxo
	rm -rf ~/Documents/Max\ 8/Packages/maxgo/externals/maxgo.mxo
	cp -r ./maxgo.mxo ~/Documents/Max\ 8/Packages/maxgo/externals/

