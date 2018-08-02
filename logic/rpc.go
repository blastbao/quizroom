package main

import (
	inet "chatroom/libs/net"
	"chatroom/libs/proto"
	"net"
	"net/rpc"

	log "github.com/thinkboy/log4go"
	"encoding/json"
	"fmt"
	"time"
	"strconv"
	"chatroom/libs/define"
	"errors"
)

func InitRPC(auther Auther) (err error) {
	var (
		network, addr string
		c             = &RPC{auther: auther}
	)
	rpc.Register(c)
	for i := 0; i < len(Conf.RPCAddrs); i++ {
		log.Info("start listen rpc addr: \"%s\"", Conf.RPCAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.RPCAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go rpcListen(network, addr)
	}
	return
}

func rpcListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		log.Info("rpc addr: \"%s\" close", addr)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// RPC
type RPC struct {
	auther Auther
}

func (r *RPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

// Connect auth and registe login
func (r *RPC) Connect(arg *proto.ConnArg, reply *proto.ConnReply) (err error) {
	if arg == nil {
		err = ErrConnectArgs
		log.Error("Connect() error(%v)", err)
		return
	}
	var (
		uid      int64
		seq      int32
		userInfo *proto.RedisUserInfo
	)
	userInfo, uid, reply.RoomId, err = r.auther.Auth(arg.Token, arg.Guid, arg.RoomId)
	if err != nil {
		return
	}
	if seq, err = connect(uid, arg.Server, reply.RoomId); err == nil {
		reply.Key = encode(uid, seq)
	}
	if err = UserConnectPushRoom(userInfo, reply.RoomId); err != nil {
		log.Error("push room when user connect broadcast is error %v, uid:%v, guid:%v, nickname:%v", err, uid, userInfo.Guid, userInfo.Nickname)
	}
	return
}

func (r *RPC) SendMsgBroadCast(arg *proto.SendSmsBroadcastArg, reply *proto.SendSmsBroadcastReply) (err error) {
	var (
		p        = arg.P
		roomId   = arg.RoomId
		key      = arg.Key
		vByte    []byte
		userId   int64
		colorInt int
	)
	p.Operation = define.OP_BRAOADCAST_SMS_REPLY
	reply.Code = define.STATUS_SENDBROADCAST_FAIL
	reply.Msg = "fail"
	var msgBody proto.BroadcastBody
	if err = json.Unmarshal(p.Body, &msgBody); err != nil {
		return
	}
	msgBody.Timestamp = GetMicroTime()
	msgBody.RoomId = roomId
	if userId, _, err = decode(key); err != nil{
		return
	}
	fontColor := userId % 5
	if colorInt, err = strconv.Atoi(strconv.FormatInt(fontColor, 10)); err != nil{
		return
	}
	msgBody.FontColor = Conf.BroadcastBulletScreenColor[colorInt]

	if vByte, err = json.Marshal(msgBody); err != nil {
		return
	}
	if err = broadcastRoomKafka(roomId, vByte, true, p.Operation, p.SeqId, p.Ver); err != nil {
		return
	}
	sendTime := time.Now().Unix()
	//bulletscreen write to timeline comment
	if err = BroadcastBulletScreenWriteToTimeLine(roomId, &msgBody, sendTime); err != nil{
		log.Error("BroadcastBulletScreenWriteToTimeLine error %v rid: %v, msgbody: %+v", err, roomId, msgBody)
	}
	reply.Code = define.STATUS_SENDBROADCAST_SUCCESS
	reply.Msg = "success"
	return
}

func (r *RPC) BroadcastLike(arg *proto.BroadcastLikeArg, reply *proto.BroadcastLikeReply) (err error) {
	var (
		p      = arg.P
		roomId = arg.RoomId
		//key    = arg.Key
		vByte   []byte
		LikeNum int64
	)
	p.Operation = define.OP_BRAOADCAST_LIKE_REPLY
	reply.Code = define.STATUS_FAIL
	reply.Msg = "fail"
	var LikeBody proto.BroadcastLikeBody
	if err = json.Unmarshal(p.Body, &LikeBody); err != nil {
		return
	}

	if roomId != LikeBody.RoomId {
		err = errors.New("room_id is wrong")
		return
	}
	LikeNum, err = BroadcastRedisLike(LikeBody.ChannelId, LikeBody.LikeNum)
	if err != nil {
		return
	}
	LikeBody.LikeNum = LikeNum

	if vByte, err = json.Marshal(LikeBody); err != nil {
		return
	}
	if err = broadcastRoomKafka(roomId, vByte, false, p.Operation, p.SeqId, p.Ver); err != nil {
		return
	}
	reply.Code = define.STATUS_SUCCESS
	reply.Msg = "success"
	return
}

func (r *RPC) BroadcastUserEvent(arg *proto.BroadcastUserEventArg, reply *proto.BroadcastUserEventReply) (err error) {
	var (
		p = arg.P
		//key    = arg.Key
		//vByte []byte
		toUid int64
		res   map[int32][]string
	)
	p.Operation = define.OP_BRAOADCAST_USER_EVENT_REPLY
	reply.Code = define.STATUS_FAIL
	reply.Msg = "send event fail"
	var userEventBody proto.BroadcastUserEvent
	if err = json.Unmarshal(p.Body, &userEventBody); err != nil {
		return
	}

	var user proto.RedisUserInfo
	if err = GetRedisUserInfo(userEventBody.To, &user); err != nil {
		log.Error("get to guid is error: %v", userEventBody.To)
		return
	}
	//IM接收人ID
	toUid, err = strconv.ParseInt(fmt.Sprintf("%v", user.Id), 10, 64)
	if err != nil {
		fmt.Errorf("to user id in body is invalid")
		return
	}
	//获取接收用户router
	res = genSubKey(toUid)
	if len(res) == 0 {
		reply.Msg = "user was offline"
		return errors.New("user offline")
	}

	/*	if vByte, err = json.Marshal(userEventBody); err != nil {
			return
		}*/
	for serverID, keys := range res {
		log.Debug("send user event serverID: %v ; 返回结果keys: %v, 版本号: %v", serverID, keys, p.Ver)
		if err = mpushKafka(serverID, keys, p.Body, p.Operation, p.SeqId, p.Ver); err != nil {
			reply.Msg = "Send the abnormal"
			log.Error("err: %v", err)
			return
		}
	}
	reply.Code = define.STATUS_SUCCESS
	reply.Msg = "success"
	return
}

// Disconnect notice router offline
func (r *RPC) Disconnect(arg *proto.DisconnArg, reply *proto.DisconnReply) (err error) {
	if arg == nil {
		err = ErrDisconnectArgs
		log.Error("Disconnect() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, seq, err = decode(arg.Key); err != nil {
		log.Error("decode(\"%s\") error(%s)", arg.Key, err)
		return
	}
	if reply.Has, err = disconnect(uid, seq, arg.RoomId); err != nil {
		return
	}
	if reply.Has {
		if err = UserDisconnectPushRoom(uid, arg.RoomId); err != nil {
			return
		}
	}
	return
}

/**
Gets the message send timestamp
 */
func GetMicroTime() int64 {
	microTime := fmt.Sprintf("%d", time.Now().UnixNano())
	milTimeStr := microTime[:13]
	milTime, err := strconv.ParseInt(milTimeStr, 10, 64)
	if err != nil {
		log.Error("generation mirco-time is error: %v ", err)
	}
	return milTime

}
