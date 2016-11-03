package clipper

import (
	"net"
	"log"
	"io"
)

func StartServer() {
	l, err := net.Listen("tcp", "localhost:8686")
	if err != nil {
		log.Fatalln("listen error...", err)
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalln("accpet error...", err)
		}
		go handleConnection(c)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		log.Println("prepare to read...")
		nr, err := conn.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln("read error...", err)
		}
		log.Println("read n=", nr)
		log.Println("buf=", string(buf[:nr]))
		conn.Write(buf[:nr])
	}
}
