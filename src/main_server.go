package main

import (
	"clipper"
	"flag"
)

var masterAddr = flag.String("master", "", "-master")

func main() {
	flag.Parse()
	server := clipper.NewServer()
	server.StartUp(*masterAddr)
}
