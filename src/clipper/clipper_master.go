package clipper

import (
	"net"
	"log"
	"io"
	"encoding/binary"
	"sync"
	"time"
	"encoding/json"
)

var _startTime time.Time

type clipperInfo struct {
	path string
	time float64
}

type master struct {
	connections  []net.Conn
	lastCopyConn net.Conn
	mutex        sync.Mutex
	info         map[net.Conn]clipperInfo
}

func init() {
	_startTime = time.Now()
}

func NewMaster() *master {
	return &master{
		info: make(map[net.Conn]clipperInfo),
	}
}

func (m *master) StartUp() {
	l, err := net.Listen("tcp", ":8686")
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go m.handleConnection(c)
	}
}

func (m *master) handleConnection(c net.Conn) {
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
		case MSG_REGISTER:
			m.handleMsgRegister(c, bytes)
		case MSG_SET_CLIPPER_INFO:
			m.handleMsgSetClipperInfo(c, bytes)
		case MSG_GET_CLIPPER_INFO:
			m.handleMsgGetClipperInfo(c, bytes)
		default:
			log.Fatalln("error msg type...", req.MsgID)
		}
	}
}

func (m *master) handleMsgRegister(c net.Conn, bytes []byte) {
	m.mutex.Lock()
	exist := false
	for _, v := range m.connections {
		if v == c {
			exist = true
			break
		}
	}
	if !exist {
		m.connections = append(m.connections, c)
	}
	m.mutex.Unlock()
}

func (m *master) handleMsgSetClipperInfo(c net.Conn, bytes []byte) {
	req := reqSetClipperInfo{}
	json.Unmarshal(bytes, &req)
	m.mutex.Lock()
	m.info[c] = clipperInfo{
		path: req.Data,
		time: time.Since(_startTime).Seconds(),
	}
	m.mutex.Unlock()
}

func (m *master) handleMsgGetClipperInfo(conn net.Conn, bytes []byte) {
	var tempTime float64
	path := ""
	addr := ""
	for c, v := range m.info {
		if v.time > tempTime {
			tempTime = v.time
			path = v.path
			addr = c.RemoteAddr().String()
		}
	}
	respBytes, _ := json.Marshal(&respGetClipperInfo{
		Path: path,
		Addr: addr,
	})
	conn.Write(respBytes)
}
