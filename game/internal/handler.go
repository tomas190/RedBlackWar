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
			log.Debug("进来了0")
			log.Debug("同一用户相同连接重复登录~")
			return
		} else { // 用户相同，链接不相同
			log.Debug("进来了1")
			// 用户处理
			if p.room != nil {
				for i, userId := range p.room.UserLeave {
					log.Debug("AllocateUser 长度~:%v", len(p.room.UserLeave))
					// 把玩家从掉线列表中移除
					if userId == p.Id {
						p.room.UserLeave = append(p.room.UserLeave[:i], p.room.UserLeave[i+1:]...)
						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
						break
					}
					log.Debug("AllocateUser 长度~:%v", len(p.room.UserLeave))
				}
			}

			rId := gameHall.UserRoom[p.Id]
			user, _ := gameHall.UserRecord.Load(p.Id)
			if user != nil {
				log.Debug("进来了4")
				u := user.(*Player)
				login := &pb_msg.LoginInfo_S2C{}
				login.PlayerInfo = new(pb_msg.PlayerInfo)
				login.PlayerInfo.Id = u.Id
				login.PlayerInfo.NickName = u.NickName
				login.PlayerInfo.HeadImg = u.HeadImg
				login.PlayerInfo.Account = u.Account
				a.WriteMsg(login)

				p.ConnAgent.Destroy()
				p.ConnAgent = a
				p.ConnAgent.SetUserData(p)
				p.IsOnline = true

				// 返回游戏大厅数据
				RspGameHallData(u)
			}

			room, _ := gameHall.RoomRecord.Load(rId)
			if room != nil {
				// 玩家如果已在游戏中，则返回房间数据
				r := room.(*Room)

				for i, userId := range r.UserLeave {
					log.Debug("AllocateUser 长度~:%v", len(r.UserLeave))
					// 把玩家从掉线列表中移除
					if userId == p.Id {
						r.UserLeave = append(r.UserLeave[:i], r.UserLeave[i+1:]...)
						log.Debug("AllocateUser 清除玩家记录~:%v", userId)
						break
					}
					log.Debug("AllocateUser 长度~:%v", len(r.UserLeave))
				}

				p.room.GetGodGableId()
				//更新房间列表
				p.room.UpdatePlayerList()
				maintainList := p.room.PackageRoomPlayerList()
				p.room.BroadCastExcept(maintainList, p)
			}
		}
	} else if !gameHall.agentExist(a) { // 玩家首次登入
		log.Debug("进来了2")

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

			u.PassWord = m.GetPassWord()
			u.Token = m.GetToken()

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
			c4c.UserLogoutCenter(p.Id, p.PassWord, p.Token) //, p.PassWord
			p.IsOnline = false
			DeletePlayer(p)
			gameHall.UserRecord.Delete(p.Id)
			leaveHall := &pb_msg.PlayerLeaveHall_S2C{}
			a.WriteMsg(leaveHall)
			p.ConnAgent.Close()
		} else {
			var exist bool
			for _, v := range p.room.UserLeave {
				if v == p.Id {
					exist = true
				}
			}
			if exist == false {
				p.room.UserLeave = append(p.room.UserLeave, p.Id)
			}
			p.IsOnline = false
			leaveHall := &pb_msg.PlayerLeaveHall_S2C{}
			a.WriteMsg(leaveHall)
		}

	}
}
