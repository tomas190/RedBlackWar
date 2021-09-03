package internal

import (
	"encoding/json"
	"fmt"
)

const (
	RECODE_CHAOCHUXIANHONG   = 4444
	RECODE_BREATHSTOP        = 1000
	RECODE_PLAYERDESTORY     = 1001
	RECODE_PLAYERBREAKLINE   = 1002
	RECODE_MONEYNOTFULL      = 1003
	RECODE_JOINROOMIDERR     = 1004
	RECODE_PEOPLENOTFULL     = 1005
	RECODE_SELLTENOTDOWNBET  = 1006
	RECODE_NOTDOWNBETSTATUS  = 1007
	RECODE_NOTDOWNMONEY      = 1008
	RECODE_PLAYERHAVESAME    = 1009
	RECODE_DOWNBETMONEYFULL  = 1010
	RECODE_RoomCfgMoneyERROR = 1011
)

var recodeText = map[int32]string{
	RECODE_CHAOCHUXIANHONG:   "4444",
	RECODE_BREATHSTOP:        "用户长时间未响应心跳,停止心跳",
	RECODE_PLAYERDESTORY:     "用户已在其他地方登录",
	RECODE_PLAYERBREAKLINE:   "玩家断开服务器连接,关闭链接",
	RECODE_MONEYNOTFULL:      "玩家金额不足,设为观战",
	RECODE_JOINROOMIDERR:     "请求加入的房间号不正确",
	RECODE_PEOPLENOTFULL:     "房间人数不够,不能开始游戏",
	RECODE_SELLTENOTDOWNBET:  "当前结算阶段,不能进行操作",
	RECODE_NOTDOWNBETSTATUS:  "当前不是下注阶段,玩家不能行动",
	RECODE_NOTDOWNMONEY:      "玩家金额不足,不能进行下注",
	RECODE_PLAYERHAVESAME:    "当前房间已存在相同的用户ID",
	RECODE_DOWNBETMONEYFULL:  "玩家下注金额已限红1-20000",
	RECODE_RoomCfgMoneyERROR: "房间下注配置金额错误",
}

func jsonData() {
	reCode, err := json.Marshal(recodeText)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}

	data := string(reCode)
	fmt.Println("S2C jsonData String ~", data)
}
