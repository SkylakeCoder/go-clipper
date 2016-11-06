package main

import (
	"clipper"
	"flag"
)

var opType = flag.Uint("op", 0, "-op")
var path = flag.String("path", "", "-path")

func main() {
	flag.Parse()
	client := clipper.NewClient()
	op := clipper.OpType(*opType)
	client.StartUp(op, *path)
}
