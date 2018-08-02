package main

import (
	"quizroom/libs/proto"
	"errors"
	log "github.com/thinkboy/log4go"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string, guid string, rid int32) (userInfo *proto.RedisUserInfo, userId int64, roomId int32, err error)
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

func (a *DefaultAuther) Auth(token string, guid string, rid int32) (userInfo *proto.RedisUserInfo, userId int64, roomId int32, err error) {
	var user proto.RedisUserInfo
	if err = GetRedisUserInfo(guid, &user); err != nil {
		//发生认证失败回执
		return
	}

	if token == user.Token && user.Id != int64(0) { //pass
		userId = user.Id
		roomId = rid
		log.Debug("login success! uid:%v, guid:%v, token:%v, room:%v", userId, guid, token, roomId)
	} else {
		//认证失败回执
		log.Debug("auth fail! guid:%v, token:%v", guid, token)
		err = errors.New("user no found")
		return
	}
	userInfo = &user
	return
}
