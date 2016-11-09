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
	"io"
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

func (c *client) getSavePath(srcPath string, destPath string) string {
	f, ferr := os.Open(destPath)
	if ferr == nil {
		defer f.Close()
	}
	fi, fierr := f.Stat()
	var srcPS, destPS string
	if strings.Contains(srcPath, "/") {
		srcPS = "/"
	} else {
		srcPS = "\\"
	}
	if strings.Contains(destPath, "/") {
		destPS = "/"
	} else {
		destPS = "\\"
	}
	split := strings.Split(srcPath, srcPS)
	fileName := split[len(split) - 1]
	if ferr == nil && fierr == nil && fi.IsDir() {
		destPath += destPS + fileName
	} else {
		split = strings.Split(destPath, destPS)
		path := ""
		for i := 0; i < len(split) - 1; i++ {
			path += split[i] + destPS
		}
		destPath = path + fileName
	}
	return destPath
}

func (c *client) requestFile(addr string, srcPath string, destPath string) {
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
	bufSize := make([]byte, 8)
	_, err = io.ReadFull(conn, bufSize)
	if err != nil {
		log.Fatalln(err)
	}
	fileSize := binary.LittleEndian.Uint64(bufSize)
	var currSize uint64 = 0
	savePath := c.getSavePath(srcPath, destPath)
	dummySavePath := savePath + ".tmp"
	f, err := os.Create(dummySavePath)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		bufLen := make([]byte, 4)
		io.ReadFull(conn, bufLen)
		l := binary.LittleEndian.Uint32(bufLen)
		buf := make([]byte, l)
		nr, err := io.ReadFull(conn, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			break
		}
		_, err = f.Write(buf[:nr])
		if (err != nil) {
			log.Fatalln(err)
		}
		currSize += uint64(nr)
		if currSize >= fileSize {
			break
		}
	}
	f.Close()
	err = os.Rename(dummySavePath, savePath)
	if err != nil {
		log.Fatalln(err)
	}
}
