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
	robotData       = "robotData"
	StatementDB     = "StatementDB"
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

	// 建立index索引
	createUniqueIndex(surPlusDB, []string{"updatetime"})
	createUniqueIndex(accessDB, []string{"id"})
	createUniqueIndex(StatementDB, []string{"id"})
}

func createUniqueIndex(cName string, keys []string) {
	s, c := connect(dbName, cName)
	defer s.Close()

	// 設定統計表唯一索引
	index := mgo.Index{
		Key:        keys,  // 索引鍵
		Background: true,  // 不长时间占用写锁
		Unique:     false, // 唯一索引(在表中會不會有重複資料)
		DropDups:   true,  // 存在資料後建立, 則自動刪除重複資料
	}

	err := c.EnsureIndex(index)
	if err != nil {
		log.Debug("mongodb建立index失败:%v", err)
	} else {
		log.Debug("mongodb建立index:%v,%v", cName, index)
	}
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

//InsertWinMoney 插入赢钱数据
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

//InsertLoseMoney 插入输钱数据
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

	SurplusPool = sur.PoolMoney
	log.Debug("surplusPoolDB 数据: %v", sur.PoolMoney)

	err := c.Insert(sur)
	if err != nil {
		log.Error("<----- 数据库插入SurplusPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 数据库插入SurplusPool数据成功 ~ ----->")

	SurPool := &SurPool{}
	SurPool.GameId = conf.Server.GameID
	SurPool.SurplusPool = sur.PoolMoney
	SurPool.PlayerTotalLoseWin = sur.HistoryLose - sur.HistoryWin
	SurPool.PlayerTotalLose = sur.HistoryLose
	SurPool.PlayerTotalWin = sur.HistoryWin
	SurPool.TotalPlayer = sur.PlayerNum
	SurPool.FinalPercentage = 0.5
	SurPool.PercentageToTotalWin = 1
	SurPool.CoefficientToTotalPlayer = sur.PlayerNum * 0
	SurPool.PlayerLoseRateAfterSurplusPool = 0.7
	SurPool.DataCorrection = 0
	SurPool.PlayerWinRate = 0.6
	SurPool.RandomCountAfterWin = 0
	SurPool.RandomCountAfterLose = 0
	SurPool.RandomPercentageAfterWin = 0.6
	SurPool.RandomPercentageAfterLose = 0.6
	FindSurPool(SurPool)
}

type SurPool struct {
	GameId                         string  `json:"game_id" bson:"game_id"`
	PlayerTotalLose                float64 `json:"player_total_lose" bson:"player_total_lose"`
	PlayerTotalWin                 float64 `json:"player_total_win" bson:"player_total_win"`
	PercentageToTotalWin           float64 `json:"percentage_to_total_win" bson:"percentage_to_total_win"`
	TotalPlayer                    int32   `json:"total_player" bson:"total_player"`
	CoefficientToTotalPlayer       int32   `json:"coefficient_to_total_player" bson:"coefficient_to_total_player"`
	FinalPercentage                float64 `json:"final_percentage" bson:"final_percentage"`
	PlayerTotalLoseWin             float64 `json:"player_total_lose_win" bson:"player_total_lose_win" `
	SurplusPool                    float64 `json:"surplus_pool" bson:"surplus_pool"`
	PlayerLoseRateAfterSurplusPool float64 `json:"player_lose_rate_after_surplus_pool" bson:"player_lose_rate_after_surplus_pool"`
	DataCorrection                 float64 `json:"data_correction" bson:"data_correction"`
	PlayerWinRate                  float64 `json:"player_win_rate" bson:"player_win_rate"`
	RandomCountAfterWin            float64 `json:"random_count_after_win" bson:"random_count_after_win"`
	RandomCountAfterLose           float64 `json:"random_count_after_lose" bson:"random_count_after_lose"`
	RandomPercentageAfterWin       float64 `json:"random_percentage_after_win" bson:"random_percentage_after_win"`
	RandomPercentageAfterLose      float64 `json:"random_percentage_after_lose" bson:"random_percentage_after_lose"`
}

func FindSurPool(SurP *SurPool) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	sur := &SurPool{}
	err := c.Find(nil).One(sur)
	if err != nil {
		InsertSurPool(SurP)
	} else {
		SurP.SurplusPool = (SurP.PlayerTotalLose - (SurP.PlayerTotalWin * sur.PercentageToTotalWin) - float64(SurP.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage
		SurP.FinalPercentage = sur.FinalPercentage
		SurP.PercentageToTotalWin = sur.PercentageToTotalWin
		SurP.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
		SurP.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
		SurP.DataCorrection = sur.DataCorrection
		SurP.PlayerWinRate = sur.PlayerWinRate
		SurP.RandomCountAfterWin = sur.RandomCountAfterWin
		SurP.RandomCountAfterLose = sur.RandomCountAfterLose
		SurP.RandomPercentageAfterWin = sur.RandomPercentageAfterWin
		SurP.RandomPercentageAfterLose = sur.RandomPercentageAfterLose
		UpdateSurPool(SurP)
	}
}

func GetFindSurPool() *SurPool {
	s, c := connect(dbName, surPool)
	defer s.Close()

	sur := &SurPool{}
	err := c.Find(nil).One(sur)
	if err != nil {
		log.Debug("获取GetFindSurPool数据为空:%v", err)
		sur.GameId = conf.Server.GameID
		sur.TotalPlayer = FindIdCount()
		sur.FinalPercentage = 0.5
		sur.PercentageToTotalWin = 1
		sur.CoefficientToTotalPlayer = sur.TotalPlayer * 0
		sur.PlayerLoseRateAfterSurplusPool = 0.7
		sur.DataCorrection = 0
		sur.PlayerWinRate = 0.6
		sur.RandomCountAfterWin = 0
		sur.RandomCountAfterLose = 0
		sur.RandomPercentageAfterWin = 0.6
		sur.RandomPercentageAfterLose = 0.6
		InsertSurPool(sur)
		return sur
	}
	return sur
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

	err := c.Update(bson.M{}, sur)
	if err != nil {
		log.Error("<----- 更新 SurPool数据失败 ~ ----->:%v", err)
		return
	}
	log.Debug("<----- 更新SurPool数据成功 ~ ----->")
}

// 玩家的记录
type PlayerDownBetRecode struct {
	Id              string        `json:"id" bson:"id"`                             // 玩家Id
	GameId          string        `json:"game_id" bson:"game_id"`                   // gameId
	RoundId         string        `json:"round_id" bson:"round_id"`                 // 随机Id
	RoomId          string        `json:"room_id" bson:"room_id"`                   // 所在房间
	DownBetInfo     *DownBetMoney `json:"down_bet_info" bson:"down_bet_info"`       // 玩家各注池下注的金额
	DownBetTime     int64         `json:"down_bet_time" bson:"down_bet_time"`       // 下注时间
	StartTime       int64         `json:"start_time" bson:"start_time"`             // 开始时间
	EndTime         int64         `json:"end_time" bson:"end_time"`                 // 结束时间
	CardResult      *CardData     `json:"card_result" bson:"card_result"`           // 当局开牌结果
	SettlementFunds float64       `json:"settlement_funds" bson:"settlement_funds"` // 当局输赢结果(税后)
	SpareCash       float64       `json:"spare_cash" bson:"spare_cash"`             // 剩余金额
	TaxRate         float64       `json:"tax_rate" bson:"tax_rate"`                 // 税率
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
func GetDownRecodeList(page, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
	s, c := connect(dbName, accessDB)
	defer s.Close()

	var wts []PlayerDownBetRecode

	n, err := c.Find(selector).Count()
	if err != nil {
		return nil, 0, err
	}
	log.Debug("获取 %v 条数据,limit:%v", n, limit)
	skip := (page - 1) * limit
	err = c.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
	if err != nil {
		return nil, 0, err
	}
	return wts, n, nil
}

//GetDownRecodeList 获取盈余池数据
func GetSurPoolData(selector bson.M) (SurPool, error) {
	s, c := connect(dbName, surPool)
	defer s.Close()

	var wts SurPool

	err := c.Find(selector).One(&wts)
	if err != nil {
		return wts, err
	}
	return wts, nil
}

type ChipDownBet struct {
	Chip1    int32
	Chip10   int32
	Chip50   int32
	Chip100  int32
	Chip1000 int32
}

type RobotDATA struct {
	RoomId   string       `json:"room_id" bson:"room_id"`
	RoomTime int64        `json:"room_time" bson:"room_time"`
	RobotNum int          `json:"robot_num" bson:"robot_num"`
	RedPot   *ChipDownBet `json:"red_pot" bson:"red_pot"`
	BlackPot *ChipDownBet `json:"black_pot" bson:"black_pot"`
	LuckPot  *ChipDownBet `json:"luck_pot" bson:"luck_pot"`
}

func InsertRobotData(rb *RobotDATA) {
	s, c := connect(dbName, robotData)
	defer s.Close()

	err := c.Insert(rb)
	if err != nil {
		log.Debug("插入机器人数据失败:%v", err)
		return
	}
	//log.Debug("插入机器人数据成功~")
}

//GetDownRecodeList 获取盈余池数据
func GetRobotData() ([]RobotDATA, error) {
	s, c := connect(dbName, robotData)
	defer s.Close()

	var rd []RobotDATA

	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location()).Unix()
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location()).Unix()

	selector := bson.M{}
	selector["room_time"] = bson.M{"$gte": startTime, "$lte": endTime}

	err := c.Find(selector).All(&rd)

	if err != nil {
		log.Debug("获取机器人数据失败:%v", err)
		return nil, err
	}
	log.Debug("获取机器人数据成功:%v,长度为:%v", rd, len(rd))
	return rd, nil
}

type StatementData struct {
	Id                 string  `json:"id" bson:"id"`
	GameId             string  `json:"game_id" bson:"game_id"`
	GameName           string  `json:"game_name" bson:"game_name"`
	StartTime          int64   `json:"start_time" bson:"start_time"`
	EndTime            int64   `json:"end_time" bson:"end_time"`
	DownBetTime        int64   `json:"down_bet_time" bson:"down_bet_time"`
	PackageId          uint16  `json:"package_id" bson:"package_id"`
	WinStatementTotal  float64 `json:"win_statement_total" bson:"win_statement_total"`
	LoseStatementTotal float64 `json:"lose_statement_total" bson:"lose_statement_total"`
	BetMoney           int32   `json:"bet_money" bson:"bet_money"`
}

func InsertStatementDB(sd *StatementData) {
	s, c := connect(dbName, StatementDB)
	defer s.Close()

	err := c.Insert(sd)
	if err != nil {
		log.Debug("插入游戏统计数据失败:%v", err)
		return
	}
	log.Debug("插入游戏统计数据成功~")
}

func GetStatementList(selector bson.M) ([]StatementData, error) {
	s, c := connect(dbName, StatementDB)
	defer s.Close()

	var wts []StatementData

	err := c.Find(selector).All(&wts)
	if err != nil {
		return nil, err
	}
	return wts, nil
}
