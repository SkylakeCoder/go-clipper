package clipper

import (
	"net"
	"encoding/json"
	"encoding/binary"
)

type msgType byte

const (
	MSG_NULL msgType = iota
	MSG_REGISTER
	MSG_SET_CLIPPER_INFO
	MSG_GET_CLIPPER_INFO
)

type commonReq struct {
	MsgID msgType
}

type reqRegister struct {
	commonReq
}

type reqSetClipperInfo struct {
	commonReq
	Data string
}

type reqGetClipperInfo struct {
	commonReq
}

type respGetClipperInfo struct {
	Addr string
	Path string
}

func uintToBytes(v int) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(v))
	return bytes
}

func sendRegisterReq(c net.Conn) {
	msg := reqRegister{}
	msg.MsgID = MSG_REGISTER
	bytes, _ := json.Marshal(&msg)
	c.Write(uintToBytes(len(bytes)))
	c.Write(bytes)
}

func sendSetClipperInfoReq(c net.Conn, data string) {
	msg := reqSetClipperInfo{
		Data: data,
	}
	msg.MsgID = MSG_SET_CLIPPER_INFO
	bytes, _ := json.Marshal(&msg)
	c.Write(uintToBytes(len(bytes)))
	c.Write(bytes)
}

func sendGetClipperInfoReq(c net.Conn) {
	msg := reqGetClipperInfo{}
	msg.MsgID = MSG_GET_CLIPPER_INFO
	bytes, _ := json.Marshal(&msg)
	c.Write(uintToBytes(len(bytes)))
	c.Write(bytes)
}