package proto

type NoArg struct {
}

type NoReply struct {
}

type PushMsgArg struct {
	Key string
	P   Proto
}

type PushMsgsArg struct {
	Key    string
	PMArgs []*PushMsgArg
}

type PushMsgsReply struct {
	Index int32
}

type MPushMsgArg struct {
	Keys []string
	P    Proto
}

type MPushMsgReply struct {
	Index int32
}

type MPushMsgsArg struct {
	PMArgs []*PushMsgArg
}

type MPushMsgsReply struct {
	Index int32
}

type BoardcastArg struct {
	P Proto
}

type BoardcastRoomArg struct {
	RoomId int32
	P      Proto
}

type RoomsReply struct {
	RoomIds map[int32]struct{}
}

type AuthBody struct {
	Guid   string `json:"guid"`
	Token  string `json:"token"`
	RoomId int32  `json:"room_id"`
}

type SendSmsBroadcastArg struct {
	P      *Proto
	RoomId int32
	Key    string
}

type SendSmsBroadcastReply struct {
	Code int32
	Msg  string
}

type ReturnBody struct {
	Code int32
	Msg  string
}

type ClientReply struct {
	Code int32
	Msg  string
}

//user like
type BroadcastLikeArg struct {
	P      *Proto
	RoomId int32
	Key    string
}

type BroadcastLikeReply struct {
	Code int32
	Msg  string
}

//user event
type BroadcastUserEventArg struct {
	P      *Proto
	RoomId int32
	Key    string
}

type BroadcastUserEventReply struct {
	Code int32
	Msg  string
}
