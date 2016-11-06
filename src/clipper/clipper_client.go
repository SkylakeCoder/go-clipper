package clipper

import (
	"net"
	"log"
	"io/ioutil"
	"os"
	"encoding/json"
	"strings"
)

type client struct {

}

func NewClient() *client {
	return &client{}
}

func (c *client) StartUp(op OpType, path string) {
	c.notifyClipperInfo(op, path)
}

func (c *client) notifyClipperInfo(op OpType, path string) {
	conn, err := net.Dial("tcp", "localhost:8686")
	if err != nil {
		log.Fatalln(err)
	}
	if op == OP_SET {
		sendSetClipperInfoReq(conn, path)
	} else {
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
	split := strings.Split(addr, ":")
	conn, err := net.Dial("tcp", split[0] + ":8687")
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
	split = strings.Split(srcPath, "/")
	fileName := split[len(split) - 1]
	if ferr == nil && fierr == nil && fi.IsDir() {
		destPath += "/" + fileName
	} else {
		split = strings.Split(destPath, "/")
		path := ""
		for i := 0; i < len(split) - 1; i++ {
			path += split[i] + "/"
		}
		destPath = path + "/" + fileName
	}
	log.Println("client: destPath=", destPath)
	ioutil.WriteFile(destPath, buf[:nr], 0644)
}
