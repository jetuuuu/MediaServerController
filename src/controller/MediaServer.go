package controller

import proto "code.google.com/p/goprotobuf/proto"

import (
	"fmt"
	"net"
	"./test"
	"time"
)

const (
	BUFFSIZE = 4096
)

type MediaServer struct {
	connectUser net.Conn
	connectServer net.Conn
}

func (this *MediaServer) Start() {
	err := this.connectToServer("0.0.0.0", "4000")
	if err != nil {
		return
	}
	go this.waitClose()

	response := this.createVideoMix()
	if response == nil {
		fmt.Println("VideoMix can not be created.", this.connectUser.LocalAddr().String())
	}

	mixer_id := response.GetId()
	fmt.Println("mixer id:", mixer_id)

	response = this.createEndPoint()
	if response == nil {
		fmt.Println("EndPoint can not be created.", this.connectUser.LocalAddr().String())
	}
	endpoint_id := response.GetId()
	fmt.Println("endpoint id:", response.GetId())

	localMedia := this.getEndPointMedia(&endpoint_id)
	_, err = this.connectUser.Write([]byte(localMedia.GetIp()))
	if err != nil {
		this.connectUser.Close()
		this.detachEndPoint()
		return
	}

	this.attachEndPoint(mixer_id, endpoint_id)
	this.setReceiver(endpoint_id, mixer_id);

	//this.connectUser.Close()
	//this.connectServer.Close()
}

func (this *MediaServer) waitClose() {
	for {
		_, err := this.connectUser.Write([]byte("~"))
		if err != nil {
			this.connectUser.Close()
			this.detachEndPoint()
			break
		}
		time.Sleep(time.Second)
	}
}

func (this *MediaServer) detachEndPoint() {
	fmt.Println("detachEndPoint")
}

func (this *MediaServer) setReceiver(endpoint_id string, mixer_id string) {
	_, err := this.sendRequest(&test.MediaServerReq{
		Command: test.MediaServerReq_SetReceiver.Enum(),
		Id: &endpoint_id,
		Params: []string{mixer_id},
	})

	if err != nil {
		fmt.Println(err)
	}

}

func (this *MediaServer) connectToServer(ip string, port string) error {
	connection, err := net.Dial("tcp", ip + ":" + port)
	if err != nil {
		return err
	}
	buff := make([]byte, 1024)
	n, err := connection.Read(buff)
	if err != nil {
		fmt.Println(err)
		return err
	}
	ipPort := string(buff[:n])
	this.connectServer, err = net.Dial("tcp", ipPort)
	if err != nil {
		fmt.Println(err)
		return err
	}

	go checkServerConnection(connection)

	return nil
}

func checkServerConnection(connect net.Conn) {
	for {
		buff := make([]byte, BUFFSIZE)
		_, err := connect.Read(buff)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func (this *MediaServer) getEndPointMedia(endpoint_id *string) *test.Media {
	response, err := this.sendRequest(&test.MediaServerReq{
		Command: test.MediaServerReq_GetEndPointMedia.Enum(),
		Id: endpoint_id,
	})

	if err != nil {
		fmt.Println(err)
	}

	return response.GetMedia()
}

func (this *MediaServer) createVideoMix() *test.MediaServerRep {
	response, err := this.sendRequest(&test.MediaServerReq{
		Command: test.MediaServerReq_CreateVideoMix.Enum(),
	})

	if err != nil {
		fmt.Println(err)
	}

	return response
}
//TODO:уточнить по поводу того, как передавать id endpoint'a
func (this *MediaServer) attachEndPoint(mixer_id string, endpoint_id string) *test.MediaServerRep {
	response, err := this.sendRequest(&test.MediaServerReq{
		Command: test.MediaServerReq_AttachEndPoint.Enum(),
		Id: &endpoint_id,
		Params:  []string{mixer_id},
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return response
}

func (this *MediaServer) createEndPoint() *test.MediaServerRep {
	response, err := this.sendRequest(&test.MediaServerReq{
		Command: test.MediaServerReq_CreateEndPoint.Enum(),
	})

	if err != nil {
		fmt.Println(err)
	}

	return response
}

func (this *MediaServer) sendRequest(request *test.MediaServerReq) (*test.MediaServerRep, error) {
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	this.connectServer.Write(data)

	return this.getMediaServerRep(), nil
}

func (this *MediaServer) getMediaServerRep() *test.MediaServerRep {
	buff := make([]byte, BUFFSIZE)
	n, err := this.connectServer.Read(buff)
	if err != nil {
		fmt.Println("Cann't read server response. ", this.connectServer.LocalAddr().Network())
		return nil
	}
	response := new(test.MediaServerRep)
	err = proto.Unmarshal(buff[0: n], response)
	if err != nil {
		fmt.Println("Protobuf error unmarshal")
		return nil
	}
	return response
}
