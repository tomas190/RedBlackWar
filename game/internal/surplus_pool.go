package internal

//盈余池数据存入数据库
type SurplusPoolDB struct {
	TimeNow        string  //记录时间（分为时间戳/字符串显示）
	Rid            string  //房间ID
	TotalWinMoney  float64 //玩家当局总赢
	TotalLoseMoney float64 //玩家当局总输
	PoolMoney      float64 //盈余池
	HistoryWin     float64 //玩家历史总赢
	HistoryLose    float64 //玩家历史总输
	PlayerNum      int32   //历史玩家人数
}

const (
	taxRate    float64 = 0.05 //税率
	SurplusTax float64 = 0.2  //指定盈余池的百分随机数
)

//盈余池
var SurplusPool float64 = 0

//记录进入大厅玩家的数量,为了统计 盈余池 - 6
var AllPlayerCount []string

var (
	AllHistoryWin  float64 = 0
	AllHistoryLose float64 = 0
)

//返回记录的玩家总数量
func RecordPlayerCount() int32 {
	//log.Debug("游戏玩过总人数数量: %v", int32(len(AllPlayerCount)))
	return int32(len(AllPlayerCount))
}
