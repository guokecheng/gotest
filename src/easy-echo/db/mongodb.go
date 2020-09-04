package db

import (
	"easy-echo/config"
	"fmt"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
)

var (
	MgoSession *mgo.Session
	mgoLock    sync.Mutex
)

func InitMongoDB() (err error){
	db, err := NewDB()
	if err != nil {
		return
	}
	defer db.Session.Close()
	return
}

func NewSession() (*mgo.Session, error) {

	if MgoSession == nil {
		mgoLock.Lock()
		defer mgoLock.Unlock()

		addrs := strings.Split(config.Cfg.DB.Mongo, ",")
		timeout := time.Duration(60) * time.Second

		dbUser := config.Cfg.DB.DbUser
		dbPwd := config.Cfg.DB.DbPwd

		dialInfo := &mgo.DialInfo{
			Addrs:    addrs,
			Timeout:  timeout,
			Password: dbPwd,
			Username: dbUser,
		}

		if MgoSession == nil {
			var err error
			MgoSession, err = mgo.DialWithInfo(dialInfo)
			if err != nil {
				fmt.Printf("[NewSession] %v", err.Error())
				return nil, err
			}
			MgoSession.SetMode(mgo.Monotonic, true)
		}
	}

	return MgoSession.Clone(), nil
}

func NewDB() (*mgo.Database, error) {
	session, err := NewSession()
	if err != nil {
		return nil, err
	}
	return session.DB(config.Cfg.DB.DbBusName), nil
}
