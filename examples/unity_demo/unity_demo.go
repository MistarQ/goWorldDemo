package main

import (
	"github.com/xiaonanln/goworld"
)

var (
	_SERVICE_NAMES = []string{
		"OnlineService",
		"SpaceService",
	}
)

func main() {
	goworld.RegisterSpace(&MySpace{}) // 注册自定义的Space类型

	goworld.RegisterService("OnlineService", &OnlineService{}, 3)
	goworld.RegisterService("SpaceService", &SpaceService{}, 3)

	// account
	goworld.RegisterEntity("Account", &Account{})
	// player
	goworld.RegisterEntity("Player", &Player{})
	// monster
	goworld.RegisterEntity("Monster", &Monster{})
	// black hole
	goworld.RegisterEntity("BlackHole", &BlackHole{})

	// 运行游戏服务器
	goworld.Run()
}
