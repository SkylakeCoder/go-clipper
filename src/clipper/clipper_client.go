package clipper

import (
	"net"
	"log"
	"io/ioutil"
	"os"
	"encoding/json"
	"strings"
	"encoding/binary"
	"fmt"
)

type client struct {
	selfPath string
	masterAddr string
}

func NewClient() *client {
	return &client{}
}

func (c *client) StartUp(op OpType, path string, masterAddr string, selfPath string) {
	c.selfPath = selfPath
	c.masterAddr = masterAddr
	c.notifyClipperInfo(op, path)
}

func (c *client) notifyClipperInfo(op OpType, path string) {
	if op == OP_SET {
		tmpPath := strings.Replace(c.selfPath, ".exe", "", -1)
		tmpPath = strings.Replace(tmpPath, "main_client", "tmp.d", -1)
		buf_port, err := ioutil.ReadFile(tmpPath)
		if err != nil {
			log.Fatalln(err)
		}
		port := binary.LittleEndian.Uint32(buf_port)
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			log.Fatalln(err)
		}
		sendSetClipperInfoReq(conn, path, 0)
	} else {
		conn, err := net.Dial("tcp", c.masterAddr)
		if err != nil {
			log.Fatalln(err)
		}
		sendGetClipperInfoReq(conn)
		buf := make([]byte, 1024)
		nr, err := conn.Read(buf)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("client:", string(buf[:nr]))
		resp := respGetClipperInfo{}
		json.Unmarshal(buf[:nr], &resp)
		c.requestFile(resp.Addr, resp.Path, path)
	}
}

func (c *client) requestFile(addr string, srcPath string, destPath string) {
	log.Println("client: destPath=", destPath)
	fixedAddr := addr
	if strings.Contains(addr, "127.0.0.1") {
		split := strings.Split(addr, ":")
		serverPort := split[1]
		split = strings.Split(c.masterAddr, ":")
		fixedAddr = split[0] + ":" + serverPort
	}
	conn, err := net.Dial("tcp", fixedAddr)
	if err != nil {
		log.Fatalln(err)
	}
	sendRequestFileReq(conn, srcPath)
	buf := make([]byte, MAX_BUFF)
	nr, err := conn.Read(buf)
	if err != nil {
		log.Fatalln(err)
	}

	f, ferr := os.Open(destPath)
	if ferr == nil {
		defer f.Close()
	}
	fi, fierr := f.Stat()
	ps := string(os.PathSeparator)
	split := strings.Split(srcPath, ps)
	fileName := split[len(split) - 1]
	if ferr == nil && fierr == nil && fi.IsDir() {
		destPath += ps + fileName
	} else {
		split = strings.Split(destPath, ps)
		path := ""
		for i := 0; i < len(split) - 1; i++ {
			path += split[i] + ps
		}
		destPath = path + ps + fileName
	}
	log.Println("client: destPath=", destPath)
	ioutil.WriteFile(destPath, buf[:nr], 0644)
}
