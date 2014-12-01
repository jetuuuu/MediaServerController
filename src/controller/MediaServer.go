package controller

import proto "code.google.com/p/goprotobuf/proto"

import (
	"fmt"
	"net"
	"./mediaserver"
	"time"
	//"strings"
)

const (
	BUFFSIZE = 4096
	MG_ID = 1
)

type mediaObject struct {
	mg, mixer_id, endpoint_id uint32
}

type MediaServer struct {
	connectUser net.Conn
	connectServer net.Conn
	info mediaObject
}

func (this *MediaServer) Start() {
	err := this.connectToServer("4000")
	if err != nil {
		return
	}
	//go this.waitClose()

	response := this.createVideoMix()
	if response == nil {
		fmt.Println("VideoMix can not be created.", this.connectUser.RemoteAddr().String())
	}

	this.info.mg = MG_ID
	this.info.mixer_id = response.Obj.GetId()
	fmt.Println("mixer id:", this.info.mixer_id)

	response = this.createEndPoint()
	if response == nil {
		fmt.Println("EndPoint can not be created.", this.connectUser.RemoteAddr().String())
	}

	this.info.endpoint_id = response.Obj.GetId()
	fmt.Println("endpoint id:", this.info.endpoint_id)

	localMedia := this.getEndPointMedia()
	_, err = this.connectUser.Write([]byte(localMedia.GetIp()))
	if err != nil {
		this.connectUser.Close()
		this.detachEndPoint()
		return
	}

	this.attachEndPoint()
	this.setReceiver();

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

func (this *MediaServer) setReceiver() {
	_, err := this.sendRequest(&mediaserver.MediaServerReq{
		Command: mediaserver.MediaServerReq_SetReceiver.Enum(),
		Obj: &mediaserver.MediaServerObject{Mg: &this.info.mg, Id: &this.info.endpoint_id,},
		Params: []*mediaserver.MediaServerParam{
			&mediaserver.MediaServerParam{Obj: &mediaserver.MediaServerObject{Mg: &this.info.mg, Id: &this.info.mixer_id,},},
		},
	})

	if err != nil {
		fmt.Println(err)
	}

}

func (this *MediaServer) connectToServer(port string) error {

	var buf [1024]byte
	addr, _ := net.ResolveUDPAddr("udp", ":4000")
	sock, _ := net.ListenUDP("udp", addr)
	rlen, remote, _ := sock.ReadFromUDP(buf[:])
	fmt.Println("Connect from:", remote)
	fmt.Println(rlen, remote)
	sock.Close()
	//ipPort := string(buf[:rlen])
	connect, err := net.Dial("tcp", remote.IP.String() + ":43510")
	fmt.Println("-----")
	if err != nil {
		fmt.Println(err.Error())
	}

	this.connectServer = connect
	//go checkServerConnection(connection)

	return nil
}

func checkServerConnection(connect net.Conn) {
	for {
		buff := make([]byte, BUFFSIZE)
		_, err := connect.Read(buff)
		if err != nil {
			fmt.Println(err)
			//закрыть все соединения и выйти ?!
			break
		}
	}
}

func (this *MediaServer) getEndPointMedia() *mediaserver.Media {
	response, err := this.sendRequest(&mediaserver.MediaServerReq{
		Command: mediaserver.MediaServerReq_GetEndPointMedia.Enum(),
		Obj: &mediaserver.MediaServerObject{Mg: &this.info.mg, Id:&this.info.endpoint_id, },
	})

	if err != nil {
		fmt.Println(err)
	}

	return response.GetMedia()
}

func (this *MediaServer) createVideoMix() *mediaserver.MediaServerRep {
	response, err := this.sendRequest(&mediaserver.MediaServerReq{
		Command: mediaserver.MediaServerReq_CreateVideoMix.Enum(),
	})

	if err != nil {
		fmt.Println(err)
	}

	return response
}

func (this *MediaServer) attachEndPoint() *mediaserver.MediaServerRep {
	response, err := this.sendRequest(&mediaserver.MediaServerReq{
		Command: mediaserver.MediaServerReq_AttachEndPoint.Enum(),
		Obj: &mediaserver.MediaServerObject{Mg: &this.info.mg, Id: &this.info.endpoint_id, },
		Params: []*mediaserver.MediaServerParam{
 			&mediaserver.MediaServerParam{Obj: &mediaserver.MediaServerObject{Mg: &this.info.mg, Id: &this.info.mixer_id,},},
		},
	})

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return response
}

func (this *MediaServer) createEndPoint() *mediaserver.MediaServerRep {
	response, err := this.sendRequest(&mediaserver.MediaServerReq{
		Command: mediaserver.MediaServerReq_CreateEndPoint.Enum(),
	})

	if err != nil {
		fmt.Println(err)
	}

	return response
}

func (this *MediaServer) sendRequest(request *mediaserver.MediaServerReq) (*mediaserver.MediaServerRep, error) {
	data, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	this.connectServer.Write(data)

	return this.getMediaServerRep(), nil
}

func (this *MediaServer) getMediaServerRep() *mediaserver.MediaServerRep {
	buff := make([]byte, BUFFSIZE)
	n, err := this.connectServer.Read(buff)
	if err != nil {
		fmt.Println("Cann't read server response. ", err.Error())
		return nil
	}
	response := new(mediaserver.MediaServerRep)
	err = proto.Unmarshal(buff[0: n], response)
	if err != nil {
		fmt.Println("Protobuf error unmarshal")
		return nil
	}
	return response
}
