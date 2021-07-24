package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/log"
	"time"
)

//JoinGameRoom 加入游戏房间
func (r *Room) JoinGameRoom(p *Player) {
	//寻找可用的座位号
	//p.SeatNum = r.FindUsableSeat()
	//r.PlayerList[p.SeatNum] = p

	if p.IsRobot == false {
		FindPlayerID(p)
	}
	//将用户添加到用户列表
	r.PlayerList = append(r.PlayerList, p)

	p.room = r

	p.GameState = InGameRoom

	//进入房间玩家是否大于 50金币，否则处于观战状态
	p.PlayerMoneyHandler()

	//获取最新40局游戏数据(小于40局则全部显示出来)
	p.GetRoomCordData(r)

	// 看数据用,打印玩家列表信息
	//r.PrintPlayerList()

	if p.IsRobot == false {
		//更新房间列表
		r.UpdatePlayerList()
		maintainList := r.PackageRoomPlayerList()
		r.BroadCastMsg(maintainList)
	}

	//判断房间人数是否小于两人，否则不能开始运行
	if r.PlayerLength() < 2 {
		//房间游戏不能开始,房间设为等待状态
		r.RoomStat = RoomStatusNone

		msgInfo := &pb_msg.MsgInfo_S2C{}
		msgInfo.Msg = recodeText[RECODE_PEOPLENOTFULL]
		p.SendMsg(msgInfo)
		log.Debug("房间当前人数不足，无法开始游戏 ~")

		//返回前端房间信息
		msg := &pb_msg.JoinRoom_S2C{}
		roomData := p.room.RspRoomData()
		msg.RoomData = roomData
		p.SendMsg(msg)
		//log.Debug("返回客户端房间信息 JoinRoom_S2C ~")

		return
	}

	//只要不小于两人,就属于游戏状态
	p.Status = PlayGame

	//返回前端房间信息
	msg := &pb_msg.JoinRoom_S2C{}
	roomData := p.room.RspRoomData()
	msg.RoomData = roomData
	if r.GameStat == DownBet {
		msg.GameTime = DownBetTime - r.counter
		//log.Debug("加入房间 DownBetTime.GameTime: %v", msg.GameTime)
	} else {
		msg.GameTime = SettleTime - r.counter
		//log.Debug("加入房间 SettleTime GameTime: %v", msg.GameTime)
	}
	p.SendMsg(msg)
	//log.Debug("返回客户端房间信息 JoinRoom_S2C ~")

	if r.RoomStat != RoomStatusRun {
		// None和Over状态都直接开始运行游戏
		r.StartGameRun()
	} else {
		if r.GameStat == Settle { //这里给前端发送消息 做处理
			msg := &pb_msg.MsgInfo_S2C{}
			msg.Msg = recodeText[RECODE_SELLTENOTDOWNBET]
			p.SendMsg(msg)

			//log.Debug("当前结算阶段, 不能进行操作 ~")
		}
	}
}

//ExitFromRoom 从房间退出处理
func (r *Room) ExitFromRoom(p *Player) {

	//清空用户数据
	p.GameState = InGameHall
	p.DownBetMoneys = new(DownBetMoney)
	p.TotalAmountBet = 0
	p.IsAction = false
	p.ContinueVot = new(ContinueBet)
	p.ContinueVot.DownBetMoneys = new(DownBetMoney)
	p.WinTotalCount = 0
	p.PotWinList = nil
	p.CardTypeList = nil
	p.RedBlackList = nil
	p.HallRoomData = nil
	p.RedWinCount = 0
	p.BlackWinCount = 0
	p.LuckWinCount = 0
	p.NotOnline = 0
	p.TwentyData = nil

	//从房间列表删除玩家信息,更新房间列表
	for k, v := range r.PlayerList {
		if v != nil && v.Id == p.Id {
			if v.IsRobot == false {
				p.room = nil
				//userRoomMap = make(map[string]*Room)
				log.Debug("p.id:%v k:%v", p.Id, k)
				r.PlayerList = append(r.PlayerList[:k], r.PlayerList[k+1:]...) //这里两个同样的用户名退出，会报错
				log.Debug("%v 玩家从房间列表删除成功 ~", v.Id)
			} else {
				p.room = nil
				r.PlayerList = append(r.PlayerList[:k], r.PlayerList[k+1:]...)
				//log.Debug("%v 机器从房间列表删除成功 ~", v.Id)
				//创建机器人 todo
				//robot := gRobotCenter.CreateRobot()
				//r.JoinGameRoom(robot)
			}
		}
	}

	//更新房间赌神ID
	r.GetGodGableId()
	//更新房间列表
	r.UpdatePlayerList()
	maintainList := r.PackageRoomPlayerList()
	r.BroadCastExcept(maintainList, p)

	//广播其他玩家该玩家退出房间
	leave := &pb_msg.LeaveRoom_S2C{}
	leave.PlayerInfo = new(pb_msg.PlayerInfo)
	leave.PlayerInfo.Id = p.Id
	leave.PlayerInfo.NickName = p.NickName
	leave.PlayerInfo.HeadImg = p.HeadImg
	leave.PlayerInfo.Account = p.Account
	p.SendMsg(leave)

	//更新大厅时间和数据
	RspGameHallData(p)

	//log.Debug("Player Exit from the Room SUCCESS ~")
}

//ExitFromRoom 从房间退出处理
func (r *Room) RobotExitFromRoom(p *Player) {

	//清空用户数据
	p.GameState = InGameHall
	p.DownBetMoneys = new(DownBetMoney)
	p.TotalAmountBet = 0
	p.IsAction = false
	p.ContinueVot = new(ContinueBet)
	p.ContinueVot.DownBetMoneys = new(DownBetMoney)
	p.WinTotalCount = 0
	p.PotWinList = nil
	p.CardTypeList = nil
	p.RedBlackList = nil
	p.HallRoomData = nil
	p.RedWinCount = 0
	p.BlackWinCount = 0
	p.LuckWinCount = 0
	p.NotOnline = 0
	p.TwentyData = nil

	//从房间列表删除玩家信息,更新房间列表
	for k, v := range r.PlayerList {
		if v != nil && v.Id == p.Id {
			if v.IsRobot == false {
				p.room = nil
				//userRoomMap = make(map[string]*Room)
				log.Debug("p.id:%v k:%v", p.Id, k)
				r.PlayerList = append(r.PlayerList[:k], r.PlayerList[k+1:]...) //这里两个同样的用户名退出，会报错
				log.Debug("%v 玩家从房间列表删除成功 ~", v.Id)
			} else {
				p.room = nil
				r.PlayerList = append(r.PlayerList[:k], r.PlayerList[k+1:]...)
				//log.Debug("%v 机器从房间列表删除成功 ~", v.Id)
				//创建机器人 todo
				//robot := gRobotCenter.CreateRobot()
				//r.JoinGameRoom(robot)
			}
		}
	}

	//更新大厅时间和数据
	RspGameHallData(p)

	//log.Debug("Player Exit from the Room SUCCESS ~")
}

//LoadRoomRobots 装载机器人
func (r *Room) LoadRoomRobots(num int) {
	log.Debug("房间: %v ----- 装载 %v个机器人", r.RoomId, num)
	r.IsLoadRobots = true
	for i := 0; i < num; i++ {
		time.Sleep(time.Millisecond)
		robot := gRobotCenter.CreateRobot()
		r.JoinGameRoom(robot)
	}
}
