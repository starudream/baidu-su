package main

import (
	"flag"
	"os"

	"github.com/go-sdk/lib/app"
	"github.com/go-sdk/lib/log"
)

var (
	Path string
	Help bool
)

func init() {
	flag.StringVar(&Path, "config", "config.json", "config")
	flag.BoolVar(&Help, "help", false, "help")
	flag.Parse()

	if Help {
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	a := app.New("baidu-su")
	defer a.Recover()

	a.Add(StartJob)

	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}
}
