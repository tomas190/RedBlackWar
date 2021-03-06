package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/log"
	"time"
)

func (p *Player) Init() {
	p.ConnAgent = nil
	p.uClientDelay = 0
	p.Index = 0

	p.TotalAmountBet = 0
	p.IsAction = false
	p.DownBetMoneys = new(DownBetMoney)
	p.ContinueVot = new(ContinueBet)
	p.ContinueVot.DownBetMoneys = new(DownBetMoney)
	p.GameState = InGameHall

	p.TaxPreMoney = 0
	p.ResultMoney = 0
	p.WinResultMoney = 0
	p.LoseResultMoney = 0

	p.room = new(Room)

	p.WinTotalCount = 0
	p.PotWinList = nil
	p.CardTypeList = nil
	p.RedBlackList = nil
	p.TwentyData = nil
	p.RedWinCount = 0
	p.BlackWinCount = 0
	p.LuckWinCount = 0
	p.IsOnline = true

	p.LockMoney = 0

	p.HallRoomData = nil

	p.IsRobot = false
	p.NotOnline = 0

	p.LockChan = make(chan bool)
}

//CreatPlayer 创建用户信息
func CreatPlayer() *Player {
	p := &Player{}
	p.Init()

	log.Debug("CreatePlayer index ~ : %v", p.Index)
	return p
}


//SendMsg 发送消息客户端
func (p *Player) SendMsg(msg interface{}) {
	if !p.IsRobot && p.ConnAgent != nil {
		p.ConnAgent.WriteMsg(msg)
	}
}

//onClientBreathe 客户端呼吸，长时间未执行该函数可能已经断网，将主动踢掉
func (p *Player) onClientBreathe() {
	p.uClientDelay = 0
}

//这里是直接设置断线状态，每局结束会断定玩家是否在线，不是则踢掉。
//否则会出现玩家刷新页面生成新的go程，但是玩家还是在线，会导致直接将玩家的当前链接断开
//StartBreathe 开始呼吸。
func (p *Player) StartBreathe() {
	ticker := time.NewTicker(time.Second * 3)
	go func() {
		for { //循环
			<-ticker.C
			p.uClientDelay++
			//已经超过9秒没有收到客户端心跳，踢掉好了
			//log.Debug("p.id:%v ,p.uClientDelay++:%v ", p.Id, p.uClientDelay)

			if p.uClientDelay > 3 {
				p.IsOnline = false

				errMsg := &pb_msg.MsgInfo_S2C{}
				errMsg.Msg = recodeText[RECODE_BREATHSTOP]
				p.SendMsg(errMsg)

				log.Debug("用户长时间未响应心跳,停止心跳~: %v", p.Id)
				return
			}
		}
	}()
}
