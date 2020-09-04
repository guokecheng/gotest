package cache

import (
	"easy-echo/config"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/mojocn/base64Captcha"
)

var (
	masterRds *redis.Client
	slaveRds  *redis.Client
	redisLock    sync.Mutex
)

//customizeRdsStore An object implementing Store interface
type customizeRdsStore struct {
	redisClient *redis.Client
}

// customizeRdsStore implementing Set method of  Store interface
func (s *customizeRdsStore) Set(id string, value string) {
	err := s.redisClient.Set(id, value, time.Minute*10).Err()
	if err != nil {
		panic(err)
	}
}

// customizeRdsStore implementing Get method of  Store interface
func (s *customizeRdsStore) Get(id string, clear bool) (value string) {
	val, err := s.redisClient.Get(id).Result()
	if err != nil {
		panic(err)
	}
	if clear {
		err := s.redisClient.Del(id).Err()
		if err != nil {
			panic(err)
		}
	}
	return val
}

func InitRedis() (err error) {

	redisLock.Lock()
	defer redisLock.Unlock()

	redisSentinelHosts := config.Cfg.Cache.RedisSentinelHosts
	redisSentinelMasterName := config.Cfg.Cache.RedisSentinelMasterName
	redisSentinelAuth := config.Cfg.Cache.RedisSentinelAuth
	masterRedisHost := config.Cfg.Cache.MasterRedisHost
	masterRedisAuth := config.Cfg.Cache.MasterRedisAuth
	slaveRedisHost := config.Cfg.Cache.SlaveRedisHost
	slaveRedisAuth := config.Cfg.Cache.SlaveRedisAuth

	if redisSentinelHosts == "" && (masterRedisHost == "" || slaveRedisHost == "") {
		err = errors.New("config MasterRedisHost or SlaveRedisHost not found")
		return
	}

	if redisSentinelHosts != "" {

		masterRds = redis.NewFailoverClient(&redis.FailoverOptions {
			MasterName:    redisSentinelMasterName,
			SentinelAddrs: strings.Split(redisSentinelHosts, ","),
			Password:      redisSentinelAuth,
		})

		if err = masterRds.Ping().Err(); err != nil {
			return
		}

		slaveRds = redis.NewFailoverClient(&redis.FailoverOptions {
			MasterName:    redisSentinelMasterName,
			SentinelAddrs: strings.Split(redisSentinelHosts, ","),
			Password:      redisSentinelAuth,
		})

		if err = slaveRds.Ping().Err(); err != nil {
			return
		}

	}else {

		masterRds = redis.NewClient(&redis.Options{
			Addr:       masterRedisHost,
			Password:   masterRedisAuth,
			MaxRetries: 3,
		})
		if _, err = masterRds.Ping().Result(); err != nil {
			return
		}
		slaveRds = redis.NewClient(&redis.Options{
			Addr:       slaveRedisHost,
			Password:   slaveRedisAuth,
			MaxRetries: 3,
		})
		if _, err = slaveRds.Ping().Result(); err != nil {
			return
		}
	}

	customStore := customizeRdsStore{masterRds}
	base64Captcha.SetCustomStore(&customStore)

	return
}
