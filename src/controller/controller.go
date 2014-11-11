package controller

import (
 	"fmt"
	"net"
)

type MediaServerController struct {
	host string
	port string
	listener net.Listener
}

func (this *MediaServerController) Init(host string, port string) {
	this.host = host
	this.port = port
	this.listener, _ = net.Listen("tcp", ":" + this.port)
}

func (this *MediaServerController) Start() {
	for {
		connection, err := this.listener.Accept();
		if err != nil {
			fmt.Println("connection fail")
			continue
		}
		go handleRequest(connection)
	}
}

func handleRequest(connection net.Conn) {
	fmt.Println(connection.RemoteAddr().String())
	ms := new(MediaServer)
	ms.connectUser = connection
	ms.Start()
}
