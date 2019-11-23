package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)
}

func rpcNewAgent(args []interface{}) {
	log.Debug("---------------新链接请求连接-----------------")
	a := args[0].(gate.Agent)
	p := CreatPlayer()

	//将用户信息塞到链接上
	p.ConnAgent = a
	p.ConnAgent.SetUserData(p)

	//开始呼吸
	p.StartBreathe()
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	//断开链接，删除用户信息，将用户链接设为空
	p, ok := a.UserData().(*Player)
	if ok {
		log.Debug("Player Close Websocket address ~ : %v ", p.Id)

		p.IsOnline = false

		errMsg := &pb_msg.MsgInfo_S2C{}
		errMsg.Msg = recodeText[RECODE_PLAYERBREAKLINE]
		p.SendMsg(errMsg)

		if p.IsInRoom == false && p.IsAction == false {
			c4c.UserLogoutCenter(p.Id, p.PassWord, p.Token) //, p.PassWord
		}
		//log.Debug("玩家断开服务器连接,关闭链接~")
		DeletePlayer(p)
	}
	a.SetUserData(nil)
	a.Close()
	a.Destroy()
}
