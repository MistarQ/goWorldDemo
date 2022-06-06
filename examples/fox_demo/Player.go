package main

import (
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Player 对象代表一名玩家
type Player struct {
	entity.Entity
}

// OnCreated 在Player对象创建后被调用
func (a *Player) OnCreated() {
	a.Attrs.SetDefaultInt("coin", 250)
	a.SetClientSyncing(true)
}

func (a *Player) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetPersistent(true)
	desc.SetUseAOI(true, 100)
	desc.DefineAttr("name", "AllClients", "Persistent")
	desc.DefineAttr("coin", "Client", "Persistent")
}

func (a *Player) AccountCall(callerID common.EntityID) {
	a.Call(callerID, "OnAccountCall", a.ID)
}

// OnGetSpaceID is called by SpaceService
func (a *Player) OnGetSpaceID(spaceID common.EntityID) {
	// let account enter space with spaceID
	a.EnterSpace(spaceID, entity.Vector3{})

	gwlog.Debugf("plyaer spaceId is %s", a.Space)
	gwlog.Debugf("my spaceId is %s", spaceID)
}
