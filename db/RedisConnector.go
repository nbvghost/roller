package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nbvghost/roller/entity"
	"github.com/nbvghost/roller/iface"
	"reflect"

	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/nbvghost/glog"
	"github.com/nbvghost/gweb/thread"
)

type keyType struct {
	//Key string
	Table  iface.ISQLModels
	UserID uint64
}
type RedisConnector struct {
	_redisClient *redis.Client
	config       *entity.AppConfig
	/*NotifyMessageChannel string
	NotifyOfflineChannel string
	BroadcastChannel     string
	SyncModelFieldChangeChannel     string*/
	//subscribeMessages chan *redis.Message
	cacheKey  map[string]keyType
	cacheHKey map[string]keyType
	sync.RWMutex
}
type IChannel interface {
	GetMsg() *redis.Message
	Unmarshal()
	Marshal(Msg *redis.Message)
}
type Channel struct {
	Msg     *redis.Message `json:",omitempty"`
	Handler func()         `json:",omitempty"`
}

func (c *Channel) GetMsg() *redis.Message {

	return c.Msg
}

func (c *Channel) Unmarshal() {

}

func (c *Channel) Marshal(Msg *redis.Message) {

}

var NotifyChannel = &Notify{Channel: &Channel{Msg: &redis.Message{Channel: "NotifyChannel"}}}
var BroadcastChannel = &Broadcast{Channel: &Channel{Msg: &redis.Message{Channel: "BroadcastChannel"}}}
var ServiceDiscoverChannel = &ServiceDiscover{Channel: &Channel{Msg: &redis.Message{Channel: "ServiceDiscoverChannel"}}}

//const BroadcastChannel string = "BroadcastChannel"
//const ServiceDiscoverChannel string = "ServiceDiscoverChannel"

//BroadcastChannel
type Broadcast struct {
	*Channel
	UserID   uint64
	Data     string
	WorldKey string
	AreaID   int64
}

//NotifyMessageChannel
type notifyAction = uint64

const NotifyActionMessage notifyAction = 10
const NotifyActionOffline notifyAction = 20

type Notify struct {
	*Channel
	Action notifyAction
}

//service discover
type RedisDiscoverName string

var RedisDiscoverSendEmail RedisDiscoverName = "ServiceDiscoverSendEmail"
var RedisDiscoverForceOffline RedisDiscoverName = "ServiceDiscoverForceOffline"
var RedisDiscoverNotice RedisDiscoverName = "ServiceDiscoverNotice"

//var RedisDiscover RedisDiscoverName = "ServiceDiscoverNotice"

type ServiceDiscover struct {
	*Channel
	IP   string
	Port uint64
	DB   uint64
	Name RedisDiscoverName
}

