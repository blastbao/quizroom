package main

import (
	"chatroom/libs/net/xrpc"
	"time"
	"chatroom/libs/proto"
	"github.com/name5566/leaf/log"
	"encoding/json"
	"chatroom/libs/define"
)

const (
	syncCountDelay = 1 * time.Second
)

var (
	RoomCountMap   = make(map[int32]int32) // roomid:count
	ServerCountMap = make(map[int32]int32) // server:count
)

func MergeCount() {
	var (
		c                     *xrpc.Clients
		err                   error
		roomId, server, count int32
		counter               map[int32]int32
		roomCount             = make(map[int32]int32)
		serverCount           = make(map[int32]int32)
	)
	// all comet nodes
	for _, c = range routerServiceMap {
		if c != nil {
			if counter, err = allRoomCount(c); err != nil {
				continue
			}
			for roomId, count = range counter {
				roomCount[roomId] += count
			}
			if counter, err = allServerCount(c); err != nil {
				continue
			}
			for server, count = range counter {
				serverCount[server] += count
			}
		}
	}
	RoomCountMap = roomCount
	ServerCountMap = serverCount
}

/*
func RoomCount(roomId int32) (count int32) {
	count = RoomCountMap[roomId]
	return
}
*/

func SyncCount() {
	for {
		MergeCount()
		time.Sleep(syncCountDelay)
	}
}

func SyncRoomCount() {
	timer := time.NewTicker(Conf.ROOMCOUNTERTIMER)

	select {
	case <-timer.C:
		BroadcastRoomCount()
	}

}

/**
	广播房间人数
 */
func BroadcastRoomCount() {
	if len(RoomCountMap) == 0 {
		return
	}
	for roomId, counter := range RoomCountMap {
		channelId, err := getChannelIdByRoomId(roomId)
		if err != nil {
			log.Error("BroadcastRoomCount get channel_id by room_id fail room_id : %v, counter: %v  error: %v", roomId, counter, err)
			continue
		}
		msg := proto.BroadcastRoomCounter{}
		msg.RoomId = roomId
		msg.Counter = counter
		msg.ChannelId = channelId

		vByte, err := json.Marshal(msg)
		if err != nil {
			log.Error("BroadcastRoomCount json Marshal error room_id : %v, channelId %v,counter: %v  error: %v", roomId, channelId, counter, err)
			continue
		}
		if err := broadcastRoomKafka(roomId, vByte, false, define.OP_BRAOADCAST_ROOM_COUNTER, 1, define.PROTO_VER); err != nil {
			log.Error("BroadcastRoomCount send fail; room_id : %v,channelId %v, counter: %v  error: %v", roomId, channelId, counter, err)
			continue
		}
		log.Debug("BroadcastRoomCount send success; room_id : %v, channel_id: %v,  counter: %v ", roomId, channelId, counter)
	}
}
