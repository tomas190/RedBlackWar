package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
	"github.com/name5566/leaf/gate"
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
	p.RedWinCount = 0
	p.BlackWinCount = 0
	p.LuckWinCount = 0
	p.IsOnline = true

	p.HallRoomData = nil

	p.IsRobot = false
	p.NotOnline = 0
}

// 用户缓存数据
var mapPlayerIndex uint32
var mapGlobalPlayer map[uint32]*Player
var mapUserIDPlayer map[string]*Player

// 初始化全局用户列表
func InitMapPlayer() {
	mapPlayerIndex = 0
	mapGlobalPlayer = make(map[uint32]*Player)
	mapUserIDPlayer = make(map[string]*Player)
}

//CreatPlayer 创建用户信息
func CreatPlayer() *Player {
	p := &Player{}
	p.Init()
	mapGlobalPlayer[mapPlayerIndex] = p

	p.Index = mapPlayerIndex
	log.Debug("CreatePlayer index ~ : %v", p.Index)
	mapPlayerIndex++
	return p
}

//RegisterPlayer 注册用户信息
func RegisterPlayer(p *Player) {
	log.Debug("RegisterPlayer ~ : %v", p.Id)
	// 获取用户当前是否已经存在
	up, ok := mapUserIDPlayer[p.Id]

	// 如果有相同的ID，则断开和删除当前的用户链接，让新用户登录
	if ok {
		log.Debug("Have the same Player ID Login :%v", up.Id)

		errMsg := &pb_msg.MsgInfo_S2C{}
		errMsg.Msg = recodeText[RECODE_PLAYERDESTORY]
		p.SendMsg(errMsg)
		log.Debug("用户已在其他地方登录~")

		//up.ConnAgent.Close()
		DeletePlayer(up)
	}
	//将链接的Player数据赋值给map缓存
	mapUserIDPlayer[p.Id] = p
}

//DeletePlayer 删除用户信息
func DeletePlayer(p *Player) {
	// 删除mapGlobalPlayer用户索引
	delete(mapGlobalPlayer, p.Index)

	up, ok := mapUserIDPlayer[p.Id]
	if ok && up.Index == p.Index {
		// 删除mapUserIDPlayer用户索引
		delete(mapUserIDPlayer, p.Id)
		log.Debug("DeletePlayer SUCCESS ~ : %v", p.Id)
	} else {
		log.Debug("DeletePlayer come to nothing ~ : %v", p.Id)
	}
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

//SyncScoreChangeToCenter 同步金币变化到中心服
func (p *Player) SyncScoreChangeToCenter(reason string) {
	//跨模块调用到login，然后由login调用到Center
	//login.ChanRPC.Go("SyncCenterScoreChange", p.ID, p.fWinScore, p.fLoseScore)
	c4c.UserSyncScoreChange(p, reason)
}

func (p *Player) PlayerLoginHandle(userId string, a gate.Agent) {
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


	p.ConnAgent = a
	p.ConnAgent.SetUserData(p)
	p.IsOnline = true

	v, _ := gameHall.UserRecord.Load(userId)
	u := v.(*Player)

	login := &pb_msg.LoginInfo_S2C{}
	login.PlayerInfo = new(pb_msg.PlayerInfo)
	login.PlayerInfo.Id = u.Id
	login.PlayerInfo.NickName = u.NickName
	login.PlayerInfo.HeadImg = u.HeadImg
	login.PlayerInfo.Account = u.Account
	a.WriteMsg(login)

	// 返回游戏大厅数据
	RspGameHallData(u)

	rId := gameHall.UserRoom[p.Id]
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
