package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/log"
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
	gh.userAndRoom = make(map[string]*Room)
}

//CreatGameRoom 创建游戏房间
func (gh *GameHall) CreatGameRoom() *Room {
	r := &Room{}
	r.RoomInit()
	return r
}

//PlayerJoinRoom 玩家大厅加入房间
func (gh *GameHall) PlayerJoinRoom(rid string, p *Player) {
	for _, r := range gh.roomList {
		for _, v := range r.PlayerList {
			if v != nil && v.Id == p.Id {
				p.room = r
				p.IsOnline = true
				log.Debug("玩家数据 :%v", p.IsOnline)
				msg := &pb_msg.JoinRoom_S2C{}
				roomData := p.room.RspRoomData()
				msg.RoomData = roomData
				if p.room.GameStat == DownBet {
					msg.GameTime = DownBetTime - p.room.counter
				} else {
					msg.GameTime = SettleTime - p.room.counter
				}
				p.SendMsg(msg)

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

	for _, room := range gh.roomList {
		if room != nil && room.RoomId == rid { // 这里要不要遍历房间，查看用户id是否存在
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
