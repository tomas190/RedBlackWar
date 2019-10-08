package gate

import (
	"server/game"
	"server/msg"
	"server/msg/Protocal"
)

func init() {
	// 指定消息 Hello 路由到 game 模块
	//msg.Processor.SetRouter(&pb_msg.Test{},game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.Ping{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.LoginInfo_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.JoinRoom_C2S{},  game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.LeaveRoom_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.PlayerAction_C2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.PlayerLeaveHall_C2S{}, game.ChanRPC)
}
