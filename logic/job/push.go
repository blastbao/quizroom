package main

import (
	"encoding/json"
	"quizroom/libs/define"
	"quizroom/libs/proto"
	"math/rand"

	log "github.com/thinkboy/log4go"
)

type pushArg struct {
	ServerId int32
	SubKeys  []string
	Msg      []byte
	RoomId   int32
	ProOp    int32
}

var (
	pushChs []chan *pushArg
)

func InitPush() {
	pushChs = make([]chan *pushArg, Conf.PushChan)
	for i := 0; i < Conf.PushChan; i++ {
		pushChs[i] = make(chan *pushArg, Conf.PushChanSize)
		go processPush(pushChs[i])
	}
}

// push routine
func processPush(ch chan *pushArg) {
	var arg *pushArg
	for {
		arg = <-ch
		mPushComet(arg.ServerId, arg.SubKeys, arg.Msg, arg.ProOp)
	}
}

func push(msg []byte) (err error) {
	m := &proto.KafkaMsg{}
	if err = json.Unmarshal(msg, m); err != nil {
		log.Error("json.Unmarshal(%s) error(%s)", msg, err)
		return
	}
	switch m.OP {
	case define.KAFKA_MESSAGE_MULTI:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{ServerId: m.ServerId, SubKeys: m.SubKeys, Msg: m.Msg, RoomId: define.NoRoom, ProOp: m.PreOp}
	case define.KAFKA_MESSAGE_BROADCAST:
		broadcast(m.Msg, m.PreOp, m.PreVer)
	case define.KAFKA_MESSAGE_BROADCAST_ROOM:
		room := roomBucket.Get(int32(m.RoomId))
		if m.Ensure {
			go room.EPush(m.PreVer, m.PreOp, m.Msg)
		} else {
			err = room.Push(m.PreVer, m.PreOp, m.Msg)
			if err != nil {
				log.Error("room.Push(%s) roomId:%d error(%v)", m.Msg, err)
			}
		}
	default:
		log.Error("unknown operation:%s", m.OP)
	}
	return
}
