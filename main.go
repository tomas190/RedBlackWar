package main

import (
	"github.com/name5566/leaf"
	lconf "github.com/name5566/leaf/conf"
	"RedBlack-War/conf"
	"RedBlack-War/game"
	"RedBlack-War/gate"
)

func main() {
	// 加载配置
	lconf.LogLevel = conf.Server.LogLevel
	lconf.LogPath = conf.Server.LogPath
	lconf.LogFlag = conf.LogFlag
	lconf.ConsolePort = conf.Server.ConsolePort
	lconf.ProfilePath = conf.Server.ProfilePath

	// 注册模块
	leaf.Run(
		game.Module,
		gate.Module,
		//login.Module,  //login模块直接在game模块里面处理
	)
}
