package controller

import proto "code.google.com/p/goprotobuf/proto"

import (
	"fmt"
	"net"
	"./test"
)

const (
	BUFFSIZE = 4096
)

type MediaServer struct {
	connectUser net.Conn
	connectServer net.Conn
}

func (this *MediaServer) Start() {
	//TODO: добавить повторную попытку создать VideoMix
	response := this.createVideoMix()
	if response == nil {
		//TODO: повторная попытка
	}
	mixer_id := response.GetId()
	fmt.Println("mixer id:", mixer_id)
	response = this.createEndPoint()
	if response == nil {
		//TODO: повторная попытка
	}
	endpoint_id := response.GetId()
	localMedia := this.getEndPointMedia(&endpoint_id)

	this.connectUser.Write([]byte(localMedia.GetIp()))
	
	fmt.Println("endpoint id:", response.GetId())
	this.attachEndPoint(mixer_id, endpoint_id)
	this.connectUser.Close()
	this.connectServer.Close()
}

func (this *MediaServer) getEndPointMedia(endpoint_id *string) *test.Media {
	response, err := this.sendRequest(&test.MediaServerReq{
		Command: test.MediaServerReq_GetEndPointMedia.Enum(),
		Id: endpoint_id,
	})

	if err != nill {
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
		Params:  []string{mixer_id, endpoint_id},
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
	//TODO: так же добавить возможность повторного запроса
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
