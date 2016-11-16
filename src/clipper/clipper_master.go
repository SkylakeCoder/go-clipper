package clipper

import (
	"net"
	"log"
	"io"
	"encoding/binary"
	"sync"
	"time"
	"encoding/json"
	"strings"
	"fmt"
)

var _startTime time.Time

type clipperInfo struct {
	path string
	time float64
	port uint32
}

type master struct {
	connections  []net.Conn
	lastCopyConn net.Conn
	mutex        sync.Mutex
	info         map[net.Conn]clipperInfo
	currentPort  uint32
	portMutex    sync.Mutex
}

func init() {
	_startTime = time.Now()
}

func NewMaster() *master {
	return &master{
		info: make(map[net.Conn]clipperInfo),
		currentPort: 8686,
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
		_, err := io.ReadFull(c, msgLenBuf)
		if err != nil {
			log.Println(err)
			break
		}
		bytes := make([]byte, binary.LittleEndian.Uint32(msgLenBuf))
		_, err = io.ReadFull(c, bytes)
		if err != nil {
			log.Println(err)
			break
		}
		req := commonReq{}
		err = json.Unmarshal(bytes, &req)
		if err != nil {
			log.Println(err)
			break
		}
		log.Println("master: msgID=", req.MsgID, " remoteAddr=", c.RemoteAddr())
		switch msgType(req.MsgID) {
		case MSG_REGISTER:
			m.handleMsgRegister(c, bytes)
		case MSG_SET_CLIPPER_INFO:
			m.handleMsgSetClipperInfo(c, bytes)
		case MSG_GET_CLIPPER_INFO:
			m.handleMsgGetClipperInfo(c, bytes)
		case MSG_REQUEST_ASSIGN_PORT:
			m.handleMsgRequestAssignPort(c, bytes)
		default:
			log.Fatalln("master: error msg type...", req.MsgID)
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
		port: req.Port,
	}
	m.mutex.Unlock()
}

func (m *master) handleMsgGetClipperInfo(conn net.Conn, bytes []byte) {
	var tempTime float64
	addr := ""
	var info clipperInfo
	for c, v := range m.info {
		if v.time > tempTime {
			tempTime = v.time
			addr = c.RemoteAddr().String()
			info = v
		}
	}
	split := strings.Split(addr, ":")
	fixedAddr := fmt.Sprintf("%s:%d", split[0], info.port)
	respBytes, _ := json.Marshal(&respGetClipperInfo{
		Path: info.path,
		Addr: fixedAddr,
	})
	conn.Write(respBytes)
}

func (m *master) handleMsgRequestAssignPort(conn net.Conn, bytes[]byte) {
	m.portMutex.Lock()
	m.currentPort++
	respBytes, _ := json.Marshal(&respAssignPort{
		Port: m.currentPort,
	})
	m.portMutex.Unlock()

	conn.Write(respBytes)
}
