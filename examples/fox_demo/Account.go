package main

import (
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// 玩家类型
type Account struct {
	// 自定义对象类型必须继承entity.Entity
	entity.Entity
	logIn bool
}

// OnCreated 在Player对象创建后被调用
func (a *Account) OnCreated() {
	gwlog.Debugf("account created", a)
}

func (a *Account) DescribeEntityType(desc *entity.EntityTypeDesc) {
}

// OnClientDisconnected 在客户端掉线或者给了Player后触发
func (a *Account) OnClientDisconnected() {
	gwlog.Debugf("destroying %s ...", a)
	a.Destroy()
}

// Register_Client 是处理玩家注册请求的RPC函数
func (a *Account) Register_Client(username string, password string) {
	gwlog.Debugf("Register %s %s", username, password)
	goworld.GetOrPutKVDB("password$"+username, password, func(oldVal string, err error) {
		if err != nil {
			a.CallClient("ShowInfo", "Server Error： "+err.Error()) // 服务器错误
			return
		}

		if oldVal == "" {
			player := goworld.CreateEntityLocally("Player") // 创建一个Player对象然后立刻销毁，产生一次存盘
			player.Attrs.SetStr("name", username)
			player.Destroy()
			goworld.PutKVDB("playerID$"+username, string(player.ID), func(err error) {
				a.CallClient("ShowInfo", "Registered Successfully, please click login.") // 注册成功，请点击登录
			})
		} else {
			a.CallClient("ShowInfo", "Sorry, this account aready exists.") // 抱歉，这个账号已经存在
		}
	})
}

// Login_Client 是处理玩家登录请求的RPC函数
func (a *Account) Login_Client(username string, password string) {
	gwlog.Debugf("%s.Login: username=%s, password=%s", a, username, password)
	if a.logIn {
		// logining
		gwlog.Errorf("%s has already started to log in.", a)
		return
	}

	gwlog.Infof("%s started log in with username %s password %s ...", a, username, password)
	a.logIn = true
	goworld.GetKVDB("password$"+username, func(correctPassword string, err error) {
		if err != nil {
			a.logIn = false
			a.CallClient("ShowInfo", "Server Error： "+err.Error()) // 服务器错误
			return
		}

		if correctPassword == "" {
			a.logIn = false
			a.CallClient("ShowInfo", "Account does not exist.") // 账号不存在
			return
		}

		if password != correctPassword {
			a.logIn = false
			a.CallClient("ShowInfo", "Invalid password or username") // 密码错误
			return
		}

		goworld.GetKVDB("playerID$"+username, func(_playerID string, err error) {
			if err != nil {
				a.logIn = false
				a.CallClient("ShowInfo", "Server Error："+err.Error()) // 服务器错误
				return
			}
			playerID := common.EntityID(_playerID)
			goworld.LoadEntityLocally("Player", playerID)
			a.Call(playerID, "AccountCall", a.ID)
		})
	})
}

func (a *Account) OnAccountCall(callerID common.EntityID) {
	player := goworld.GetEntity(callerID)
	a.logIn = false
	a.GiveClientTo(player)
	goworld.CallServiceShardIndex("SpaceService", 0, "GetSpaceID", player.ID)
	// goworld.CallServiceShardIndex("SpaceService", 0, "GetSpaceID", player.ID)
	// goworld.CallServiceAny("SpaceService", "GetSpaceID", player.ID)
}
