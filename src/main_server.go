package main

import (
	"clipper"
	"flag"
	"os"
)

var masterAddr = flag.String("master", "", "-master")

func main() {
	flag.Parse()
	server := clipper.NewServer()
	server.StartUp(*masterAddr, os.Args[0])
}
