package internal

import (
	"github.com/name5566/leaf/module"
	"RedBlack-War/base"
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

	initMongoDB()

	gameHall.Init()
	InitMapPlayer()

	//机器人初始化并开始
	gRobotCenter.Init()
	gRobotCenter.Start()

	//中心服初始化,主动请求Token
	c4c.Init()
	c4c.CreatConnect()
	//c4c.ReqCenterToken()
}

func (m *Module) OnDestroy() {
	c4c.onDestroy()
}
