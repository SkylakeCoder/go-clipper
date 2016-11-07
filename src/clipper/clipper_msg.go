package clipper

import (
	"net"
	"encoding/json"
	"encoding/binary"
)

type msgType byte
type OpType byte
const MAX_BUFF uint32 = 1024 * 1024 * 2

const (
	OP_NULL OpType = iota
	OP_SET
	OP_GET
)

const (
	MSG_NULL msgType = iota
	MSG_REGISTER
	MSG_SET_CLIPPER_INFO
	MSG_GET_CLIPPER_INFO
	MSG_REQUEST_FILE
	MSG_REQUEST_ASSIGN_PORT
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
	Port uint32
}

type reqGetClipperInfo struct {
	commonReq
}

type respGetClipperInfo struct {
	Addr string
	Path string
}

type reqRequestFile struct {
	commonReq
	Path string
}

type reqAssignPort struct {
	commonReq
}

type respAssignPort struct {
	Port uint32
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

func sendSetClipperInfoReq(c net.Conn, data string, port uint32) {
	msg := reqSetClipperInfo{
		Data: data,
		Port: port,
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

func sendRequestFileReq(c net.Conn, path string) {
	msg := reqRequestFile{
		Path: path,
	}
	msg.MsgID = MSG_REQUEST_FILE
	bytes, _ := json.Marshal(&msg)
	c.Write(uintToBytes(len(bytes)))
	c.Write(bytes)
}

func sendRequestAssignPortReq(c net.Conn) {
	msg := reqAssignPort{}
	msg.MsgID = MSG_REQUEST_ASSIGN_PORT
	bytes, _ := json.Marshal(&msg)
	c.Write(uintToBytes(len(bytes)))
	c.Write(bytes)
}