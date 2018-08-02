package proto

// TODO optimize struct after replace kafka
type KafkaMsg struct {
	OP       string   `json:"op"`
	RoomId   int32    `json:"roomid,omitempty"`
	ServerId int32    `json:"server,omitempty"`
	SubKeys  []string `json:"subkeys,omitempty"`
	Msg      []byte   `json:"msg"`
	Ensure   bool     `json:"ensure,omitempty"`
	PreOp    int32    `json:"pre_op,omitempty"`
	PreSeq   int32    `json:"pre_seq,omitempty"`
	PreVer   int16    `json:"pre_ver,omitempty"`
}
