package internal

import (
	"C"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"RedBlack-War/conf"
	"time"
)

var (
	session *mgo.Session
)

const (
	dbName          = "REDBLACKWAR-Game" //REDBLACKWAR-Game
	playerID        = "playerID"
	settleWinMoney  = "settleWinMoney"
	settleLoseMoney = "settleLoseMoney"
	surPlusDB       = "surplusPool"
)

// 连接数据库集合的函数 传入集合 默认连接IM数据库
func initMongoDB() {
	// 此处连接正式线上数据库  下面是模拟的直接连接
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{conf.Server.MongoDBAddr},
		Timeout:  60 * time.Second,
		Database: conf.Server.MongoDBAuth,
		Username: conf.Server.MongoDBUser,
		Password: conf.Server.MongoDBPwd,
	}

	var err error
	session, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatal("Connect DataBase 数据库连接失败: %v ", err)
	}
	log.Debug("Connect DataBase 数据库连接成功~")

	//打开数据库
	session.SetMode(mgo.Monotonic, true)

}

func connect(dbName, cName string) (*mgo.Session, *mgo.Collection) {
	s := session.Copy()
	c := s.DB(dbName).C(cName)
	return s, c
}

func InsertPlayerID(ID string) error {
	s, c := connect(dbName, playerID)
	defer s.Close()

	err := c.Insert(ID)
	return err
}

func FindPlayerID(ID string) {
	s, c := connect(dbName, playerID)
	defer s.Close()

	err := c.Find(bson.M{"id": ID})
	if err != nil {
		log.Debug("not Found Player ID")
		err2 := InsertPlayerID(ID)
		if err2 != nil {
			log.Error("<----- 数据库用户ID数据失败 ~ ----->:%v", err)
			return
		}
		SurplusPool -= 6
		log.Debug("<----- 数据库用户ID数据成功 ~ ----->")
	}
}

// 插入房间数据
func InsertWinMoney(base interface{}) {
	s, c := connect(dbName, settleWinMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 数据库插入结算数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入结算数据成功 ~ ----->")

}

// 插入房间数据
func InsertLoseMoney(base interface{}) {
	s, c := connect(dbName, settleLoseMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 数据库插入结算数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入结算数据成功 ~ ----->")

}

func InsertSurplusPool(sur *SurplusPoolDB) {
	s, c := connect(dbName, surPlusDB)
	defer s.Close()

	log.Debug("surplusPoolDB 数据: %v", sur)
	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurplusPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurplusPool数据成功 ~ ----->")
}
