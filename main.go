package main

import "C"

func main() {
	// not called by Max
}

//export ext_main
func ext_main(uintptr) {
	test()
}
