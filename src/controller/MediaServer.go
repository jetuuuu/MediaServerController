package controller

import proto "code.google.com/p/goprotobuf/proto"

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"./test"
)

type MediaServer struct {
	connect net.Conn
}

func (this *MediaServer) Start() {
	mixerID := this.createVideoMix()
	fmt.Println("mixer id:", mixerID)
	endpointID := this.createEndPoint()
	fmt.Println("endpoint id:", endpointID)
	this.attachEndPoint(mixerID, endpointID)
	this.connect.Close()
}

func (this *MediaServer) createVideoMix() int {
	err := this.sendRequest(&test.ServerRequest{
		Command: test.ServerRequest_CreateVideoMix.Enum(),
	})

	if err != nil {
		panic("Protobuf panic")
	}

	//ждем ответа от сервера с id микшера
	return rand.Intn(100)
}

func (this *MediaServer) attachEndPoint(mixerID int, endpointID int) {
	err := this.sendRequest(&test.ServerRequest{
		Command: test.ServerRequest_AttachEndPoint.Enum(),
		Params:  []string{strconv.Itoa(mixerID), strconv.Itoa(endpointID)},
	})

	if err != nil {
		panic("Protobuf panic")
	}
}

func (this *MediaServer) createEndPoint() int {
	ip := this.connect.LocalAddr().String()

	err := this.sendRequest(&test.ServerRequest{
		Command: test.ServerRequest_CreateEndPoint.Enum(),
		Sdp: &test.SDP{
			V: proto.String("0\n"),
			O: proto.String(this.getID(ip) + " 0 IN IP4 " + strings.Split(ip, ":")[0] + "\n"),
			C: proto.String("IN IP4 " + strings.Split(ip, ":")[0] + "\n"),
			M: proto.String("video " + strings.Split(ip, ":")[1] + " RTP/AVP\n"),
		},
	})

	if err != nil {
		panic("Protobuf panic")
	}

	return rand.Intn(100)
}

func (this *MediaServer) sendRequest(request *test.ServerRequest) error {
	data, err := proto.Marshal(request)
	if err != nil {
		return err
	}

	this.connect.Write(data)

	return nil
}

func (this *MediaServer) getID(ip string) string {
	hasher := md5.New()
	hasher.Write([]byte(ip))
	return hex.EncodeToString(hasher.Sum(nil))
}
