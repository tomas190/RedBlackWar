package internal

import (
	"fmt"
	"math/rand"
	"time"
)

var packageTax map[uint16]float64

func (r *Room) RoomInit() {

	r.PlayerList = nil

	r.PackageId = 0
	r.GodGambleName = ""
	r.RoomStat = RoomStatusNone

	r.Cards = new(CardData)
	r.PotMoneyCount = new(PotRoomCount)
	r.CardTypeList = nil
	r.RPotWinList = nil
	r.GameTotalCount = 0

	r.counter = 0
	r.clock = time.NewTicker(time.Second)

	r.IsLoadRobots = false
	r.UserLeave = make([]string, 0)
	r.IsSpecial = false
	r.IsContinue = true

	r.DownBetChannel = make(chan bool)
}

func (r *Room) GetRoomNumber() string {
	roomNumber := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	return roomNumber
}
