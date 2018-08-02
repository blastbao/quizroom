package main

import (
	"chatroom/libs/define"
	"chatroom/libs/proto"
	"time"

	log "github.com/thinkboy/log4go"
	"encoding/json"
)

type Operator interface {
	// Operate process the common operation such as send message etc.
	Operate(*proto.Proto, string, int32) error
	// Connect used for auth user and return a subkey, roomid, hearbeat.
	Connect(*proto.Proto) (string, int32, time.Duration, error)
	// Disconnect used for revoke the subkey.
	Disconnect(string, int32) error
}

type DefaultOperator struct {
}

func (operator *DefaultOperator) Operate(p *proto.Proto, key string, roomId int32) error {
	/*var (
		body []byte
	)*/
	if p.Operation == define.OP_SEND_SMS {
		// call suntao's api
		// p.Body = nil
		p.Operation = define.OP_SEND_SMS_REPLY
		log.Info("send sms proto: %v", p)
	} else if p.Operation == define.OP_TEST {
		log.Debug("test operation: %s", p.Body)
		p.Operation = define.OP_TEST_REPLY
		p.Body = []byte("{\"test\":\"come on\"}")
	} else if p.Operation == define.OP_BRAOADCAST_SMS {
		data, err := SendSmsBroadCast(p, key, roomId)
		if err != nil {
			errData := proto.SendSmsBroadcastReply{Code: define.STATUS_SENDBROADCAST_FAIL, Msg: "fail"}
			vData, err := json.Marshal(errData)
			if err != nil {
				return err
			}
			p.Body = vData
		} else {
			p.Body = data
		}
		log.Debug("send broadcast proto: %v", string(p.Body))
	} else if p.Operation == define.OP_BRAOADCAST_LIKE {
		data, err := BroadcastLike(p, key, roomId)
		if err != nil {
			errData := proto.ClientReply{Code: define.STATUS_FAIL, Msg: "like fail"}
			vData, err := json.Marshal(errData)
			if err != nil {
				return err
			}
			p.Body = vData
		} else {
			p.Body = data
		}
		log.Debug("broadcast like proto: %v", string(p.Body))
	}else if p.Operation == define.OP_BRAOADCAST_USER_EVENT {
		data, err := BroadcastUserEvent(p, key, roomId)
		if err != nil {
			errData := proto.ClientReply{Code: define.STATUS_FAIL, Msg: "send fail"}
			vData, err1 := json.Marshal(errData)
			if err1 != nil {
				return err1
			}
			p.Body = vData
		} else {
			p.Body = data
		}
		log.Debug("broadcast like proto: %v", string(p.Body))
	} else {
		log.Error("request operation not valid!: %v , op :%v, ver: %v, seq: %v", string(p.Body), p.Operation, p.Ver, p.SeqId)
		errData := proto.ClientReply{Code: define.STATUS_FAIL, Msg: "request operation not valid!"}
		vData, err := json.Marshal(errData)
		if err != nil {
			return err
		}
		p.Body = vData
		//return ErrOperation
	}
	return nil
}

func (operator *DefaultOperator) Connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	key, rid, heartbeat, err = connect(p)
	return
}

func (operator *DefaultOperator) Disconnect(key string, rid int32) (err error) {
	var has bool
	if has, err = disconnect(key, rid); err != nil {
		return
	}
	if !has {
		log.Warn("disconnect key: \"%s\" not exists", key)
	}
	return
}
