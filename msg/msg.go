package msg

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network/protobuf"
	"RedBlack-War/msg/Protocal"
)

// 使用默认的 Json 消息处理器 (默认还提供了 protobuf 消息处理器)
var Processor = protobuf.NewProcessor()

func init() {
	log.Debug("msg init~")
	// 注册 UserLogin 协议
	//Processor.Register(&pb_msg.Test{})
	Processor.Register(&pb_msg.Ping{})                //--0
	Processor.Register(&pb_msg.Pong{})                //--1
	Processor.Register(&pb_msg.MsgInfo_S2C{})         //--2
	Processor.Register(&pb_msg.LoginInfo_C2S{})       //--3
	Processor.Register(&pb_msg.LoginInfo_S2C{})       //--4
	Processor.Register(&pb_msg.JoinRoom_C2S{})        //--5
	Processor.Register(&pb_msg.JoinRoom_S2C{})        //--6
	Processor.Register(&pb_msg.LeaveRoom_C2S{})       //--7
	Processor.Register(&pb_msg.LeaveRoom_S2C{})       //--8
	Processor.Register(&pb_msg.EnterRoom_S2C{})       //--9
	Processor.Register(&pb_msg.DownBetTime_S2C{})     //--10
	Processor.Register(&pb_msg.SettlerTime_S2C{})     //--11
	Processor.Register(&pb_msg.PlayerAction_C2S{})    //--12
	Processor.Register(&pb_msg.PlayerAction_S2C{})    //--13
	Processor.Register(&pb_msg.PotTotalMoney_S2C{})   //--14
	Processor.Register(&pb_msg.MaintainList_S2C{})    //--15
	Processor.Register(&pb_msg.OpenCardResult_S2C{})  //--16
	Processor.Register(&pb_msg.RoomSettleData_S2C{})  //--17
	Processor.Register(&pb_msg.GameHallTime_S2C{})    //--18
	Processor.Register(&pb_msg.GameHallData_S2C{})    //--19
	Processor.Register(&pb_msg.PlayerPoolMoney_S2C{}) //--20
	Processor.Register(&pb_msg.PlayerLeaveHall_C2S{}) //--21
	Processor.Register(&pb_msg.PlayerLeaveHall_S2C{}) //--22
}
