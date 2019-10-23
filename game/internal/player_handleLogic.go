package internal

import (
	pb_msg "RedBlack-War/msg/Protocal"
)

//PlayerMoneyHandler 玩家进入房间金额处理
func (p *Player) PlayerMoneyHandler() {
	if p.Account < RoomLimitMoney {
		//玩家观战状态不能进行投注
		p.Status = WatchGame

		errMsg := &pb_msg.MsgInfo_S2C{}
		errMsg.Msg = recodeText[RECODE_MONEYNOTFULL]
		p.SendMsg(errMsg)

		//log.Debug("玩家金额不足,设为观战~")
		return
	}
}

//GetPotWinCount 获取玩家在房间内的总局数
func (p *Player) GetPotWinCount() int32 {
	return int32(len(p.PotWinList))
}

//GetPlayerTableData 获取房间战绩数据
func (p *Player) GetRoomCordData(r *Room) {

	//最新40局游戏数据、红黑Win顺序列表、每局Win牌局类型、红黑Luck的总数量
	roomGCount := r.RoomGameCount()
	//判断房间数据是否大于40局
	if roomGCount > RoomCordCount {
		//大于40局则截取最新40局数据
		num := roomGCount - RoomCordCount
		//log.Debug("num 长度: %v", num)
		//log.Debug("r.RPotWinList 长度: %v", len(r.RPotWinList))
		p.PotWinList = append(p.PotWinList, r.RPotWinList[num:]...)
		p.CardTypeList = append(p.CardTypeList, r.CardTypeList[num:]...)
		for _, v := range p.PotWinList {
			if v.RedWin == 1 {
				p.RedWinCount++
				p.RedBlackList = append(p.RedBlackList, RedWin)
			}
			if v.BlackWin == 1 {
				p.BlackWinCount++
				p.RedBlackList = append(p.RedBlackList, BlackWin)
			}
			if v.LuckWin == 1 {
				p.LuckWinCount++
			}
		}
		p.TotalCount = RoomCordCount
	} else {
		//小于40局则截取全部房间数据
		p.PotWinList = append(p.PotWinList, r.RPotWinList...)
		p.CardTypeList = append(p.CardTypeList, r.CardTypeList...)
		for _, v := range p.PotWinList {
			if v.RedWin == 1 {
				p.RedWinCount++
				p.RedBlackList = append(p.RedBlackList, RedWin)

			}
			if v.BlackWin == 1 {
				p.BlackWinCount++
				p.RedBlackList = append(p.RedBlackList, BlackWin)
			}
			if v.LuckWin == 1 {
				p.LuckWinCount++
			}
		}
		p.TotalCount = roomGCount
	}
}

//RspRoomData 返回房间信息
func (r *Room) RspRoomData() *pb_msg.RoomData {
	room := &pb_msg.RoomData{}
	room.RoomId = r.RoomId

	for _, v := range r.PlayerList {
		if v != nil {
			data := &pb_msg.PlayerData{}
			data.PlayerInfo = new(pb_msg.PlayerInfo)
			data.PlayerInfo.Id = v.Id
			data.PlayerInfo.NickName = v.NickName
			data.PlayerInfo.HeadImg = v.HeadImg
			data.PlayerInfo.Account = v.Account
			data.DownBetMoneys = new(pb_msg.DownBetMoney)
			data.DownBetMoneys.RedDownBet = v.DownBetMoneys.RedDownBet
			data.DownBetMoneys.BlackDownBet = v.DownBetMoneys.BlackDownBet
			data.DownBetMoneys.LuckDownBet = v.DownBetMoneys.LuckDownBet
			data.TotalAmountBet = v.TotalAmountBet
			data.Status = pb_msg.PlayerStatus(v.Status)
			data.ContinueVot = new(pb_msg.ContinueBet)
			data.ContinueVot.DownBetMoneys = new(pb_msg.DownBetMoney)
			data.ContinueVot.DownBetMoneys.RedDownBet = v.ContinueVot.DownBetMoneys.RedDownBet
			data.ContinueVot.DownBetMoneys.BlackDownBet = v.ContinueVot.DownBetMoneys.BlackDownBet
			data.ContinueVot.DownBetMoneys.LuckDownBet = v.ContinueVot.DownBetMoneys.LuckDownBet
			data.ContinueVot.TotalMoneyBet = v.ContinueVot.TotalMoneyBet
			data.ResultMoney = v.ResultMoney
			data.WinTotalCount = v.WinTotalCount
			data.CardTypeList = v.CardTypeList
			for _, val := range v.PotWinList {
				pot := &pb_msg.PotWinList{}
				pot.RedWin = val.RedWin
				pot.BlackWin = val.BlackWin
				pot.LuckWin = val.LuckWin
				pot.CardType = pb_msg.CardsType(val.CardTypes)
				data.PotWinList = append(data.PotWinList, pot)
			}
			data.RedBlackList = v.RedBlackList
			data.RedWinCount = v.RedWinCount
			data.BlackWinCount = v.BlackWinCount
			data.LuckWinCount = v.LuckWinCount
			data.TotalCount = v.TotalCount
			data.IsOnline = v.IsOnline
			room.PlayerList = append(room.PlayerList, data)
		}
	}
	room.GodGableName = r.GodGambleName
	room.GameStage = pb_msg.GameStage(r.GameStat)
	room.Cards = new(pb_msg.CardData)
	if r.Cards != nil {
		room.Cards.ReadCard = r.Cards.ReadCard
		room.Cards.BlackCard = r.Cards.BlackCard
		room.Cards.RedType = pb_msg.CardsType(r.Cards.RedType)
		room.Cards.BlackType = pb_msg.CardsType(r.Cards.BlackType)
		room.Cards.LuckType = pb_msg.CardsType(r.Cards.LuckType)
	}
	room.PotMoneyCount = new(pb_msg.PotMoneyCount)
	if r.PotMoneyCount != nil {
		room.PotMoneyCount.RedMoneyCount = r.PotMoneyCount.RedMoneyCount
		room.PotMoneyCount.BlackMoneyCount = r.PotMoneyCount.BlackMoneyCount
		room.PotMoneyCount.LuckMoneyCount = r.PotMoneyCount.LuckMoneyCount
	}
		room.CardTypeList = r.CardTypeList
	for _, value := range r.RPotWinList {
		pot := &pb_msg.PotWinList{}
		pot.RedWin = value.RedWin
		pot.BlackWin = value.BlackWin
		pot.LuckWin = value.LuckWin
		pot.CardType = pb_msg.CardsType(value.CardTypes)
		room.RPotWinList = append(room.RPotWinList, pot)
	}
	return room
}

//PlayerActionDownBet 玩家行动下注
func (p *Player) ActionHandler() {
	//判断玩家是否行动,做相应处理
}
