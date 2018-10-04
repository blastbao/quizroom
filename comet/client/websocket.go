package main

import (
	"encoding/json"
	"time"

	log "github.com/thinkboy/log4go"
	"golang.org/x/net/websocket"
	"sync"
	"sync/atomic"
	"strconv"
)
var m *sync.RWMutex
var counter int32
func initWebsocket() {
	origin := "http://" + Conf.WebsocketAddr + "/sub"
	url := "ws://" + Conf.WebsocketAddr + "/sub"
	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Error("websocket.Dial(\"%s\") error(%v)", Conf.WebsocketAddr, err)
		return
	}
	defer conn.Close()
	atomic.AddInt32(&counter, 1)
	log.Debug("connect success.... %v", counter)
	proto := new(Proto)
	proto.Ver = 1
	// auth
	// test handshake timeout
	// time.Sleep(time.Second * 31)
	proto.Operation = OP_AUTH
	seqId := int32(0)
	proto.SeqId = seqId
	//proto.Body = []byte("{\"guid\":\"Y7BQIfZfURXSICsw\",\"room_id\":110,\"token\":\"6273bea6c4372ca200c8ca0c768fcdc6\"}")
	proto.Body = []byte("{\"guid\":\""+Conf.Guid+"\",\"room_id\":"+strconv.Itoa(Conf.RoomId)+",\"token\":\""+Conf.Token+"\"}")
	if err = websocketWriteProto(conn, proto); err != nil {
		log.Error("websocketWriteProto() error(%v)", err)
		return
	}
	if err = websocketReadProto(conn, proto); err != nil {
		log.Error("websocketReadProto() error(%v)", err)
		return
	}
	//log.Debug("auth ok, proto: %v", proto)
	seqId++
	// writer
	go func() {
		proto1 := new(Proto)
		for {
			// heartbeat
			proto1.Operation = OP_HEARTBEAT
			proto1.SeqId = seqId
			proto1.Body = nil
			if err = websocketWriteProto(conn, proto1); err != nil {
				log.Error("tcpWriteProto() error(%v)", err)
				return
			}
			// test heartbeat
			//time.Sleep(time.Second * 31)
			seqId++
			// op_test
			proto1.Operation = OP_TEST
			proto1.SeqId = seqId
			if err = websocketWriteProto(conn, proto1); err != nil {
				log.Error("tcpWriteProto() error(%v)", err)
				return
			}
			seqId++
			time.Sleep(10000 * time.Millisecond)
		}
	}()
	// reader
	for {
		if err = websocketReadProto(conn, proto); err != nil {
			log.Error("tcpReadProto() error(%v)", err)
			return
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			log.Debug("receive heartbeat")
			if err = conn.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
				log.Error("conn.SetReadDeadline() error(%v)", err)
				return
			}
		} else if proto.Operation == OP_TEST_REPLY {
			log.Debug("body: %s", string(proto.Body))
		} else if proto.Operation == OP_SEND_SMS_REPLY {
			log.Debug("body: %s", string(proto.Body))
		}
	}
}

func websocketReadProto(conn *websocket.Conn, p *Proto) error {
	_, er := json.Marshal(p)
	if er != nil{
		log.Debug("msg: %v", er)
	}

	return nil
	//return websocket.JSON.Receive(conn, p)
}

func websocketWriteProto(conn *websocket.Conn, p *Proto) error {
	if p.Body == nil {
		p.Body = []byte("{sssssssssssssssssssssssssssssssssssssssssss}")
	}
	//log.Debug("msg: %s", string(p.Body))
	return websocket.JSON.Send(conn, p)
}
