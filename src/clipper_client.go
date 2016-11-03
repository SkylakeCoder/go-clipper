package main

import (
	"net"
	"log"
	"time"
)

func ConnectServer() {
	c, err := net.Dial("tcp", "localhost:8686")
	if err != nil {
		log.Fatalln("dial error...", err)
	}
	c.Write([]byte("hello"))
	buf := make([]byte, 1024)
	nr, err := c.Read(buf)
	if err != nil {
		log.Fatalln("client read error: ", err)
	}
	log.Println("receive: ", string(buf[:nr]))
	time.Sleep(time.Second * 6)
}

func main() {
	ConnectServer()
}