package main

import (
	"flag"
	"runtime"

	log "github.com/thinkboy/log4go"
)

func main() {
	flag.Parse()
	if err := InitConfig(); err != nil {
		panic(err)
	}
	runtime.GOMAXPROCS(Conf.MaxProc)
	log.LoadConfiguration(Conf.Log)
	defer log.Close()
	log.Debug("start...")
	if Conf.Type == ProtoTCP {
		initTCP()
	} else if Conf.Type == ProtoWebsocket {
		num := Conf.Bench
		for i := 0; i < num; i++ {
			go initWebsocket()
		}
		select {}
		log.Debug("end...")
	} else if Conf.Type == ProtoWebsocketTLS {
		initWebsocketTLS()
	}
}
