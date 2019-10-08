package internal

import (
	"github.com/name5566/leaf/log"
	pb_msg "server/msg/Protocal"
	"strconv"
	"time"
)

func (gh *GameHall) Init() {
	gh.maxPlayerInHall = 5000
	log.Debug("GameHall Init~!!! This gameHall can hold %d player running ~", gh.maxPlayerInHall)
	//r := gh.CreatGameRoom()
	//gh.roomList[0] = r
	//log.Debug("大厅房间数量: %d, 房间号: %v", len(gh.roomList), gh.roomList[0].RoomId)

	for i := 0; i < 6; i++ {
		time.Sleep(time.Millisecond)
		r := gh.CreatGameRoom()
		ri := i + 1
		r.RoomId = strconv.Itoa(ri)
		gh.roomList[i] = r
		log.Debug("大厅房间数量: %d,房间号: %v", i, gh.roomList[i].RoomId)
	}
}

//CreatGameRoom 创建游戏房间
func (gh *GameHall) CreatGameRoom() *Room {
	r := &Room{}
	r.RoomInit()
	return r
}

//PlayerJoinRoom 玩家大厅加入房间
func (gh *GameHall) PlayerJoinRoom(rid string, p *Player) {
	p.IsOnline = true
	for _, room := range gameHall.roomList {
		if room != nil {
			for _, v := range room.PlayerList {
				if v != nil && v.IsRobot == false && v.Id == p.Id {
					msg := &pb_msg.JoinRoom_S2C{}
					roomData := p.room.RspRoomData()
					msg.RoomData = roomData
					if room != nil && room.RoomId == rid {
						if room.GameStat == DownBet {
							msg.GameTime = DownBetTime - room.counter
							log.Debug("加入房间 DownBetTime.GameTime: %v", msg.GameTime)
						} else {
							msg.GameTime = SettleTime - room.counter
							log.Debug("加入房间 SettleTime GameTime: %v", msg.GameTime)
						}
					}
					p.SendMsg(msg)
					log.Debug("进来了呀~~~~~~~~~")

					//玩家各注池下注金额
					pool := &pb_msg.PlayerPoolMoney_S2C{}
					pool.DownBetMoney = new(pb_msg.DownBetMoney)
					pool.DownBetMoney.RedDownBet = p.DownBetMoneys.RedDownBet
					pool.DownBetMoney.BlackDownBet = p.DownBetMoneys.BlackDownBet
					pool.DownBetMoney.LuckDownBet = p.DownBetMoneys.LuckDownBet
					p.SendMsg(pool)

					//更新列表
					p.room.UpdatePlayerList()
					maintainList := p.room.PackageRoomPlayerList()
					p.room.BroadCastMsg(maintainList)

					return
				}
			}
		}
	}

	for _, room := range gh.roomList {
		if room != nil && room.RoomId == rid { // 这里要不要遍历房间，查看用户id是否存在
			//for _, v := range room.PlayerList {
			//	if v != nil && v.Id == p.Id {
			//		msg := &pb_msg.MsgInfo_S2C{}
			//		msg.Msg = recodeText[RECODE_PLAYERHAVESAME]
			//		v.ConnAgent.WriteMsg(msg)
			//		log.Debug("当前房间已存在相同的用户ID~")
			//		return
			//	}
			//}
			//加入房间
			room.JoinGameRoom(p)
			return
		}
	}
	msg := &pb_msg.MsgInfo_S2C{}
	msg.Error = recodeText[RECODE_JOINROOMIDERR]
	p.SendMsg(msg)

	log.Debug("请求加入的房间号不正确~")
}

//LoadHallRobots 为每个房间装载机器人
func (gh *GameHall) LoadHallRobots(num int) {
	for _, room := range gh.roomList {
		if room != nil {
			room.LoadRoomRobots(num)
		}
	}
}
