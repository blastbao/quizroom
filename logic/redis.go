package main

import (
	"github.com/garyburd/redigo/redis"
	"fmt"
	"time"
	"strconv"
	log "github.com/thinkboy/log4go"
	"quizroom/libs/proto"
)

var pool *redis.Pool

func InitRedis() {
	pool = &redis.Pool{
		MaxIdle:     Conf.RedisMaxIdle,     //池子里的最大空闲连接
		MaxActive:   Conf.RedisMaxActive,   // max number of connections
		IdleTimeout: Conf.RedisIdleTimeout, // 空闲连接存活时间
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", Conf.RedisAddr,
				//设置连接、读写超时，防止网络异常时长时间的等待。
				redis.DialConnectTimeout(15*time.Second),
				redis.DialReadTimeout(10*time.Second),
				redis.DialWriteTimeout(10*time.Second),
			)
			if err != nil {
				log.Error("redis connect error: %v, addr :%v", err, Conf.RedisAddr)
			}
			return c, err
		},
	}
}

func GetRedisUserInfo(guid string, user *proto.RedisUserInfo) (error error) {
	c := pool.Get()
	defer c.Close()
	value, err := redis.Values(c.Do("HGETALL", "user.id:"+guid))
	if err != err {
		log.Error("token %V is wrong HGETALL error: %v", guid, err)
		return
	}

	if err := redis.ScanStruct(value, user); err != nil {
		log.Error("redis ScanStruct error: %v", err)
		return
	}
	return
}

func GetRedisGroupUser(groupId string) (userIds map[string]int64, error error) {
	c := pool.Get()
	defer c.Close()
	num, err := c.Do("ZCARD", "group:members.id:"+groupId)
	if err != nil {
		log.Error("GetRedisGroupUser redis ZCARD error: %v", err)
		return
	}
	offMsgLen, err := strconv.ParseInt(fmt.Sprintf("%d", num), 10, 64)
	if err != nil {
		log.Error("GetRedisGroupUser len value error: %v", err)
		return
	}
	if offMsgLen == 0 {
		return
	}
	collect, err := redis.Strings(c.Do("ZRANGE", "group:members.id:"+groupId, 0, -1))
	if err != nil {
		fmt.Println(err)
		return
	}
	var user proto.RedisUserInfo
	userIds = make(map[string]int64, len(collect))
	for _, v := range collect {
		value, err := redis.Values(c.Do("HGETALL", "user.id:"+v))
		if err != err {
			log.Error("guid %V is wrong HGETALL error: %v", v, err)
			return
		}
		if err := redis.ScanStruct(value, &user); err != nil {
			log.Error("GetRedisGroupUser redis ScanStruct error: %v", err)
			return
		}
		//userIds = append(userIds, user.Id)
		userIds[user.Guid] = user.Id
		user = proto.RedisUserInfo{}
	}
	log.Debug("群: %v 成员%v", groupId, userIds)
	return
}

func GetRedisGroupUserShield(groupId string) (shieldUsers map[string]string, error error) {
	c := pool.Get()
	defer c.Close()
	num, err := c.Do("ZCARD", "group:shield.members.id:"+groupId)
	if err != nil {
		log.Error("GetRedisGroupUser redis ZCARD error: %v", err)
		return
	}
	offMsgLen, err := strconv.ParseInt(fmt.Sprintf("%d", num), 10, 64)
	if err != nil {
		log.Error("GetRedisGroupUser len value error: %v", err)
		return
	}
	if offMsgLen == 0 {
		return
	}
	collect, err := redis.Strings(c.Do("ZRANGE", "group:shield.members.id:"+groupId, 0, -1))
	if err != nil {
		log.Error("get group shield info error: %v", err)
		return
	}
	shieldUsers = make(map[string]string, len(collect))
	for _, v := range collect {
		shieldUsers[v] = v
	}
	return
}
func GetRedisSingleUserShield(Guid string) (shieldUsers map[string]string, error error) {
	c := pool.Get()
	defer c.Close()
	num, err := c.Do("ZCARD", "user:shield.id:"+Guid)
	if err != nil {
		log.Error("GetRedisSingleUserShield redis ZCARD error: %v", err)
		return
	}
	offMsgLen, err := strconv.ParseInt(fmt.Sprintf("%d", num), 10, 64)
	if err != nil {
		log.Error("GetRedisSingleUserShield len value error: %v", err)
		return
	}
	if offMsgLen == 0 {
		return
	}
	collect, err := redis.Strings(c.Do("ZRANGE", "user:shield.id:"+Guid, 0, -1))
	if err != nil {
		log.Error("get single shield info error: %v", err)
		return
	}
	shieldUsers = make(map[string]string, len(collect))
	for _, v := range collect {
		shieldUsers[v] = v
	}
	return
}

