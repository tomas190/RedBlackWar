package internal

import (
	"fmt"
	"math/rand"
	"time"
)

func (r *Room) RoomInit() {

	//r.RoomId = r.GetRoomNumber()
	//r.RoomId = "1"
	r.PlayerList = nil

	r.GodGambleName = ""
	r.RoomStat = RoomStatusNone

	r.Cards = new(CardData)
	r.PotMoneyCount = new(PotRoomCount)
	r.CardTypeList = nil
	r.RPotWinList = nil
	r.GameTotalCount = 0

	DownBetChannel = make(chan bool)
	RobotDownBetChan = make(chan bool)

	r.counter = 0
	r.clock = time.NewTicker(time.Second)

	r.IsLoadRobots = false
	r.UserLeave = make([]string, 0)
}

func (r *Room) GetRoomNumber() string {
	roomNumber := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	return roomNumber
}
