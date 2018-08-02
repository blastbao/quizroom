package main

import (
	"chatroom/libs/proto"
	"chatroom/libs/define"
	"encoding/json"
	log "github.com/thinkboy/log4go"
	"os"
	"fmt"
	"crypto/md5"
	"encoding/hex"
	"io"
	"encoding/base64"
	"crypto/rand"
	"database/sql"
	"time"
)

type UserConnectPushRoomStruct struct {
	Uid           int64                `json:"uid"`
	Avatar        string               `json:"avatar"`
	Nickname      string               `json:"nickname"`
	Guid          string               `json:"guid"`
	Timestamp     int64                `json:"timestamp"`
	Level         int                  `json:"level"`
	RoomId        int32                `json:"room_id"`
	AudienceCount int64                `json:"audience_count"`
	AudienceList  []*BroadcastAudience `json:"audience_list"`
}

func UserConnectPushRoom(userInfo *proto.RedisUserInfo, roomId int32) (err error) {
	if roomId == define.NoRoom {
		return nil
	}
	var (
		pushBody  UserConnectPushRoomStruct
		pBtye     []byte
		channelId string
	)
	pushBody.Uid = userInfo.Id
	pushBody.Avatar = userInfo.Avatar
	pushBody.Guid = userInfo.Guid
	pushBody.Nickname = userInfo.Nickname
	pushBody.Level = userInfo.Level
	pushBody.Timestamp = GetMicroTime()
	pushBody.RoomId = roomId
	if channelId, err = getChannelIdByRoomId(roomId); err == nil {
		if err = BroadcastRedisIncreaseAudience(channelId, userInfo.Guid); err != nil {
			log.Error("BroadcastRedisIncreaseAudience is error", err)
		}
	}
	if pushBody.AudienceCount, err = BroadcastInfoAudienceCount(channelId); err != nil {
		return
	}
	if pushBody.AudienceList, err = BroadcastGetAudienceList(channelId); err != nil {
		return
	}
	if pBtye, err = json.Marshal(pushBody); err != nil {
		return
	}
	if err = broadcastRoomKafka(roomId, pBtye, true, define.OP_BRAOADCAST_USER_CONNECT, 0, define.PROTO_VER); err != nil {
		return
	}
	return
}

func UserDisconnectPushRoom(uid int64, roomId int32) (err error) {
	var guid string
	if guid, err = getGuidByUid(uid); err != nil {
		return
	}
	var userInfo proto.RedisUserInfo
	if err = GetRedisUserInfo(guid, &userInfo); err != nil {
		return
	}
	if userInfo.Id == 0 {
		return ErrUidCannotZero
	}
	var (
		pushBody  UserConnectPushRoomStruct
		pBtye     []byte
		channelId string
	)
	pushBody.Uid = userInfo.Id
	pushBody.Avatar = userInfo.Avatar
	pushBody.Guid = userInfo.Guid
	pushBody.Nickname = userInfo.Nickname
	pushBody.Level = userInfo.Level
	pushBody.Timestamp = GetMicroTime()
	pushBody.RoomId = roomId
	if channelId, err = getChannelIdByRoomId(roomId); err == nil {
		if err = BroadcastRedisDecreaseAudience(channelId, userInfo.Guid); err != nil {
			log.Error("BroadcastRedisDecreaseAudience is error", err)
		}
	}
	if pushBody.AudienceCount, err = BroadcastInfoAudienceCount(channelId); err != nil {
		return
	}
	if pushBody.AudienceList, err = BroadcastGetAudienceList(channelId); err != nil {
		return
	}
	if pBtye, err = json.Marshal(pushBody); err != nil {
		return
	}
	if err = broadcastRoomKafka(roomId, pBtye, true, define.OP_BRAOADCAST_USER_DISCONNECT, 0, define.PROTO_VER); err != nil {
		return
	}
	return
}

func GetBroadcastRoomUsers(rid int32) {

}

//弹幕写入日志文件
func AppendToFile(rid int32, content string) (err error) {
	var (
		fileName string
		f        *os.File
	)
	ridStr := fmt.Sprintf("%d", rid)
	fileName = Conf.BroadcastBulletScreenMessageDir + "/room_" + ridStr + ".log"
	// 以只写的模式，打开文件
	if f, err = os.OpenFile(fileName, os.O_WRONLY, 0644); err != nil {
		if f, err = os.Create(fileName); err != nil {
			return
		}
	}
	defer f.Close()
	// 查找文件末尾的偏移量
	n, _ := f.Seek(0, os.SEEK_END)
	// 从末尾的偏移量开始写入内容
	_, err = f.WriteAt([]byte(content+"\n"), n)
	return
}

//bulletscreen write to timeline comment
func BroadcastBulletScreenWriteToTimeLine(rid int32, msgBody *proto.BroadcastBody, sendTime int64) (err error) {
	var (
		GUID      string
		stmt      *sql.Stmt
		channelId string
		//result    sql.Result
	)
	db, err1 := getDb()
	if err1 != nil {
		err = err1
		return
	}
	defer db.Close()
	GUID = GetGuid()
	tx, err1 := db.Begin()
	if err1 != nil {
		log.Error("db.Begin is error", err1)
		err = err1
		return
	}
	if stmt, err = tx.Prepare("INSERT comment SET guid=?,feed_id=?,user_id=?,content=?,created_at=?"); err != nil {
		tx.Rollback()
		return
	}
	defer stmt.Close()
	if channelId, err = getChannelIdByRoomId(rid); err != nil {
		tx.Rollback()
		return
	}
	Now := time.Unix(sendTime, 0)
	if _, err = stmt.Exec(GUID, channelId, msgBody.Guid, msgBody.Content, Now.Format("2006-01-02 03:04:05")); err != nil {
		tx.Rollback()
		return
	}

	if _, err = tx.Exec("UPdate feed set comment_count=comment_count+1 where guid=?", channelId); err != nil {
		tx.Rollback()
		return
	}

	if err = BroadcastRedisInsertComment(GUID, msgBody, sendTime); err != nil {
		tx.Rollback()
		return
	}

	if err = BroadcastRedisTimeLineCommentIncrease(channelId); err != nil {
		tx.Rollback()
		return
	}

	if err = BroadcastRedisWriteComment(channelId, GUID, sendTime); err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Commit(); err != nil {
		return
	}
	return
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}
