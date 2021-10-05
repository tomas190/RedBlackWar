package internal

import (
	"RedBlack-War/base"
	"github.com/name5566/leaf/module"
)

var (
	skeleton = base.NewSkeleton()
	ChanRPC  = skeleton.ChanRPCServer

	c4c = &Conn4Center{}
)

type Module struct {
	*module.Skeleton
}

func (m *Module) OnInit() {
	m.Skeleton = skeleton

	//packageTax = make(map[uint16]uint8)

	initMongoDB()

	gameHall.Init()

	//机器人初始化并开始
	gRobotCenter.Init()
	gRobotCenter.Start()

	//中心服初始化,主动请求Token
	c4c.Init()
	c4c.CreatConnect()
	//c4c.ReqCenterToken()

	//中心服日志初始化
	cc.Init()

	go StartHttpServer()

}

func (m *Module) OnDestroy() {
	gameHall.UserRecord.Range(func(key, value interface{}) bool {
		p := value.(*Player)
		if p.LockMoney > 0 {
			c4c.UnlockSettlement(p.Id, p.LockMoney)
		}
		c4c.UserLogoutCenter(p.Id, p.PassWord, p.Token)
		return true
	})
}