//删除zset信息
func RemoveOfflineSingleMessage(c redis.Conn, toUid string, v string) {
	//删除已发送的消息
	result, err := c.Do("ZREM", toUid, v)
	if err != nil {
		log.Error("ZREM error: %v", err)
	}
	log.Debug("zset remove toUid : %v, message id : %v, result :%v", toUid, v, result)
}

//删除message id
func RemoveMessageId(c redis.Conn, v string) {
	result, err := c.Do("DEL", "message.id:"+v)
	if err != nil {
		log.Error("OfflineMsgSingleChatRead delete messageIderror: %v", err)
	}
	log.Debug("remove message id : %v, result :%v", v, result)
}

//根据guid获取黑名单
func GetBlockList(guid string) (blacklist []string, err error) {
	c := pool.Get()
	defer c.Close()

	blacklist, err = redis.Strings(c.Do("ZRANGE", "block.id:"+guid, 0, -1))
	if err != nil {
		log.Debug("get redis zset data error: %v", err)
		return
	}
	return
}

func BroadcastRedisLike(channel_id string, increase int64) (likeNum int64, err error) {
	c := pool.Get()
	defer c.Close()
	_, err = c.Do("HINCRBY", "broadcast:info:id:"+channel_id, "like_count", increase)
	if err != nil {
		log.Error("redis HINCRBY error channel id %v, increase:%v", channel_id, increase)
		return
	}
	likeNum, err = redis.Int64(c.Do("HGET", "broadcast:info:id:"+channel_id, "like_count"))
	if err != nil {
		log.Error("redis HGET error channel id %v, increase:%v", channel_id, increase)
		return
	}
	return
}

func getGuidByUid(uid int64) (guid string, err error) {
	var uidStr string
	uidStr = strconv.FormatInt(uid, 10)
	c := pool.Get()
	defer c.Close()
	if guid, err = redis.String(c.Do("GET", "user.id.guid:"+uidStr)); err != nil {
		return
	}
	if guid == "" {
		return "", ErrGetGuid
	}
	return
}

func BroadcastRedisGetAudience(channelId string) (advanceNum int, err error) {
	c := pool.Get()
	defer c.Close()
	num, err := c.Do("ZCARD", "broadcast:audience:list.id:"+channelId)
	if err != nil {
		return
	}
	advanceNum, err = strconv.Atoi(fmt.Sprintf("%d", num))
	if err != nil {
		return
	}
	return
}

func BroadcastRedisIncreaseAudience(channelId string, guid string) (err error) {
	/*var advanceNum int
	if advanceNum, err = BroadcastRedisGetAudience(channelId); err != nil {
		return
	}*/
	c := pool.Get()
	defer c.Close()
	var BroadcastInfo proto.BroadcastInfo
	if BroadcastInfo, err = BroadcastRedisGetInfo(channelId); err != nil {
		return
	}
	if BroadcastInfo.Guid == guid {
		return
	}
	if _, err = c.Do("HINCRBY", "broadcast:info:id:"+channelId, "audience_count", 1); err != nil {
		return
	}
	/*if advanceNum >= 50 {
		return
	}*/
	if _, err = c.Do("ZADD", "broadcast:audience:list.id:"+channelId, GetMicroTime(), guid); err != nil {
		return
	}
	return
}
func BroadcastRedisDecreaseAudience(channelId string, guid string) (err error) {
	c := pool.Get()
	defer c.Close()
	var BroadcastInfo proto.BroadcastInfo
	if BroadcastInfo, err = BroadcastRedisGetInfo(channelId); err != nil {
		return
	}
	if BroadcastInfo.Guid == guid {
		return
	}
	if _, err = c.Do("HINCRBY", "broadcast:info:id:"+channelId, "audience_count", -1); err != nil {
		return
	}
	if _, err = c.Do("ZREM", "broadcast:audience:list.id:"+channelId, guid); err != nil {
		return
	}
	return
}

