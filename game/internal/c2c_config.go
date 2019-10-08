package internal

const (
	msgServerLogin   string = "/GameServer/Login/login"
	msgUserLogin     string = "/GameServer/GameUser/login"
	msgUserLogout    string = "/GameServer/GameUser/loginout"
	msgUserWinScore  string = "/GameServer/GameUser/winSettlement"
	msgUserLoseScore string = "/GameServer/GameUser/loseSettlement"
)

//BaseMessage 基本消息结构
type BaseMessage struct {
	Event string      `json:"event"` // 事件
	Data  interface{} `json:"data"`  // 数据
}

//ServerLogin 服务器登录
type ServerLogin struct {
	Host    string `json:"host"`    // 主机
	Port    string `json:"port"`    // 端口
	GameId  string `json:"game_id"` // 游戏Id
	DevName string `json:"dev_name"`
	DevKey  string `json:"dev_key"`
}

//UserReq 用户请求，用登录登出
type UserReq struct {
	ID       string `json:"id"`
	GameId   string `json:"game_id"`
	PassWord string `json:"password"`
	DevName  string `json:"dev_name"`
	DevKey   string `json:"dev_key"`
}

//ServerLoginRsp 服务器登录返回
type ServerLoginRsp struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
}

//UserAuth 用户认证数据
type UserAuth struct {
	DevName string `json:"dev_name"`
	DevKey  string `json:"dev_key"`
}

//UserScoreSync 同步分值数据
type UserScoreSync struct {
	ID         string  `json:"id"`
	CreateTime int64   `json:"create_time"`
	PayReason  string  `json:"pay_reason"`
	Money      float64 `json:"money"`
	LockMoney  float64 `json:"lock_money"`
	PreMoney   float64 `json:"pre_money"`
	Order      string  `json:"order"` //唯一ID,方便之后查询
	GameId     string  `json:"game_id"`
	RoundId    string  `json:"round_id"` //唯一ID，识别多人是否在同一局游戏
}

//UserChangeScore 用户分值改变
type UserChangeScore struct {
	Auth UserAuth      `json:"auth"`
	Info UserScoreSync `json:"info"`
}

//UserInfo 用户信息
type UserInfo struct {
	ID      string
	Nick    string
	HeadImg string
	Score   float64
}

//UserCallback 用户登录回调函数保存
type UserCallback struct {
	Data     UserInfo
	Callback func(data *UserInfo)
}