func (connector *RedisConnector) PubServiceDiscover() *redis.IntCmd {
	ServiceDiscoverChannel.Msg.Channel = "ServiceDiscoverChannel"
	ServiceDiscoverChannel.Msg.Pattern = "*"
	b, _ := json.Marshal(ServiceDiscoverChannel)
	ServiceDiscoverChannel.Msg.Payload = string(b)
	return connector._redisClient.Publish(ServiceDiscoverChannel.Msg.Channel, ServiceDiscoverChannel.Msg)
}
func (connector *RedisConnector) PubNotify() *redis.IntCmd {
	NotifyChannel.Msg.Channel = "NotifyChannel"
	NotifyChannel.Msg.Pattern = "*"
	b, _ := json.Marshal(NotifyChannel)
	ServiceDiscoverChannel.Msg.Payload = string(b)
	return connector._redisClient.Publish(NotifyChannel.Msg.Channel, NotifyChannel.Msg)
}
func (connector *RedisConnector) PubBroadcast() *redis.IntCmd {
	BroadcastChannel.Msg.Channel = "BroadcastChannel"
	BroadcastChannel.Msg.Pattern = "*"
	b, _ := json.Marshal(BroadcastChannel)
	BroadcastChannel.Msg.Payload = string(b)
	return connector._redisClient.Publish(BroadcastChannel.Msg.Channel, BroadcastChannel.Msg)
}
func (connector *RedisConnector) createRedis(config *entity.AppConfig) *RedisConnector {

	connector.cacheKey = make(map[string]keyType)
	connector.cacheHKey = make(map[string]keyType)
	connector.config = config

	connector._redisClient = connector.retryOpenRedisDB()

	newClientChan := make(chan bool)
	thread.NewCoroutine(func() {

		var _NotifyChannel <-chan *redis.Message
		var _BroadcastChannel <-chan *redis.Message
		var _ServiceDiscoverChannel <-chan *redis.Message
		//var syncModelFieldChangeChannel <-chan *redis.Message

		_NotifyChannel = connector._redisClient.Subscribe(NotifyChannel.Msg.Channel).Channel()
		_BroadcastChannel = connector._redisClient.Subscribe(BroadcastChannel.Msg.Channel).Channel()
		_ServiceDiscoverChannel = connector._redisClient.Subscribe(ServiceDiscoverChannel.Msg.Channel).Channel()

		for {

			select {
			case msg := <-_NotifyChannel:

				json.Unmarshal([]byte(msg.Payload), NotifyChannel)
				NotifyChannel.Msg = msg
				if NotifyChannel.Handler != nil {
					NotifyChannel.Handler()
				}
				//connector.subscribeMessages <- msg
			case msg := <-_BroadcastChannel:
				//connector.subscribeMessages <- msg
				json.Unmarshal([]byte(msg.Payload), BroadcastChannel)
				BroadcastChannel.Msg = msg
				if BroadcastChannel.Handler != nil {
					BroadcastChannel.Handler()
				}

			case msg := <-_ServiceDiscoverChannel:
				//connector.subscribeMessages <- msg
				json.Unmarshal([]byte(msg.Payload), ServiceDiscoverChannel)
				ServiceDiscoverChannel.Msg = msg
				if ServiceDiscoverChannel.Handler != nil {
					ServiceDiscoverChannel.Handler()
				}
			case <-newClientChan:
				_NotifyChannel = connector._redisClient.Subscribe(NotifyChannel.Msg.Channel).Channel()
				_BroadcastChannel = connector._redisClient.Subscribe(BroadcastChannel.Msg.Channel).Channel()
				_ServiceDiscoverChannel = connector._redisClient.Subscribe(ServiceDiscoverChannel.Msg.Channel).Channel()

			}
		}

	}, func(v interface{}, stack []byte) {
		glog.Trace(v)
		glog.Trace(string(stack))
	})

	thread.NewCoroutine(func() {
		for {
			//stats, _ := json.Marshal(connector._redisClient.PoolStats())
			//glog.Trace("Redis-Stats：", string(stats))
			if glog.Error(connector._redisClient.Ping().Err()) {

				connector._redisClient = connector.retryOpenRedisDB()
				newClientChan <- true
			}
			time.Sleep(time.Second)
		}

	}, func(v interface{}, stack []byte) {

		glog.Trace(v)
		glog.Trace(string(stack))

	})

	return connector
}

func (connector *RedisConnector) retryOpenRedisDB() *redis.Client {
	_redisClient := redis.NewClient(&redis.Options{
		Addr:     connector.config.RedisDBUrl,
		Password: connector.config.RedisDBPassword, DB: connector.config.RedisDB,
		MinIdleConns: runtime.NumCPU() * 10,
		DialTimeout:  time.Second * 60,
		ReadTimeout:  time.Minute,
	})
	return _redisClient
}

func (connector *RedisConnector) Pipeline() redis.Pipeliner {
	return connector._redisClient.Pipeline()
}
func (connector *RedisConnector) Pipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return connector._redisClient.Pipelined(fn)
}

