package clipper

import (
	"net"
	"log"
	"encoding/binary"
	"io"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"strings"
)

type server struct {
	connToMaster net.Conn
	selfPath string
	port uint32
}

func NewServer() *server {
	return &server{}
}

func (s *server) StartUp(masterAddr string, selfPath string) {
	s.selfPath = selfPath
	s.connectToMaster(masterAddr)
	s.startServe()
}

func (s *server) connectToMaster(masterAddr string) {
	var err error
	s.connToMaster, err = net.Dial("tcp", masterAddr)
	if err != nil {
		log.Fatalln(err)
	}
	sendRegisterReq(s.connToMaster)
}

func (s *server) startServe() {
	sendRequestAssignPortReq(s.connToMaster)
	buf := make([]byte, 1024)
	nr, _ := s.connToMaster.Read(buf)
	resp := respAssignPort{}
	json.Unmarshal(buf[:nr], &resp)
	s.port = resp.Port
	buf_port := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf_port, resp.Port)
	path := strings.Replace(s.selfPath, ".exe", "", -1)
	path = strings.Replace(path, "main_server", "tmp.d", -1)
	ioutil.WriteFile(path, buf_port, 0644)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", resp.Port))
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go s.handleConnection(c)
	}
}

func (s *server) handleConnection(c net.Conn) {
	msgLenBuf := make([]byte, 4)
	for {
		_, err := c.Read(msgLenBuf)
		bytes := make([]byte, binary.LittleEndian.Uint32(msgLenBuf))
		_, err = io.ReadFull(c, bytes)
		if err != nil {
			break
		}
		req := commonReq{}
		err = json.Unmarshal(bytes, &req)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("master: msgID=", req.MsgID, " remoteAddr=", c.RemoteAddr())
		switch msgType(req.MsgID) {
		case MSG_SET_CLIPPER_INFO:
			s.handleSetClipperInfoReq(c, bytes)
		case MSG_REQUEST_FILE:
			s.handleRequestFile(c, bytes)
		default:
			log.Fatalln("server: error msg type...", req.MsgID)
		}
	}
}

func (s *server) handleSetClipperInfoReq(c net.Conn, bytes []byte) {
	req := reqSetClipperInfo{}
	json.Unmarshal(bytes, &req)
	sendSetClipperInfoReq(s.connToMaster, req.Data, s.port)
}

func (s *server) handleRequestFile(c net.Conn, bytes []byte) {
	req := reqRequestFile{}
	json.Unmarshal(bytes, &req)
	fbytes, err := ioutil.ReadFile(req.Path)
	if err != nil {
		log.Fatalln(err, ":", req.Path)
	}
	_, err = c.Write(fbytes)
	if err != nil {
		log.Fatalln(err)
	}
}
