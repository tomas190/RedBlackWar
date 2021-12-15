package internal

import "sync"

// 大厅全局变量
var gameHall GameHall

const (
	RoomNumber = 6
)

//GameHall 描述游戏大厅
type GameHall struct {
	roomList        [RoomNumber]*Room
	UserRecord      sync.Map // 用户记录
	RoomRecord      sync.Map // 房间记录
	UserRoom        sync.Map // 用户房间

	OrderIDRecord sync.Map // orderID对应user
}