func (connector *RedisConnector) TxPipeline() redis.Pipeliner {
	return connector._redisClient.TxPipeline()
}
func (connector *RedisConnector) TxPipelined(fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return connector._redisClient.TxPipelined(fn)
}
func (connector *RedisConnector) Persistence(DB *gorm.DB, UserID uint64) error {

	cacheKey := make(map[string]keyType)
	cacheHKey := make(map[string]keyType)

	connector.RLock()

	if UserID == 0 {
		for Key := range connector.cacheKey {
			cacheKey[Key] = connector.cacheKey[Key]
		}
		for Key := range connector.cacheHKey {
			cacheHKey[Key] = connector.cacheHKey[Key]
		}
	} else {

		for Key, value := range connector.cacheKey {
			if value.UserID == UserID {
				cacheKey[Key] = connector.cacheKey[Key]
			}
		}
		for Key, value := range connector.cacheHKey {
			if value.UserID == UserID {
				cacheHKey[Key] = connector.cacheHKey[Key]
			}
		}
	}

	connector.RUnlock()

	if len(connector.cacheKey) > 0 {
		glog.Trace("Redis Keys", "CacheKey", connector.cacheKey)
	}
	if len(connector.cacheHKey) > 0 {
		glog.Trace("Redis Keys", "CacheHKey", connector.cacheHKey)
	}

	//reflect.Copy()
	for Key := range cacheKey {

		tableJson, err := connector.baseGet(string(Key))
		if glog.Error(err) {
			continue
		}

		item := make(map[string]interface{})
		json.Unmarshal([]byte(tableJson), &item)

		delete(item, "ID")
		delete(item, "CreatedAt")
		delete(item, "UpdatedAt")
		err = DB.Model(cacheKey[Key].Table).Updates(item).Error
		if glog.Error(err) == false {
			connector.delCacheKey(Key)
		}

	}

	for Key := range cacheHKey {

		tableJsonMap, err := connector.baseHGetAll(string(Key))
		if glog.Error(err) {
			continue
		}

		for _, value := range tableJsonMap {
			item := make(map[string]interface{})
			json.Unmarshal([]byte(value), &item)

			ID := uint64(item["ID"].(float64))

			delete(item, "ID")
			delete(item, "CreatedAt")
			delete(item, "UpdatedAt")
			//ID, err := strconv.ParseUint(uint64(item["ID"].(float64)), 10, 64)
			//err=DB.Table(value.Table).Where("ID=?",ID).Updates(item).Error
			t := reflect.TypeOf(cacheHKey[Key].Table)
			v := reflect.New(t.Elem()).Elem()
			v.FieldByName("ID").SetUint(ID)
			err = DB.Model(v.Interface()).Updates(item).Error
			glog.Error(err)
		}
		if glog.Error(err) == false {
			connector.delCacheHKey(Key)
		}

	}

	return nil

}
func (connector *RedisConnector) delCacheHKey(key string) {
	connector.Lock()
	defer connector.Unlock()
	//connector.cacheHKey=append(connector.cacheHKey,keyType{Key:key,Table:table})
	delete(connector.cacheHKey, key)
}
func (connector *RedisConnector) delCacheKey(key string) {
	connector.Lock()
	defer connector.Unlock()
	delete(connector.cacheKey, key)
}
func (connector *RedisConnector) addCacheHKey(table iface.ISQLModels, UserID uint64, key string) {
	connector.Lock()
	defer connector.Unlock()
	//connector.cacheHKey=append(connector.cacheHKey,keyType{Key:key,Table:table})
	connector.cacheHKey[key] = keyType{Table: table, UserID: UserID}
}
func (connector *RedisConnector) addCacheKey(table iface.ISQLModels, UserID uint64, key string) {
	connector.Lock()
	defer connector.Unlock()
	connector.cacheKey[key] = keyType{Table: table, UserID: UserID}
}
func (connector *RedisConnector) baseGet(key string) (string, error) {
	stringCmd := connector._redisClient.Get(key)
	return stringCmd.Result()
}

