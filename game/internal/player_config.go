package internal

import (
	"github.com/name5566/leaf/gate"
)

//玩家状态 (玩家金额小于50视为观战，玩家中途加入结算阶段，视为观战)
//(观战状态针对于房间内部观战玩家禁止下注。)
type PlayerStatus int32

const (
	PlayGame  PlayerStatus = 1 //游戏状态
	WatchGame PlayerStatus = 2 //观战状态
)

type GameStates int32

const (
	InGameHall GameStates = 1 //在大厅状态
	InGameRoom GameStates = 2 //在房间状态
)

const (
	GamesNumLimit = 72 //玩家获取房间游戏数据上限
)

//定义全局用户房间
var userRoomMap map[string]*Room

//玩家各注池下注的金额
type DownBetMoney struct {
	RedDownBet   int32
	BlackDownBet int32
	LuckDownBet  int32
}

//玩家各注池是否下了注
type DownPotType struct {
	RedDownPot   bool
	BlackDownPot bool
	LuckDownPot  bool
}

//续投金额
type ContinueBet struct {
	DownBetMoneys *DownBetMoney
	TotalMoneyBet int32
}

type HallDataList struct {
	Rid              string  //房间ID
	HallCardTypeList []int32 //玩家大厅停留列表
	HallRedBlackList []int32 //玩家大厅停留列表
}

type GameDataList struct {
	ResultCount int32 // 0为Lose，1为Win
	ResultMoney float64
}

type Player struct {
	// 玩家代理链接
	ConnAgent gate.Agent
	// 客户端延迟
	uClientDelay int32

	Index uint32 // 下一版这里改动一些，将Player的map[Index]去掉

	Id       string
	RoundId  string
	PassWord string
	Token    string
	NickName string
	HeadImg  string
	Account  float64 // 玩家金额

	room          *Room         //玩家房间信息
	DownBetMoneys *DownBetMoney //玩家各注池下注的金额
	Status        PlayerStatus  //玩家状态
	IsAction      bool          //玩家是否行动了
	ContinueVot   *ContinueBet  //续投，记录玩家上局的下注金额。
	GameState     GameStates    //大厅和房间状态

	TaxPreMoney     float64 //税前Win的总金额
	ResultMoney     float64 //本局游戏结束的钱,Win and Lose
	WinResultMoney  float64 //本局Win的Money
	LoseResultMoney float64 //本局Lose的Money

	TotalAmountBet int32          //玩家房间总下注金额
	WinTotalCount  int32          //玩家房间获胜Win总次数
	PotWinList     []*GameWinList //底池每局Win总列表
	CardTypeList   []int32        //卡牌类型列表
	RedBlackList   []int32        //每局红黑Win总顺序列表  红为 1,黑为 2
	RedWinCount    int32          //Win总列表红Win的局数
	BlackWinCount  int32          //Win总列表黑Win的局数
	LuckWinCount   int32          //Win总列表幸运的局数
	TotalCount     int32          //总列表总局数

	HallRoomData []*HallDataList //玩家大厅房间数据
	TwentyData   []*GameDataList //20局玩家数据截取

	IsOnline bool //玩家是否在线
	//SeatNum        int32          //玩家座位号

	NotOnline int32

	// 是否是机器人
	IsRobot bool
}
