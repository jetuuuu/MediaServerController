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

const (
	BUFFSIZE = 4096
)

type MediaServer struct {
	connect net.Conn
}

func (this *MediaServer) Start() {
	//TODO: добавить повторную попытку создать VideoMix
	response := this.createVideoMix()
	fmt.Println("mixer id:", response.GetId())
	endpointID := this.createEndPoint()
	fmt.Println("endpoint id:", endpointID)
	this.attachEndPoint(mixerID, endpointID)
	this.connect.Close()
}

func (this *MediaServer) createVideoMix() *test.MediaServerRep {
	response, err := this.sendRequest(&test.MediaServerReq{
		Command: test.ServerRequest_CreateVideoMix.Enum(),
	})

	if err != nil {
		panic("Protobuf panic")
	}

	//ждем ответа от сервера с id микшера
	return rand.Intn(100)
}

func (this *MediaServer) attachEndPoint(mixerID int, endpointID int) *test.MediaServerRep {
	err := this.sendRequest(&test.MediaServerReq{
		Command: test.ServerRequest_AttachEndPoint.Enum(),
		Params:  []string{strconv.Itoa(mixerID), strconv.Itoa(endpointID)},
	})

	if err != nil {
		panic("Protobuf panic")
	}
}

func (this *MediaServer) createEndPoint() *test.MediaServerRep {
	ip := this.connect.LocalAddr().String()

	err := this.sendRequest(&test.MediaServerReq{
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

func (this *MediaServer) sendRequest(request *test.MediaServerReq) (*test.MediaServerRep, error) {
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	this.connect.Write(data)

	return this.getMediaServerRep(),nil
}

func (this *MediaServer) getMediaServerRep() *test.MediaServerRep {
	//TODO: так же добавить возможность повторного запроса
	buff := make([]byte, BUFFSIZE)
	n, err := this.connect.Read(buff)
	if err != nil {
		fmt.Println("Cann't read server response. ", this.connect.LocalAddr().Network())
		return nil
	}
	response := new(MediaServerRep)
	return proto.Unmarshal(buff[0: n], response)
}

func (this *MediaServer) getID(ip string) string {
	hasher := md5.New()
	hasher.Write([]byte(ip))
	return hex.EncodeToString(hasher.Sum(nil))
}