//根据room id 获取channel id
func getChannelIdByRoomId(roomId int32) (channelId string, err error) {
	var ridStr string
	ridStr = fmt.Sprintf("%d", roomId)
	c := pool.Get()
	defer c.Close()
	if channelId, err = redis.String(c.Do("GET", "broadcast:room:room.id:"+ridStr)); err != nil {
		return
	}
	if channelId == "" {
		return "", ErrGetChannelId
	}
	return
}

func BroadcastGetAudienceList(channelId string) (AudienceList []*BroadcastAudience, err error) {
	c := pool.Get()
	defer c.Close()
	var (
		collect [] string
		//AudienceList []*BroadcastAudience
	)
	if collect, err = redis.Strings(c.Do("ZREVRANGE", "broadcast:audience:list.id:"+channelId, 0, Conf.BroadcastAudience-1)); err != nil {
		return
	}
	if len(collect) == 0 {
		return
	}
	log.Debug("audience list guids:%v", collect)
	for _, v := range collect {
		value, err1 := redis.Values(c.Do("HGETALL", "user.id:"+v))
		if err1 != nil {
			log.Error("token %V is wrong HGETALL error: %v", v, err1)
			return
		}
		var user BroadcastAudience
		if err = redis.ScanStruct(value, &user); err != nil {
			log.Error("redis ScanStruct error: %v", err)
			return
		}
		AudienceList = append(AudienceList, &user)
	}
	return
}

func BroadcastInfoAudienceCount(channelId string) (audienceCount int64, err error) {
	c := pool.Get()
	defer c.Close()
	var live_highest_audience int64
	audienceCount, err = redis.Int64(c.Do("HGET", "broadcast:info:id:"+channelId, "audience_count"))
	if err != nil {
		log.Error("BroadcastInfoAudienceCount HGET error channel id %v, audienceCount:%v", channelId, audienceCount)
		return
	}

	live_highest_audience, err = redis.Int64(c.Do("HGET", "broadcast:info:id:"+channelId, "live_highest_audience"))
	if err != nil {
		log.Error("BroadcastInfoAudienceCount HGET error channel id %v, live_highest_audience:%v", channelId, live_highest_audience)
		return
	}
	if audienceCount > live_highest_audience{
		if r, err1 := c.Do("HSET", "broadcast:info:id:"+channelId, "live_highest_audience", live_highest_audience); err1 != nil{
			log.Error("HSET broadcast:info:id:%v is error :%v, result: %v", channelId, err1, r)
		}
	}

	if audienceCount < 0 {
		audienceCount = 0
	}
	return
}

func BroadcastRedisGetInfo(channelId string) (BroadcastInfo proto.BroadcastInfo, err error) {
	c := pool.Get()
	defer c.Close()
	value, err1 := redis.Values(c.Do("HGETALL", "broadcast:info:id:"+channelId))
	if err1 != nil {
		err = err1
		return
	}
	if err = redis.ScanStruct(value, &BroadcastInfo); err != nil {
		log.Error("BroadcastRedisGetInfo redis ScanStruct error: %v", err)
		return
	}
	return
}

type BroadcastAudience struct {
	Nickname string `redis:"nickname" json:"nickname"`
	Guid     string `redis:"guid" json:"guid"`
	Avatar   string `redis:"avatar" json:"avatar"`
	Level    int    `redis:"level" json:"level"`
}

func BroadcastRedisWriteComment(feedId string, commentId string, sendTime int64)(err error)  {
	c := pool.Get()
	defer c.Close()
	_, err = c.Do("ZADD", "feed.comment.id:"+feedId, sendTime,commentId)
	return
}


func BroadcastRedisTimeLineCommentIncrease(feedId string) (err error) {
	c := pool.Get()
	defer c.Close()
	if _, err = c.Do("HINCRBY", "feed.content.id:"+feedId, "comment_count", 1); err != nil {
		return
	}
	return
}

func BroadcastRedisInsertComment(commentId string, msgBody *proto.BroadcastBody, sendTime int64)(err error)  {
	c := pool.Get()
	defer c.Close()
	_, err = c.Do("HMSET", "comment.id:"+commentId, "content", msgBody.Content, "parent_id", "", "time", sendTime, "user_id", msgBody.Guid)
	return
}

func QuizLive() (channelID string, err error) {
	c := pool.Get()
	defer c.Close()
	live, err1 := redis.Strings(c.Do("ZRANGE", "broadcast:quiz.living:", 0, -1))
	if err1 != nil{
		return "", err1
	}
	if len(live) > 0{
		channelID = live[0]
	}
	return

}