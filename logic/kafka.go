package main

import (
	"encoding/json"
	"quizroom/libs/define"
	"quizroom/libs/encoding/binary"
	"quizroom/libs/proto"

	"github.com/Shopify/sarama"
	log "github.com/thinkboy/log4go"
)

const (
	KafkaPushsTopic = "KafkaPushsTopic"
)

var (
	producer sarama.AsyncProducer
)

func InitKafka(kafkaAddrs []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err = sarama.NewAsyncProducer(kafkaAddrs, config)
	go handleSuccess()
	go handleError()
	return
}

func handleSuccess() {
	var (
		pm *sarama.ProducerMessage
	)
	for {
		pm = <-producer.Successes()
		if pm != nil {
			log.Info("producer message success, partition:%d offset:%d key:%v valus:%s", pm.Partition, pm.Offset, pm.Key, pm.Value)
		}
	}
}

func handleError() {
	var (
		err *sarama.ProducerError
	)
	for {
		err = <-producer.Errors()
		if err != nil {
			log.Error("producer message error, partition:%d offset:%d key:%v valus:%s error(%v)", err.Msg.Partition, err.Msg.Offset, err.Msg.Key, err.Msg.Value, err.Err)
		}
	}
}

func mpushKafka(serverId int32, keys []string, msg []byte, operation int32, seqId int32, ver int16) (err error) {
	log.Debug("mpushKafka push keys :%v operation:%v body :%v", keys, operation, string(msg))
	var (
		vBytes []byte
		v      = &proto.KafkaMsg{
			OP:       define.KAFKA_MESSAGE_MULTI,
			ServerId: serverId,
			SubKeys:  keys,
			Msg:      msg,
			PreOp:    operation,
			PreSeq:   seqId,
			PreVer:   ver,
		}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	producer.Input() <- &sarama.ProducerMessage{Topic: KafkaPushsTopic, Value: sarama.ByteEncoder(vBytes)}
	return
}

func broadcastKafka(msg []byte, operation int32, seqId int32, ver int16) (err error) {
	log.Debug("broadcastKafka push all body :%v", string(msg))
	var (
		vBytes []byte
		v      = &proto.KafkaMsg{
			OP:     define.KAFKA_MESSAGE_BROADCAST,
			Msg:    msg,
			PreOp:  operation,
			PreSeq: seqId,
			PreVer: ver,
		}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	producer.Input() <- &sarama.ProducerMessage{Topic: KafkaPushsTopic, Value: sarama.ByteEncoder(vBytes)}
	return
}

func broadcastRoomKafka(rid int32, msg []byte, ensure bool, operation int32, seqId int32, ver int16) (err error) {
	log.Debug("broadcastRoomKafka push rid :%v operation:%v body :%v", rid, operation, string(msg))
	var (
		vBytes   []byte
		ridBytes [4]byte
		v        = &proto.KafkaMsg{
			OP:     define.KAFKA_MESSAGE_BROADCAST_ROOM,
			RoomId: rid,
			Msg:    msg,
			Ensure: ensure,
			PreOp:  operation,
			PreSeq: seqId,
			PreVer: ver,
		}
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	if err1 := AppendToFile(rid, string(vBytes));err1 != nil{
		log.Error("bullet screen write to file is error : %v ", err)
	}
	binary.BigEndian.PutInt32(ridBytes[:], rid)
	producer.Input() <- &sarama.ProducerMessage{Topic: KafkaPushsTopic, Key: sarama.ByteEncoder(ridBytes[:]), Value: sarama.ByteEncoder(vBytes)}
	return
}
