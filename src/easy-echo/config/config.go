package config

import (
	"easy-echo/logger"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"
)

type Config struct {
	Host        string `json:"host"`
	LogDir      string `json:"log_dir"`
	LogFileType string `json:"log_file_type"`
	DB          struct {
		Mongo      string `json:"mongo"`
		DbUser     string `json:"db_user"`
		DbPwd      string `json:"db_pwd"`
		DbAuthName string `json:"db_auth_name"`
		DbBusName  string `json:"db_bus_name"`
	} `json:"db"`
	Cache struct {
		MasterRedisHost         string `json:"master_redis_host"`
		MasterRedisAuth         string `json:"master_redis_auth"`
		SlaveRedisHost          string `json:"slave_redis_host"`
		SlaveRedisAuth          string `json:"slave_redis_auth"`
		RedisSentinelHosts      string `json:"redis_sentinel_hosts"`
		RedisSentinelMasterName string `json:"redis_sentinel_master_name"`
		RedisSentinelAuth       string `json:"redis_sentinel_auth"`
	} `json:"cache"`
}

var (
	Cfg     Config
	cfgLock sync.Mutex
)

var (
	// Version git version
	Version string
	// Build app build time
	Build string
)

func InitConfig(cfg *Config) (err error) {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	rand.Seed(time.Now().Unix())
	confile := flag.String("config", "etc/config.json", "config file")
	showbuild := flag.Bool("version", false, "version")
	flag.Parse()

	if *showbuild {
		fmt.Printf("Version: %s\nBuild: %s\n", Version, Build)
		return
	}
	cfg, err = loadConfig(*confile)
	if err != nil {
		glog.Fatalf("load config fail %s", err.Error())
		return
	}

	Cfg = *cfg

	// set glog log dir
	_ = flag.Lookup("log_dir").Value.Set(Cfg.LogDir)
	logger.Init(Cfg.LogDir, Cfg.LogFileType)

	defer glog.Flush()
	return
}

func loadConfig(configFile string) (*Config, error) {
	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read config file failed:%s", err.Error())
	}
	conf := new(Config)
	if err = json.Unmarshal(raw, conf); err != nil {
		return nil, fmt.Errorf("json parse failed:%s", err.Error())
	}
	return conf, nil
}
