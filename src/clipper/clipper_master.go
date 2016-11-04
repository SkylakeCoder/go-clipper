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
	return &master{}
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
	msgIDBuf := make([]byte, 1)
	msgLenBuf := make([]byte, 4)
	for {
		_, err := c.Read(msgIDBuf)
		_, err = c.Read(msgLenBuf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}
		msgLen := binary.LittleEndian.Uint32(msgLenBuf)
		switch msgType(msgIDBuf[0]) {
		case MSG_REGISTER:
			m.handleMsgRegister(c, msgLen)
		case MSG_SET_CLIPPER_INFO:
			m.handleMsgSetClipperInfo(c, msgLen)
		case MSG_GET_CLIPPER_INFO:
			m.handleMsgGetClipperInfo(c, msgLen)
		default:
			log.Fatalln("error msg type...", msgIDBuf[0])
		}
	}
}

func (m *master) handleMsgRegister(c net.Conn, msgLen uint32) {
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

func (m *master) handleMsgSetClipperInfo(c net.Conn, msgLen uint32) {
	buf := make([]byte, msgLen)
	_, err := c.Read(buf)
	if err != nil {
		log.Fatalln(err)
	}
	m.mutex.Lock()
	m.info[c] = clipperInfo{
		path: string(buf),
		time: time.Since(_startTime).Seconds(),
	}
	m.mutex.Unlock()
}

func (m *master) handleMsgGetClipperInfo(conn net.Conn, msgLen uint32) {
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
	bytes, _ := json.Marshal(&respGetClipperInfo{
		path: path,
		addr: addr,
	})
	conn.Write(bytes)
}
