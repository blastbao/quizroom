package proto

import "encoding/json"

type ConnArg struct {
	Token  string
	Server int32
	Guid   string
	RoomId int32
}

type ConnReply struct {
	Key    string
	RoomId int32
}

type DisconnArg struct {
	Key    string
	RoomId int32
}

type DisconnReply struct {
	Has bool
}

type BroadcastBody struct {
	Content         string `json:"content"`
	Guid            string `json:"guid"`
	Nickname        string `json:"nickname"`
	Level           int    `json:"level"`
	Avatar          string `json:"avatar"`
	RoomId          int32  `json:"room_id"`
	Timestamp       int64  `json:"timestamp"`
	ClientMessageId string `json:"client_message_id"`
	FontColor       string `json:"font_color"`
}

type RedisUserInfo struct {
	Token string `redis:"token"`
	Id    int64  `redis:"id"`
	Guid  string `redis:"guid"`
	//Cellphone	string	`redis:"cellphone"`
	Nickname string `redis:"nickname"`
	//Status		string	`redis:"status"`
	Avatar string `redis:"avatar"`
	Level  int    `redis:"level"`
}

type BroadcastLikeBody struct {
	ChannelId       string `json:"channel_id"`
	RoomId          int32  `json:"room_id"`
	LikeNum         int64  `json:"like_num"`
	Timestamp       int64  `json:"timestamp"`
	ClientMessageId string `json:"client_message_id"`
}

type BroadcastInfo struct {
	AudienceCount int64  `redis:"audience_count"`
	ChannelId     string `redis:"channel_id"`
	RoomId        int32  `redis:"short_id"`
	Guid          string `redis:"user_id"`
	LikeCount     int64  `redis:"like_count"`
	Type          int    `redis:"type"`
}

type BroadcastUserEvent struct {
	From    string          `json:"from"`
	To      string          `json:"to"`
	Type    int             `json:"type"`
	Content json.RawMessage `json:"content"`
}

type BroadcastRoomCounter struct {
	RoomId    int32  `json:"room_id"`
	ChannelId string `json:"channel_id"`
	Counter   int32  `json:"counter"`
}
