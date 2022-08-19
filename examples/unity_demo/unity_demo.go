package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/examples/unity_demo/properties/eType"
)

var (
	_SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
	}
)

func main() {
	defer func() {
		err := recover()
		if err != nil {
			gwlog.Errorf("system crash", err)
		}
	}()

	goworld.RegisterSpace(&MySpace{}) // 注册自定义的Space类型

	goworld.RegisterService("OnlineService", &OnlineService{}, 3)
	goworld.RegisterService("SpaceService", &SpaceService{}, 3)

	// account
	goworld.RegisterEntity(eType.Account, &Account{})
	// player
	goworld.RegisterEntity(eType.Player, &Player{})
	// monster
	goworld.RegisterEntity(eType.Monster, &Monster{})
	// black hole
	goworld.RegisterEntity(eType.BlackHole, &BlackHole{})

	// 运行游戏服务器
	goworld.Run()

}
