package clipper

import (
	"net"
	"log"
	"encoding/binary"
	"io"
	"encoding/json"
	"io/ioutil"
)

type server struct {
	connToMaster net.Conn
}

func NewServer() *server {
	return &server{}
}

func (s *server) StartUp(masterAddr string) {
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
	l, err := net.Listen("tcp", ":8687")
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
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}
		req := commonReq{}
		err = json.Unmarshal(bytes, &req)
		if err != nil {
			log.Fatalln(err)
		}
		switch msgType(req.MsgID) {
		case MSG_REQUEST_FILE:
			s.handleRequestFile(c, bytes)
		default:
			log.Fatalln("server: error msg type...", req.MsgID)
		}
	}
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
