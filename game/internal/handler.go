package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
)

func init() {
	//向当前模块（game 模块）注册 Test 消息的消息处理函数 handleTest
	//handler(&pb_msg.Test{},handleTest)
	handler(&pb_msg.Ping{}, handlePing)
	handler(&pb_msg.LoginInfo_C2S{}, handleLoginInfo)
	handler(&pb_msg.JoinRoom_C2S{}, handleJoinRoom)
	handler(&pb_msg.LeaveRoom_C2S{}, handleLeaveRoom)
	handler(&pb_msg.PlayerAction_C2S{}, handlePlayerAction)
	handler(&pb_msg.PlayerLeaveHall_C2S{}, handleLeaveHall)
}

// 异步处理
func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	// 收到的 Hello 消息
	a := args[1].(gate.Agent)

	HeartBeatHandle(a)
}

func handleLoginInfo(args []interface{}) {
	m := args[0].(*pb_msg.LoginInfo_C2S)
	a := args[1].(gate.Agent)

	log.Debug("handleLoginInfo 用户登录成功~ : %v", m)

	userId := m.GetId()
	v, ok := gameHall.UserRecord.Load(userId)
	if ok { // 说明用户已存在
		p := v.(*Player)
		if p.ConnAgent == a { // 用户和链接都相同
			log.Debug("同一用户相同连接重复登录~")
			return
		} else { // 用户相同，链接不相同
			err := gameHall.ReplacePlayerAgent(userId, a)
			if err != nil {
				log.Error("用户链接替换错误", err)
			}

			v, _ := gameHall.UserRecord.Load(userId)
			u := v.(*Player)

			login := &pb_msg.LoginInfo_S2C{}
			login.PlayerInfo = new(pb_msg.PlayerInfo)
			login.PlayerInfo.Id = u.Id
			login.PlayerInfo.NickName = u.NickName
			login.PlayerInfo.HeadImg = u.HeadImg
			login.PlayerInfo.Account = u.Account
			a.WriteMsg(login)

			rId := gameHall.UserRoom[p.Id]
			room, _ := gameHall.RoomRecord.Load(rId)
			if room != nil {
				// 玩家如果已在游戏中，则返回房间数据
				r := room.(*Room)
				enter := &pb_msg.EnterRoom_S2C{}
				enter.RoomData = r.RspRoomData()
				if p.room.GameStat == DownBet {
					enter.GameTime = DownBetTime - p.room.counter
					log.Debug("用户重新登陆 DownBetTime.GameTime: %v", enter.GameTime)
				} else {
					enter.GameTime = SettleTime - p.room.counter
					log.Debug("用户重新登陆 SettleTime.GameTime: %v", enter.GameTime)
				}
				if rID, ok := gameHall.UserRoom[userId]; ok {
					enter.RoomData.RoomId = rID // 如果用户之前在房间里后来退出，返回房间号
				}
				log.Debug("<----login 登录 resp---->%+v %+v", enter.RoomData.RoomId)
				a.WriteMsg(enter)

				p.room.GetGodGableId()
				//更新房间列表
				p.room.UpdatePlayerList()
				maintainList := p.room.PackageRoomPlayerList()
				p.room.BroadCastExcept(maintainList, p)
			}
		}
	} else if !gameHall.agentExist(a) { // 玩家首次登入
		c4c.UserLoginCenter(m.GetId(), m.GetPassWord(), m.GetToken(), func(u *Player) {
			login := &pb_msg.LoginInfo_S2C{}
			login.PlayerInfo = new(pb_msg.PlayerInfo)
			login.PlayerInfo.Id = u.Id
			login.PlayerInfo.NickName = u.NickName
			login.PlayerInfo.HeadImg = u.HeadImg
			login.PlayerInfo.Account = u.Account
			a.WriteMsg(login)

			u.Init()
			// 重新绑定信息
			u.ConnAgent = a
			a.SetUserData(u)

			RegisterPlayer(u)
			gameHall.UserRecord.Store(u.Id, u)

			// 返回游戏大厅数据
			RspGameHallData(u)
		})
	} // 同一连接上不同用户的情况对第二个用户的请求不做处理
}

func handleJoinRoom(args []interface{}) {
	m := args[0].(*pb_msg.JoinRoom_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleJoinRoom 玩家加入房间~ : %v", p.Id)

	if ok {
		gameHall.PlayerJoinRoom(m.RoomId, p)
	}
}

func handleLeaveRoom(args []interface{}) {
	//m := args[0].(*pb_msg.LeaveRoom_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLeaveRoom 玩家退出房间~ : %v", p.Id)

	if ok {
		p.PlayerReqExit()
	}
}

func handlePlayerAction(args []interface{}) {
	m := args[0].(*pb_msg.PlayerAction_C2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handlePlayerAction 玩家开始行动~ : %v", p.Id)

	if ok {
		p.SetPlayerAction(m)
	}
}

func handleLeaveHall(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLeaveHall 玩家退出大厅~ : %v", p.Id)

	if ok {

		if p.IsAction == false {
			DeletePlayer(p)
			gameHall.UserRecord.Delete(p.Id)
			c4c.UserLogoutCenter(p.Id, p.PassWord, p.Token) //, p.PassWord
			p.ConnAgent.Close()
		}

		leaveHall := &pb_msg.PlayerLeaveHall_S2C{}
		p.SendMsg(leaveHall)
	}
}
