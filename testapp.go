package main

import (
	"./src/controller/"
)

func main() {
	msc := new(controller.MediaServerController)
	msc.Init("localhost", "5000")
	msc.Start()
}
