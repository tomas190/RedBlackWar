package internal

// 大厅全局变量
var gameHall GameHall

const (
	RoomNumber = 6
)

//GameHall 描述游戏大厅
type GameHall struct {
	maxPlayerInHall uint32
	roomList        [RoomNumber]*Room
}