func (connector *RedisConnector) createRedisKey(table iface.ISQLModels, UserID uint64) string {
	key := "FX:Cache:" + table.TableName() + ":" + strconv.FormatUint(UserID, 10)
	return key
}
func (connector *RedisConnector) baseHGetAll(key string) (map[string]string, error) {
	stringStringMapCmd := connector._redisClient.HGetAll(key)
	return stringStringMapCmd.Result()
}
func (connector *RedisConnector) HGetAll(table iface.ISQLModels, UserID uint64) (map[string]string, error) {
	key := connector.createRedisKey(table, UserID)
	connector.addCacheHKey(table, UserID, key)
	stringStringMapCmd := connector._redisClient.HGetAll(key)
	return stringStringMapCmd.Result()
}
func (connector *RedisConnector) HSet(UserID uint64, Field string, Value iface.ISQLModels) error {
	b, err := json.Marshal(Value)
	glog.Error(err)
	key := connector.createRedisKey(Value, UserID)

	connector.addCacheHKey(Value, UserID, key)

	//boolCmd := connector._redisClient.HSet(key, strconv.FormatUint(Value.GetID(), 10), string(b))
	boolCmd := connector._redisClient.HSet(key, Field, string(b))
	connector._redisClient.Expire(key, time.Duration(connector.config.SessionExpires)*time.Second)
	return boolCmd.Err()
}
func (connector *RedisConnector) HGet(UserID uint64, Field string, SetValue iface.ISQLModels) (err error) {
	key := connector.createRedisKey(SetValue, UserID)

	exists := connector._redisClient.HExists(key, Field)
	if exists.Val() == false {
		return errors.New(fmt.Sprintf("redis no found key:%v", key))
	}
	stringCmd := connector._redisClient.HGet(key, Field)
	err = stringCmd.Err()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stringCmd.Val()), SetValue)
	if err != nil {
		return err
	}
	connector.addCacheHKey(SetValue, UserID, key)
	return connector._redisClient.Expire(key, time.Duration(connector.config.SessionExpires)*time.Second).Err()
}
func (connector *RedisConnector) HDel(table iface.ISQLModels, UserID uint64, Fields ...string) int64 {
	intCmd := connector._redisClient.HDel(connector.createRedisKey(table, UserID), Fields...)
	return intCmd.Val()
}
func (connector *RedisConnector) Set(UserID uint64, Value iface.ISQLModels) error {
	key := connector.createRedisKey(Value, UserID)

	connector.addCacheKey(Value, UserID, key)

	b, err := json.Marshal(Value)
	if glog.Error(err) {
		return err
	}
	boolCmd := connector._redisClient.Set(key, string(b), time.Duration(connector.config.SessionExpires)*time.Second)

	//connector._redisClient.Expire(key, time.Duration(connector.config.SessionExpires)*time.Second)

	return boolCmd.Err()
}
func (connector *RedisConnector) Get(UserID uint64, SetValue iface.ISQLModels) (err error) {
	key := connector.createRedisKey(SetValue, UserID)

	if connector._redisClient.Exists(key).Val() == 0 {
		return errors.New(fmt.Sprintf("redis no found key:%v", key))
	}
	stringCmd := connector._redisClient.Get(key)
	err = stringCmd.Err()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(stringCmd.Val()), SetValue)
	if err != nil {
		return err
	}
	connector.addCacheKey(SetValue, UserID, key)
	connector._redisClient.Expire(key, time.Duration(connector.config.SessionExpires)*time.Second)
	return err
}
func (connector *RedisConnector) Del(table iface.ISQLModels, UserID uint64) int64 {

	intCmd := connector._redisClient.Del(connector.createRedisKey(table, UserID))
	return intCmd.Val()

}
