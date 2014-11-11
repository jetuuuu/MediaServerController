package main

import (
	"./src/controller/"
)

func main() {
	msc := new(controller.MediaServerController)
	host := "192.168.0.101"
	msc.Init(host, "5000")
	msc.Start()
}
