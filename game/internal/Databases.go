package internal

import (
	"C"
	"RedBlack-War/conf"
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	accessDB        = "accessData"
	surPool         = "surplus-pool"
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

func InsertPlayerID(p *pb_msg.PlayerInfo) error {
	s, c := connect(dbName, playerID)
	defer s.Close()

	err := c.Insert(p)
	return err
}

func FindPlayerID(p *Player) {
	s, c := connect(dbName, playerID)
	defer s.Close()

	player := &pb_msg.PlayerInfo{}
	player.Id = p.Id
	player.NickName = p.NickName
	player.HeadImg = p.HeadImg
	player.Account = p.Account

	err := c.Find(bson.M{"id": player.Id}).One(player)
	if err != nil {
		log.Debug("not Found Player ID")
		err2 := InsertPlayerID(player)
		if err2 != nil {
			log.Error("<----- 数据库用户ID数据失败 ~ ----->:%v", err)
			return
		}
		log.Debug("<----- 数据库用户ID数据成功 ~ ----->")
	}
}

func FindIdCount() int32 {
	s, c := connect(dbName, playerID)
	defer s.Close()

	n, err := c.Find(nil).Count()
	if err != nil {
		log.Debug("not Found Player ID")
	}
	return int32(n)
}

//InsertWinMoney 插入房间数据
func InsertWinMoney(base interface{}) {
	s, c := connect(dbName, settleWinMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 赢钱结算数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 赢钱结算数据插入成功 ~ ----->")

}

//InsertLoseMoney 插入房间数据
func InsertLoseMoney(base interface{}) {
	s, c := connect(dbName, settleLoseMoney)
	defer s.Close()

	err := c.Insert(base)
	if err != nil {
		log.Error("<----- 输钱结算数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 输钱结算数据插入成功 ~ ----->")
}

//FindSurplusPool
func FindSurplusPool() *SurplusPoolDB {
	s, c := connect(dbName, surPlusDB)
	defer s.Close()

	sur := &SurplusPoolDB{}
	err := c.Find(nil).Sort("-updatetime").One(sur)
	if err != nil {
		log.Error("<----- 查找SurplusPool数据失败 ~ ----->:%v", err)
		return nil
	}

	return sur
}

//InsertSurplusPool 插入盈余池数据
func InsertSurplusPool(sur *SurplusPoolDB) {
	s, c := connect(dbName, surPlusDB)
	defer s.Close()

	sur.PoolMoney = (sur.HistoryLose - (sur.HistoryWin * 1)) * 0.5
	SurplusPool = sur.PoolMoney
	log.Debug("surplusPoolDB 数据: %v", sur.PoolMoney)

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurplusPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurplusPool数据成功 ~ ----->")

	SurPool := &SurPool{}
	SurPool.SurplusPool = sur.PoolMoney
	SurPool.PlayerTotalLoseWin = sur.HistoryLose - sur.HistoryWin
	SurPool.PlayerTotalLose = sur.HistoryLose
	SurPool.PlayerTotalWin = sur.HistoryWin
	SurPool.TotalPlayer = sur.PlayerNum
	SurPool.FinalPercentage = 0.5
	SurPool.PercentageToTotalWin = 1
	SurPool.CoefficientToTotalPlayer = sur.PlayerNum * 0
	SurPool.PlayerLoseRateAfterSurplusPool = 0.7

	n, _ := c.Find(nil).Count()
	if n == 0 {
		InsertSurPool(SurPool)
	} else {
		UpdateSurPool(SurPool)
	}
}

type SurPool struct {
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
}

//插入盈余池统一字段
func InsertSurPool(sur *SurPool) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	log.Debug("SurPool 数据: %v", sur)

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurPool数据成功 ~ ----->")
}

func UpdateSurPool(sur *SurPool) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	err := c.Update(&SurPool{}, sur)
	if err != nil {
		log.Error("<----- 更新 SurPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 更新SurPool数据成功 ~ ----->")
}

// 玩家的记录
type PlayerDownBetRecode struct {
	Id          string        `json:"id" bson:"id"`                       // 玩家Id
	RandId      string        `json:"rand_id" bson:"rand_id"`             // 随机Id
	RoomId      string        `json:"room_id" bson:"room_id"`             // 所在房间
	DownBetInfo *DownBetMoney `json:"down_bet_info" bson:"down_bet_info"` // 玩家各注池下注的金额
	DownBetTime int64         `json:"down_bet_time" bson:"down_bet_time"` // 下注时间
	CardResult  *CardData     `json:"card_result" bson:"card_result"`     // 当局开牌结果
	ResultMoney float64       `json:"result_money" bson:"result_money"`   // 当局输赢结果(税后)
	TaxRate     float64       `json:"tax_rate" bson:"tax_rate"`           // 税率
	//ViewInfo    *ViewInfo     `json:"view_info" bson:"view_info"`         // 返回給前端的显示信息
}

//InsertAccessData 插入运营数据接入
func InsertAccessData(data *PlayerDownBetRecode) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	log.Debug("AccessData 数据: %v", data)
	err := c.Insert(data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 运营接入数据插入成功 ~ ----->")
}

//GetDownRecodeList 获取运营数据接入
func GetDownRecodeList(skip, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	var wts []PlayerDownBetRecode

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	err = c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}
