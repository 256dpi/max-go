package main

const pkgInfo = `iLaX????`

func infoPlist(name string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>English</string>
	<key>CFBundleExecutable</key>
	<string>` + name + `</string>
	<key>CFBundleIconFile</key>
	<string></string>
	<key>CFBundleIdentifier</key>
	<string>com.maxgo.` + name + `</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>1.0.0</string>
	<key>CFBundlePackageType</key>
	<string>iLaX</string>
	<key>CFBundleSignature</key>
	<string>max2</string>
	<key>CFBundleVersion</key>
	<string>1.0</string>

	<key>CFBundleVersion</key>
	<string>1.0.0</string>
	<key>CFBundleShortVersionString</key>
	<string>1.0.0</string>
	<key>CFBundleLongVersionString</key>
	<string>` + name + ` 1.0.0</string>

	<key>CSResourcesFileMapped</key>
	<true/>
</dict>
</plist>
`
}
