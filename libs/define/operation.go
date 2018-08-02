package define

const (
	// handshake
	OP_HANDSHAKE       = int32(0)
	OP_HANDSHAKE_REPLY = int32(1)
	// heartbeat
	OP_HEARTBEAT       = int32(2)
	OP_HEARTBEAT_REPLY = int32(3)
	// send text messgae
	OP_SEND_SMS       = int32(4)
	OP_SEND_SMS_REPLY = int32(5)
	// kick user
	OP_DISCONNECT_REPLY = int32(6)
	// auth user
	OP_AUTH       = int32(7)
	OP_AUTH_REPLY = int32(8)
	// handshake with sid
	OP_HANDSHAKE_SID       = int32(9)
	OP_HANDSHAKE_SID_REPLY = int32(10)
	// raw message
	OP_RAW = int32(11)
	// room
	OP_ROOM_READY = int32(12)
	// proto
	OP_PROTO_READY  = int32(13)
	OP_PROTO_FINISH = int32(14)

	// for test
	OP_TEST       = int32(254)
	OP_TEST_REPLY = int32(255)

	//broadcast sms
	OP_BRAOADCAST_SMS              = int32(300)
	OP_BRAOADCAST_SMS_REPLY        = int32(301)
	OP_BRAOADCAST_LIKE             = int32(500)
	OP_BRAOADCAST_LIKE_REPLY       = int32(501)
	OP_BRAOADCAST_USER_EVENT       = int32(510)
	OP_BRAOADCAST_USER_EVENT_REPLY = int32(511)

	OP_BRAOADCAST_USER_CONNECT    = int32(1001)
	OP_BRAOADCAST_USER_DISCONNECT = int32(1003)
)
