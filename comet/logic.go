package main

import (
	inet "quizroom/libs/net"
	"quizroom/libs/net/xrpc"
	"quizroom/libs/proto"
	"time"

	log "github.com/thinkboy/log4go"
	"encoding/json"
	"errors"
)

var (
	logicRpcClient *xrpc.Clients
	logicRpcQuit   = make(chan struct{}, 1)

	logicService                   = "RPC"
	logicServicePing               = "RPC.Ping"
	logicServiceConnect            = "RPC.Connect"
	logicServiceDisconnect         = "RPC.Disconnect"
	logicServiceSendBroadSms       = "RPC.SendMsgBroadCast"
	logicServiceBroadcastLike      = "RPC.BroadcastLike"
	logicServiceBroadcastUserEvent = "RPC.BroadcastUserEvent"
)

func InitLogicRpc(addrs []string) (err error) {
	var (
		bind          string
		network, addr string
		rpcOptions    []xrpc.ClientOptions
	)
	for _, bind = range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		options := xrpc.ClientOptions{
			Proto: network,
			Addr:  addr,
		}
		rpcOptions = append(rpcOptions, options)
	}
	// rpc clients
	logicRpcClient = xrpc.Dials(rpcOptions)
	// ping & reconnect
	logicRpcClient.Ping(logicServicePing)
	log.Info("init logic rpc: %v", rpcOptions)
	return
}

func connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	var authBody proto.AuthBody
	if err = json.Unmarshal(p.Body, &authBody); err != nil {
		return
	}
	var (
		arg   = proto.ConnArg{Token: authBody.Token, Server: Conf.ServerId, Guid: authBody.Guid, RoomId: authBody.RoomId}
		reply = proto.ConnReply{}
	)
	if authBody.Guid == ""{
		err = errors.New("Guid can not empty")
	}
	if authBody.Token == ""{
		err = errors.New("token can not empty")
	}
	if err = logicRpcClient.Call(logicServiceConnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	key = reply.Key
	rid = reply.RoomId
	heartbeat = 5 * 60 * time.Second
	return
}

//send broadcast message
func SendSmsBroadCast(p *proto.Proto, key string, roomId int32) (body json.RawMessage, err error) {
	var (
		arg   = proto.SendSmsBroadcastArg{P: p, RoomId: roomId, Key: key}
		reply = proto.SendSmsBroadcastReply{}
	)

	if err = logicRpcClient.Call(logicServiceSendBroadSms, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceSendBroadSms, arg, err)
		return
	}
	vData, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}
	body = vData
	return
}

//user like
func BroadcastLike(p *proto.Proto, key string, roomId int32) (body json.RawMessage, err error) {
	var (
		arg   = proto.BroadcastLikeArg{P: p, RoomId: roomId, Key: key}
		reply = proto.BroadcastLikeReply{}
	)

	if err = logicRpcClient.Call(logicServiceBroadcastLike, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceBroadcastLike, arg, err)
		return
	}
	vData, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}
	body = vData
	return
}

//user event
func BroadcastUserEvent(p *proto.Proto, key string, roomId int32) (body json.RawMessage, err error) {
	var (
		arg   = proto.BroadcastUserEventArg{P: p, RoomId: roomId, Key: key}
		reply = proto.BroadcastUserEventReply{}
	)

	if err = logicRpcClient.Call(logicServiceBroadcastUserEvent, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceBroadcastUserEvent, arg, err)
		return
	}
	vData, err := json.Marshal(reply)
	if err != nil {
		return nil, err
	}
	body = vData
	return
}

func disconnect(key string, roomId int32) (has bool, err error) {
	var (
		arg   = proto.DisconnArg{Key: key, RoomId: roomId}
		reply = proto.DisconnReply{}
	)
	if err = logicRpcClient.Call(logicServiceDisconnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	has = reply.Has
	return
}
