package internal

import "time"

type RoomStatus int32

const (
	RoomStatusNone RoomStatus = 1 // 房间等待状态
	RoomStatusRun  RoomStatus = 2 // 房间运行状态
	RoomStatusOver RoomStatus = 3 // 房间结束状态
)

//房间状态 (分为下注阶段、比牌结算阶段)
//主要是针对于玩家中途加入房间，如果是下注阶段，玩家可直接进行下注。比牌结算阶段，玩家则视为观战
type GameStatus int32

const (
	DownBet GameStatus = 1 //下注阶段
	Settle  GameStatus = 2 //比牌结算阶段
)

const (
	DownBetTime = 15 //下注阶段时间 15秒
	SettleTime  = 10 //比牌结算阶段时间 10秒
)

const (
	RoomCordCount  = 40 //玩家进入房间获取房间的战绩数量。
	RoomLimitMoney = 50 //房间限定金额50,否则处于观战状态
)


//游戏状态channel
var DownBetChannel chan bool
var RobotDownBetChan chan bool

type GameWinList struct {
	RedWin    int32     //红Win为 1
	BlackWin  int32     //黑Win为 1
	LuckWin   int32     //幸运luck为 1
	CardTypes CardsType //比牌类型  1 单张,2 对子,3 顺子,4 金花,5 顺金,6 豹子
}

//房间注池下注总金额
type PotRoomCount struct {
	RedMoneyCount   int32 //红池金额数量
	BlackMoneyCount int32 //黑池金额数量
	LuckMoneyCount  int32 //Luck金额数量
}

//卡牌数据
type CardData struct {
	ReadCard  []int32
	BlackCard []int32
	RedType   CardsType
	BlackType CardsType
	LuckType  CardsType // 本局幸运类型
}

type Room struct {
	RoomId     string    //房间号
	PlayerList []*Player //玩家列表

	GodGambleName string     //赌神id
	RoomStat      RoomStatus //房间状态
	GameStat      GameStatus //游戏状态

	Cards          *CardData      //卡牌数据
	PotMoneyCount  *PotRoomCount  //房间注池下注总金额
	CardTypeList   []int32        //卡牌类型的总集合 1 单张,2 对子,3 顺子,4 金花,5 顺金,6 豹子
	RPotWinList    []*GameWinList //红黑Win、Luck、比牌类型的总集合
	GameTotalCount int32          //房间游戏的总局数

	counter int32        //已经过去多少秒
	clock   *time.Ticker //计时器
	//是否加载机器人
	IsLoadRobots bool
}
